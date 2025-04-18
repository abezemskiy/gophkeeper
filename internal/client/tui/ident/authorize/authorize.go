package authorize

import (
	"context"
	"fmt"

	"github.com/abezemskiy/gophkeeper/internal/client/handlers"
	"github.com/abezemskiy/gophkeeper/internal/client/identity"
	"github.com/abezemskiy/gophkeeper/internal/client/logger"
	"github.com/abezemskiy/gophkeeper/internal/client/tui"
	"github.com/abezemskiy/gophkeeper/internal/client/tui/app"
	"github.com/abezemskiy/gophkeeper/internal/client/tui/tools/printer"

	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// Page - страница авторизации пользователя.
func Page(ctx context.Context, ident identity.ClientIdentifier, info identity.IUserInfoStorage) func(app *app.App) tview.Primitive {
	return func(app *app.App) tview.Primitive {
		form := tview.NewForm()
		authData := &identity.AuthData{}

		form.AddInputField("Логин", "", 20, nil, func(text string) { authData.Login = text })
		form.AddPasswordField("Пароль", "", 20, '*', func(text string) { authData.Password = text })

		form.AddButton("Войти", func() {
			// authData содержит введенные логин и пароль
			if authData.Login == "" || authData.Password == "" {
				logger.ClientLog.Error("login or password can't be empty", zap.String("login", authData.Login))
				printer.Error(app, "login or password can't be empty")

				// Переключаю пользователя обратно на страницу авторизации
				app.SwitchTo(tui.Login)
				return
			}
			// Авторизирую пользователя
			correctPass, isRegister, err := handlers.Authorize(ctx, authData, ident, info)
			if err != nil {
				logger.ClientLog.Error("authorize client error", zap.String("error", error.Error(err)))
				printer.Error(app, fmt.Sprintf("authorize client error, %v", err))

				// Переключаю пользователя обратно на страницу авторизации
				app.SwitchTo(tui.Login)
				return
			}
			// Пользователь с данным логином не зарегистрирован
			if !isRegister {
				logger.ClientLog.Error("user not register", zap.String("login", authData.Login))
				printer.Error(app, "user not register")

				// Переключаю пользователя обратно на страницу авторизации
				app.SwitchTo(tui.Login)
				return
			}
			// Пароль неверный
			if !correctPass {
				logger.ClientLog.Error("wrong password", zap.String("login", authData.Login))
				printer.Error(app, "wrong password")

				// Переключаю пользователя обратно на страницу авторизации
				app.SwitchTo(tui.Login)
				return
			}
			// Авторизация прошла успешно, переключаю пользователя на страницу с его данными
			app.SwitchTo(tui.Data)
		})

		form.AddButton("Назад", func() { app.SwitchTo(tui.Home) })
		form.AddButton("Выход", func() { app.App.Stop() })

		form.SetBorder(true).SetTitle("Авторизация").SetTitleAlign(tview.AlignCenter)

		return tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(form, 10, 1, true)
	}
}
