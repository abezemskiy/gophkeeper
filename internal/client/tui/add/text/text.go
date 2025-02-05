package text

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// dataInfo - вспомогательная структура для передачи полученных от пользоавтеля данных в функцию сохранения данных в сервисе.
type dataInfo struct {
	text       data.Text
	metaInfo   string
	name       string
	createDate time.Time
	editDate   time.Time
}

// AddTextPage - TUI страница добавления нового текста пользователя.
func AddTextPage(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	passStor identity.IPasswordStorage, app app.App) tview.Primitive {

	form := tview.NewForm()
	// структура для введенной пары логин пароль
	dataInfo := dataInfo{
		createDate: time.Now(),
		editDate:   time.Now(),
	}

	form.AddInputField("Имя данных", "", 20, nil, func(text string) { dataInfo.name = text })
	form.AddInputField("Текст", "", 20, nil, func(text string) { dataInfo.text.Text = text })
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

	form.SetBorder(true).SetTitle("Добавить текст")
	return form
}

func save(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	dataInfo dataInfo, masterPass string) (bool, error) {
	// проверяю корректность сохраняемого текста
	if dataInfo.text.Text == "" {
		return false, errors.New("text can't be emphty")
	}

	// сериализую данные типа "TEXT"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.text); err != nil {
		return false, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	dataToEncr := &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.TEXT,
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
