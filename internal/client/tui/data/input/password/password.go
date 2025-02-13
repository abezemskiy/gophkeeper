package password

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
	Pass       data.Password
	MetaInfo   string
	Name       string
	CreateDate time.Time
	EditDate   time.Time
}

// AddPasswordFields - функция для заполения данных пароля пользователем.
func AddPasswordFields(form *tview.Form, dataInfo *DataInfo) {
	form.AddInputField("Имя данных", "", 20, nil, func(text string) { dataInfo.Name = text })
	form.AddInputField("Логин", "", 20, nil, func(text string) { dataInfo.Pass.Login = text })
	form.AddPasswordField("Пароль", "", 20, '*', func(text string) { dataInfo.Pass.Password = text })
	form.AddInputField("Описание", "", 20, nil, func(text string) { dataInfo.MetaInfo = text })
}

// JSONEncode - функция для сериализации данных банковской карты.
func JSONEncode(dataInfo *DataInfo) (*repoData.Data, error) {
	// Проверка валидности введенных данных
	err := validateCardData(dataInfo)
	if err != nil {
		return nil, fmt.Errorf("invalid password data, %w", err)
	}

	// сериализую данные типа "PASSWORD"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.Pass); err != nil {
		return nil, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	return &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.PASSWORD,
		Name:       dataInfo.Name,
		Metainfo:   dataInfo.MetaInfo,
		Status:     repoData.NEW,
		CreateDate: dataInfo.CreateDate,
		EditDate:   dataInfo.EditDate,
	}, nil
}

// validateCardData - функция для проверки корректности установленных данных пароля.
func validateCardData(dataInfo *DataInfo) error {
	// Проверка имени карты
	if dataInfo.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	return nil
}
