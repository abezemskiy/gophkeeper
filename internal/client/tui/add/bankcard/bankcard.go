package bankcard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/encr"
	"gophkeeper/internal/client/handlers"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/client/storage/data"
	"gophkeeper/internal/client/tui/app"
	"gophkeeper/internal/client/tui/tools/printer"
	repoData "gophkeeper/internal/repositories/data"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// dataInfo - вспомогательная структура для передачи полученных от пользоавтеля данных в функцию сохранения данных в сервисе.
type dataInfo struct {
	bank       data.Bank
	metaInfo   string
	name       string
	createDate time.Time
	editDate   time.Time
}

// AddPasswordPage - TUI страница добавления нового пароля пользователя.
func AddPasswordPage(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	passStor identity.IPasswordStorage, app app.App) tview.Primitive {

	form := tview.NewForm()
	// структура для введенной пары логин пароль
	dataInfo := dataInfo{
		createDate: time.Now(),
		editDate:   time.Now(),
	}

	form.AddInputField("Имя данных", "", 20, func(text string, _ rune) bool {
		return text != "" // Разрешаю ввод, только если установлено имя данных
	}, func(text string) {
		if text != "" {
			dataInfo.name = text // Разрешаю ввод, только если установлено имя данных
		}
	})

	form.AddInputField("Номер карты", "", 16, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil // Разрешаю ввод только если это число
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.bank.Number = num
		}
	})
	form.AddInputField("Месяц", "", 2, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil // Разрешаю ввод только если это число
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.bank.Mounth = int(num)
		}
	})
	form.AddInputField("Год", "", 2, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil // Разрешаю ввод только если это число
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.bank.Year = int(num)
		}
	})
	form.AddInputField("CVV", "", 3, func(text string, _ rune) bool {
		_, err := strconv.ParseInt(text, 10, 64)
		return err == nil // Разрешаю ввод только если это число
	}, func(text string) {
		num, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			dataInfo.bank.Year = int(num)
		}
	})
	form.AddInputField("Имя владельца", "", 20, func(text string, _ rune) bool {
		return text != "" // Разрешаю ввод, только если установлено имя данных
	}, func(text string) { dataInfo.bank.Owner = text })

	form.AddInputField("Описание", "", 20, nil, func(text string) { dataInfo.metaInfo = text })

	form.AddButton("Сохранить", func() {
		// проверяю наличие в приложении мастер пароля
		masterPass := passStor.Get()
		if masterPass == "" {
			// мастер пароль не установлен, возвращаю пользователя на страницу аутентификации.
			app.SwitchTo("login")
		}

		ok, err := save(ctx, userID, url, client, stor, dataInfo, masterPass)
		if err != nil {
			logger.ClientLog.Error("save data error", zap.String("error", error.Error(err)))
			printer.Error(app, fmt.Sprintf("save data error, %v", err))
		}
		if !ok {
			logger.ClientLog.Error("data is not unique", zap.String("name", dataInfo.name))
			printer.Error(app, fmt.Sprintf("data is not unique, name %s", dataInfo.name))
		}

		// Печатаю сообщение об успешном сохранении данных
		printer.Message(app, "data saved successfully")

		// перенаправляю пользователя на страницу данных
		app.SwitchTo("home")
	})
	form.AddButton("Отмена", func() { app.SwitchTo("add") })

	form.SetBorder(true).SetTitle("Добавить данные банковской карты")
	return form
}

func save(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	dataInfo dataInfo, masterPass string) (bool, error) {
	// сериализую данные типа "BANKCARD"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.bank); err != nil {
		return false, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	dataToEncr := &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.BANKCARD,
		Name:       dataInfo.name,
		Metainfo:   dataInfo.metaInfo,
		Status:     repoData.NEW,
		CreateDate: dataInfo.createDate,
		EditDate:   dataInfo.editDate,
	}

	// шифрую данные с помощью мастер пароля пользователя
	encrData, err := encr.EncryptData(masterPass, dataToEncr)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt data, %w", err)
	}

	// Сохраняю данные в хранилище
	ok, err := handlers.SaveEncryptedData(ctx, userID, url, client, stor, encrData)
	if err != nil {
		return false, fmt.Errorf("failed to save data, %w", err)
	}
	if !ok {
		return false, nil
	}
	return true, nil
}
