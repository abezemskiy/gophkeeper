package text

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/storage/data"
	repoData "gophkeeper/internal/repositories/data"
	"time"

	"github.com/rivo/tview"
)

// DataInfo - вспомогательная структура для передачи полученных от пользоавтеля данных в функцию сохранения данных в сервисе.
type DataInfo struct {
	Text       data.Text
	MetaInfo   string
	Name       string
	CreateDate time.Time
	EditDate   time.Time
}

// Fields - функция для заполения данных банковской карты пользователем.
func Fields(form *tview.Form, dataInfo *DataInfo) {
	form.AddInputField("Имя данных", "", 20, nil, func(text string) { dataInfo.Name = text })
	form.AddInputField("Текст", "", 100, nil, func(text string) { dataInfo.Text.Text = text })
	form.AddInputField("Описание", "", 20, nil, func(text string) { dataInfo.MetaInfo = text })
}

// JSONEncode - функция для сериализации текстовых данных.
func JSONEncode(dataInfo *DataInfo) (*repoData.Data, error) {
	// Проверка валидности введенных данных
	err := validateCardData(dataInfo)
	if err != nil {
		return nil, fmt.Errorf("invalid text data, %w", err)
	}

	// сериализую данные типа "TEXT"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.Text); err != nil {
		return nil, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	return &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.TEXT,
		Name:       dataInfo.Name,
		Metainfo:   dataInfo.MetaInfo,
		Status:     repoData.NEW,
		CreateDate: dataInfo.CreateDate,
		EditDate:   dataInfo.EditDate,
	}, nil
}

// validateCardData - функция для проверки корректности установленных текстовых данных.
func validateCardData(dataInfo *DataInfo) error {
	// Проверка имени карты
	if dataInfo.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	return nil
}
