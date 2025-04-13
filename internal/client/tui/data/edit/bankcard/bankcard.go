package bankcard

import (
	"context"
	"fmt"
	"time"

	"github.com/abezemskiy/gophkeeper/internal/client/handlers"
	"github.com/abezemskiy/gophkeeper/internal/client/identity"
	"github.com/abezemskiy/gophkeeper/internal/client/logger"
	"github.com/abezemskiy/gophkeeper/internal/client/storage"
	"github.com/abezemskiy/gophkeeper/internal/client/tui"
	"github.com/abezemskiy/gophkeeper/internal/client/tui/app"
	input "github.com/abezemskiy/gophkeeper/internal/client/tui/data/input/bankcard"
	"github.com/abezemskiy/gophkeeper/internal/client/tui/tools/printer"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// EditBankcardPage - TUI страница добавления нового пароля пользователя.
func EditBankcardPage(ctx context.Context, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	info identity.IUserInfoStorage) func(app *app.App) tview.Primitive {

	return func(app *app.App) tview.Primitive {
		form := tview.NewForm()
		// структура для сохранения данных банковской карты
		dataInfo := &input.DataInfo{
			CreateDate: time.Now(),
			EditDate:   time.Now(),
		}

		// Создаю поля для заполенения данных банковской карты
		input.Fields(form, dataInfo)

		form.AddButton("Изменить", func() {
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

				app.SwitchTo(tui.EditBankCard)
				return
			}

			// Меняю данные в хранилище на новые
			ok, err := handlers.ReplaceData(ctx, id, url, authData.Password, client, stor, userData)
			if err != nil {
				logger.ClientLog.Error("replace data error", zap.String("error", error.Error(err)))
				printer.Error(app, fmt.Sprintf("replace data error, %v", err))

				app.SwitchTo(tui.EditBankCard)
				return
			}
			if !ok {
				logger.ClientLog.Error("data is not exists", zap.String("name", dataInfo.Name))
				printer.Error(app, fmt.Sprintf("data is not exists, name %s", dataInfo.Name))

				app.SwitchTo(tui.EditBankCard)
				return
			}

			// Печатаю сообщение об успешном сохранении данных
			printer.Message(app, "data replace successfully")

			// перенаправляю пользователя на страницу данных
			app.SwitchTo(tui.Edit)
		})
		form.AddButton("Отмена", func() { app.SwitchTo(tui.Edit) })

		form.SetBorder(true).SetTitle("Изменить данные банковской карты")
		return form
	}
}
