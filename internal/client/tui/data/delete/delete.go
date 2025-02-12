package delete

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/handlers"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"
	"gophkeeper/internal/client/tui/tools/printer"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// Delete - TUI страница для удаления данных пользователя по имени данных.
func Delete(ctx context.Context, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	info identity.IUserInfoStorage) func(app *app.App) tview.Primitive {

	return func(app *app.App) tview.Primitive {
		form := tview.NewForm()
		var dataName string

		form.AddInputField("Имя данных", "", 20, nil, func(text string) { dataName = text })

		form.AddButton("Удалить", func() {
			// проверяю наличие в приложении мастер пароля
			authData, id := info.Get()
			if authData.Password == "" || authData.Login == "" || id == "" {
				printer.Message(app, "password or login not set")

				// мастер пароль не установлен, возвращаю пользователя на страницу аутентификации.
				app.SwitchTo(tui.Login)
				return
			}

			// Имя данных не задано
			if dataName == "" {
				printer.Error(app, "data name is not set")

				app.SwitchTo(tui.Delete)
				return
			}

			// Удаляю данные
			ok, err := handlers.DeleteEncryptedData(ctx, id, url, dataName, client, stor)

			if err != nil {
				logger.ClientLog.Error("delete data error", zap.String("error", error.Error(err)))
				printer.Error(app, fmt.Sprintf("delete data error, %v", err))

				app.SwitchTo(tui.Delete)
				return
			}
			if !ok {
				logger.ClientLog.Error("data is not exists", zap.String("name", dataName))
				printer.Error(app, fmt.Sprintf("data is not exists, name %s", dataName))

				app.SwitchTo(tui.Delete)
				return
			}

			// Печатаю сообщение об успешном удалении данных
			printer.Message(app, "data delete successfully")

			// перенаправляю пользователя на страницу данных
			app.SwitchTo(tui.Data)
		})
		form.AddButton("Отмена", func() { app.SwitchTo(tui.Data) })

		form.SetBorder(true).SetTitle("Удаление данных")
		return form
	}
}
