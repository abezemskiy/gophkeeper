package password

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
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

func AddPasswordPage(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	passStor identity.IPasswordStorage, app app.App) tview.Primitive {

	form := tview.NewForm()
	// структура для введенной пары логин пароль
	var pass data.Password
	var metaInfo string
	var name string

	form.AddInputField("Имя данных", "", 20, nil, func(text string) { name = text })
	form.AddInputField("Логин", "", 20, nil, func(text string) { pass.Login = text })
	form.AddPasswordField("Пароль", "", 20, '*', func(text string) { pass.Password = text })
	form.AddInputField("Описание", "", 20, nil, func(text string) { metaInfo = text })

	form.AddButton("Сохранить", func() {
		// проверяю корректность сохраняемого пароля
		if pass.Password == "" {
			printer.Error(app, "Password can't be emphty")
		}
		// проверяю корректность сохраняемого логина
		if pass.Login == "" {
			printer.Error(app, "Login can't be emphty")
		}

		// проверяю наличие в приложении мастер пароля
		masterPass := passStor.Get()
		if masterPass == "" {
			// мастер пароль не установлен, возвращаю пользователя на страницу аутентификации.
			app.SwitchTo("login")
		}

		// сериализую данные типа "PASSWORD"
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(metaInfo); err != nil {
			logger.ClientLog.Error("encode data error", zap.String("error", error.Error(err)))
			printer.Error(app, fmt.Errorf("encode data error, %v", err).Error())
		}

		// Создаю структуру типа data.Data
		dataToEncr := &repoData.Data{
			Data:       buf.Bytes(),
			Type:       repoData.PASSWORD,
			Name:       name,
			Metainfo:   metaInfo,
			Status:     repoData.NEW,
			CreateDate: time.Now(),
			EditDate:   time.Now(),
		}

		// шифрую данные с помощью мастер пароля пользователя
		encrData, err := encr.EncryptData(masterPass, dataToEncr)
		if err != nil {
			logger.ClientLog.Error("failed to encrypt data", zap.String("error", error.Error(err)))
			printer.Error(app, fmt.Errorf("failed to encrypt data, %v", err).Error())
		}

		// Сохраняю данные в хранилище
		ok, err := handlers.SaveEncryptedData(ctx, userID, url, client, stor, encrData)
		if err != nil {
			logger.ClientLog.Error("failed to save data", zap.String("error", error.Error(err)))
			printer.Error(app, fmt.Errorf("failed to save data, %v", err).Error())
		}
		if !ok {
			logger.ClientLog.Error("data is not unique", zap.String("name", name))
			printer.Error(app, fmt.Sprintf("data is not unique, name %s", name))
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

func Save(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	passStor identity.IPasswordStorage, app app.App) (bool, error){
// проверяю корректность сохраняемого пароля
if pass.Password == "" {
	printer.Error(app, "Password can't be emphty")
}
// проверяю корректность сохраняемого логина
if pass.Login == "" {
	printer.Error(app, "Login can't be emphty")
}

// проверяю наличие в приложении мастер пароля
masterPass := passStor.Get()
if masterPass == "" {
	// мастер пароль не установлен, возвращаю пользователя на страницу аутентификации.
	app.SwitchTo("login")
}

// сериализую данные типа "PASSWORD"
var buf bytes.Buffer
if err := json.NewEncoder(&buf).Encode(metaInfo); err != nil {
	logger.ClientLog.Error("encode data error", zap.String("error", error.Error(err)))
	printer.Error(app, fmt.Errorf("encode data error, %v", err).Error())
}

// Создаю структуру типа data.Data
dataToEncr := &repoData.Data{
	Data:       buf.Bytes(),
	Type:       repoData.PASSWORD,
	Name:       name,
	Metainfo:   metaInfo,
	Status:     repoData.NEW,
	CreateDate: time.Now(),
	EditDate:   time.Now(),
}

// шифрую данные с помощью мастер пароля пользователя
encrData, err := encr.EncryptData(masterPass, dataToEncr)
if err != nil {
	logger.ClientLog.Error("failed to encrypt data", zap.String("error", error.Error(err)))
	printer.Error(app, fmt.Errorf("failed to encrypt data, %v", err).Error())
}

// Сохраняю данные в хранилище
ok, err := handlers.SaveEncryptedData(ctx, userID, url, client, stor, encrData)
if err != nil {
	logger.ClientLog.Error("failed to save data", zap.String("error", error.Error(err)))
	printer.Error(app, fmt.Errorf("failed to save data, %v", err).Error())
}
if !ok {
	logger.ClientLog.Error("data is not unique", zap.String("name", name))
	printer.Error(app, fmt.Sprintf("data is not unique, name %s", name))
}

// Печатаю сообщение об успешном сохранении данных
printer.Message(app, "data saved successfully")

// перенаправляю пользователя на страницу данных
app.SwitchTo("home")
}
