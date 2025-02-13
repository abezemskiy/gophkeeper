package bankcard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/storage/data"
	repoData "gophkeeper/internal/repositories/data"
	"strconv"
	"time"

	"github.com/rivo/tview"
)

// DataInfo - вспомогательная структура для передачи полученных от пользоавтеля данных в функцию сохранения данных в сервисе.
type DataInfo struct {
	Bank       data.Bank
	MetaInfo   string
	Name       string
	CreateDate time.Time
	EditDate   time.Time
}

// AddBankCardFields - функция для заполения данных банковской карты пользователем.
func AddBankCardFields(form *tview.Form, dataInfo *DataInfo) {
	form.AddInputField("Имя данных", "", 20, func(text string, _ rune) bool {
		return text != ""
	}, func(text string) {
		if text != "" {
			dataInfo.Name = text
		}
	})

	form.AddInputField("Номер карты", "", 16, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.Bank.Number = num
		}
	})

	form.AddInputField("Месяц", "", 2, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.Bank.Mounth = int(num)
		}
	})

	form.AddInputField("Год", "", 2, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.Bank.Year = int(num)
		}
	})

	form.AddInputField("CVV", "", 3, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.Bank.CVV = int(num)
		}
	})

	form.AddInputField("Имя владельца", "", 20, func(text string, _ rune) bool {
		return text != ""
	}, func(text string) {
		dataInfo.Bank.Owner = text
	})

	form.AddInputField("Описание", "", 20, nil, func(text string) {
		dataInfo.MetaInfo = text
	})
}

// JSONEncode - функция для сериализации данных банковской карты.
func JSONEncode(dataInfo *DataInfo) (*repoData.Data, error) {
	// Проверка валидности введенных данных
	err := validateCardData(dataInfo)
	if err != nil {
		return nil, fmt.Errorf("invalid card data, %w", err)
	}

	// сериализую данные типа "BANKCARD"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.Bank); err != nil {
		return nil, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	return &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.BANKCARD,
		Name:       dataInfo.Name,
		Metainfo:   dataInfo.MetaInfo,
		Status:     repoData.NEW,
		CreateDate: dataInfo.CreateDate,
		EditDate:   dataInfo.EditDate,
	}, nil
}

// validateCardData - функция для проверки корректности установленных данных банковской карты.
func validateCardData(dataInfo *DataInfo) error {
	// Проверка имени карты
	if dataInfo.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	return nil
}
