package binary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/storage/data"
	"gophkeeper/internal/client/storage/data/tools/binary"
	repoData "gophkeeper/internal/repositories/data"
	"os"
	"time"

	"github.com/rivo/tview"
)

// DataInfo - вспомогательная структура для передачи полученных от пользоавтеля данных в функцию сохранения данных в сервисе.
type DataInfo struct {
	Binary     data.Binary
	Path       string
	MetaInfo   string
	Name       string
	CreateDate time.Time
	EditDate   time.Time
}

// AddBinaryFields - функция для заполения установки бинарных данных пользователем.
func AddBinaryFields(form *tview.Form, dataInfo *DataInfo) {
	form.AddInputField("Имя данных", "", 20, nil, func(text string) {
		if text != "" {
			dataInfo.Name = text // Разрешаю ввод, только если установлено имя данных
		}
	})
	form.AddInputField("Путь к файлу", "", 20, nil, func(text string) {
		if text != "" {
			dataInfo.Path = text // Разрешаю ввод, только если установлен путь к файлу
		}
	})
	form.AddInputField("Описание", "", 20, nil, func(text string) { dataInfo.MetaInfo = text })
}

// JSONEncode - функция для сериализации данных банковской карты.
func JSONEncode(dataInfo *DataInfo) (*repoData.Data, error) {
	// Проверка валидности введенных данных
	err := validateCardData(dataInfo)
	if err != nil {
		return nil, fmt.Errorf("invalid binary data, %w", err)
	}

	// преобразую содержимое файла в бинарный вид
	err = parseFile(dataInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file to binary, %w", err)
	}

	// сериализую данные типа "BINARY"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.Binary); err != nil {
		return nil, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	return &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.BINARY,
		Name:       dataInfo.Name,
		Metainfo:   dataInfo.MetaInfo,
		Status:     repoData.NEW,
		CreateDate: dataInfo.CreateDate,
		EditDate:   dataInfo.EditDate,
	}, nil
}

// validateCardData - функция для проверки корректности установленных бинарных данных.
func validateCardData(dataInfo *DataInfo) error {
	// Проверка имени карты
	if dataInfo.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	return nil
}

// parseFile - функция для преобразования содержимого файла в бинарный вид
func parseFile(dataInfo *DataInfo) error {
	// Определяю тип файла
	fileType, err := binary.GetFileType(dataInfo.Path)
	if err != nil {
		return fmt.Errorf("error getting file type, %w", err)
	}
	// Преобразую указанный файл в бинарный вид
	d, err := os.ReadFile(dataInfo.Path)
	if err != nil {
		return fmt.Errorf("failed to read file, %w", err)
	}
	// Устанавливаю полученные параметры в переменную
	dataInfo.Binary.Type = fileType
	dataInfo.Binary.Binary = d

	return nil
}
