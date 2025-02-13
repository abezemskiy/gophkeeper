package text

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/handlers"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"
	input "gophkeeper/internal/client/tui/data/input/text"
	"gophkeeper/internal/client/tui/tools/printer"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// AddTextPage - TUI страница добавления нового текста пользователя.
func AddTextPage(ctx context.Context, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	info identity.IUserInfoStorage) func(app *app.App) tview.Primitive {

	return func(app *app.App) tview.Primitive {
		form := tview.NewForm()
		// структура для введенной пары логин пароль
		dataInfo := &input.DataInfo{
			CreateDate: time.Now(),
			EditDate:   time.Now(),
		}

		// Создаю поля для заполенения данных пароля
		input.Fields(form, dataInfo)

		form.AddButton("Сохранить", func() {
			// проверяю наличие в приложении мастер пароля
			authData, id := info.Get()
			if authData.Password == "" {
				// мастер пароль не установлен, возвращаю пользователя на страницу аутентификации.
				app.SwitchTo(tui.Login)
				return
			}

			// Валидирую и сериализую данные для сохранения в сервисе
			userData, err := input.JSONEncode(dataInfo)
			if err != nil {
				logger.ClientLog.Error("encode data error", zap.String("error", error.Error(err)))
				printer.Error(app, fmt.Sprintf("encode data error, %v", err))

				app.SwitchTo(tui.AddText)
				return
			}

			// Сохраняю данные в хранилище
			ok, err := handlers.SaveData(ctx, id, url, authData.Password, client, stor, userData)
			if err != nil {
				logger.ClientLog.Error("save data error", zap.String("error", error.Error(err)))
				printer.Error(app, fmt.Sprintf("save data error, %v", err))

				app.SwitchTo(tui.AddText)
				return
			}
			if !ok {
				logger.ClientLog.Error("data is not unique", zap.String("name", dataInfo.Name))
				printer.Error(app, fmt.Sprintf("data is not unique, name %s", dataInfo.Name))

				app.SwitchTo(tui.AddText)
				return
			}

			// Печатаю сообщение об успешном сохранении данных
			printer.Message(app, "data saved successfully")

			// перенаправляю пользователя на страницу данных
			app.SwitchTo(tui.Data)
		})
		form.AddButton("Отмена", func() { app.SwitchTo(tui.Add) })

		form.SetBorder(true).SetTitle("Добавить текст")
		return form
	}
}
