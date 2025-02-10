package password

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	pass       data.Password
	metaInfo   string
	name       string
	createDate time.Time
	editDate   time.Time
}

// AddPasswordPage - TUI страница добавления нового пароля пользователя.
func AddPasswordPage(ctx context.Context, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	info identity.IUserInfoStorage, app *app.App) tview.Primitive {

	form := tview.NewForm()
	// структура для введенной пары логин пароль
	dataInfo := dataInfo{
		createDate: time.Now(),
		editDate:   time.Now(),
	}

	form.AddInputField("Имя данных", "", 20, nil, func(text string) { dataInfo.name = text })
	form.AddInputField("Логин", "", 20, nil, func(text string) { dataInfo.pass.Login = text })
	form.AddPasswordField("Пароль", "", 20, '*', func(text string) { dataInfo.pass.Password = text })
	form.AddInputField("Описание", "", 20, nil, func(text string) { dataInfo.metaInfo = text })

	form.AddButton("Сохранить", func() {
		// проверяю наличие в приложении мастер пароля
		authData, id := info.Get()
		if authData.Password == "" {
			// мастер пароль не установлен, возвращаю пользователя на страницу аутентификации.
			app.SwitchTo("login")
		}

		ok, err := save(ctx, id, url, client, stor, dataInfo, authData.Password)
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

	form.SetBorder(true).SetTitle("Добавить пароль")
	return form
}

func save(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	dataInfo dataInfo, masterPass string) (bool, error) {
	// проверяю корректность сохраняемого пароля
	if dataInfo.pass.Password == "" {
		return false, errors.New("password can't be emphty")
	}
	// проверяю корректность сохраняемого логина
	if dataInfo.pass.Login == "" {
		return false, errors.New("login can't be emphty")
	}

	// сериализую данные типа "PASSWORD"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(dataInfo.pass); err != nil {
		return false, fmt.Errorf("encode data error, %w", err)
	}

	// Создаю структуру типа data.Data
	userData := &repoData.Data{
		Data:       buf.Bytes(),
		Type:       repoData.PASSWORD,
		Name:       dataInfo.name,
		Metainfo:   dataInfo.metaInfo,
		Status:     repoData.NEW,
		CreateDate: dataInfo.createDate,
		EditDate:   dataInfo.editDate,
	}

	// Сохраняю данные в хранилище
	ok, err := handlers.SaveData(ctx, userID, url, masterPass, client, stor, userData)
	if err != nil {
		return false, fmt.Errorf("failed to save data, %w", err)
	}
	if !ok {
		return false, nil
	}
	return true, nil
}
