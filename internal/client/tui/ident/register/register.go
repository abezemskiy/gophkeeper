package register

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/handlers"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"
	"gophkeeper/internal/client/tui/tools/printer"

	"github.com/go-resty/resty/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// Page - страница регистрации пользователя.
// Для успешной регистрации обязательно быть онлайн.
func Page(ctx context.Context, ident identity.ClientIdentifier,
	url string, client *resty.Client) func(app *app.App) tview.Primitive {

	return func(app *app.App) tview.Primitive {
		form := tview.NewForm()
		authData := &identity.AuthData{}
		var confirmPassword string

		form.AddInputField("Логин", "", 20, nil, func(text string) { authData.Login = text })
		form.AddPasswordField("Пароль", "", 20, '*', func(text string) { authData.Password = text })
		form.AddPasswordField("Подтвердите пароль", "", 20, '*', func(text string) { confirmPassword = text })

		form.AddButton("Зарегистрироваться", func() {
			// authData содержит введенные логин и пароль
			if authData.Login == "" || authData.Password == "" {
				logger.ClientLog.Error("login or password can't be empty", zap.String("login", authData.Login))
				printer.Error(app, "login or password can't be empty")

				// Переключаю пользователя обратно на страницу регистрации
				app.SwitchTo(tui.Register)
				return
			}
			// Введенные пароли не совпадают
			if authData.Password != confirmPassword {
				logger.ClientLog.Error("passwords do not match", zap.String("login", authData.Login))
				printer.Error(app, "passwords do not match")

				// Очистка полей формы для повторной попытки регистрации
				form.GetFormItemByLabel("Пароль").(*tview.InputField).SetText("")
				form.GetFormItemByLabel("Подтвердите пароль").(*tview.InputField).SetText("")

				// Переключаю пользователя обратно на страницу регистрации
				app.SwitchTo(tui.Register)
				return
			}

			// регистрирую нового пользователя
			ok, err := handlers.Register(ctx, url, authData, client, ident)
			if err != nil {
				logger.ClientLog.Error("failed to register new user", zap.String("login", authData.Login), zap.String("error", err.Error()))
				printer.Error(app, fmt.Sprintf("failed to register new user, %v", err))

				// Очистка полей формы для повторной попытки регистрации
				form.GetFormItemByLabel("Логин").(*tview.InputField).SetText("")
				form.GetFormItemByLabel("Пароль").(*tview.InputField).SetText("")
				form.GetFormItemByLabel("Подтвердите пароль").(*tview.InputField).SetText("")

				// Переключаю пользователя обратно на страницу регистрации
				app.SwitchTo(tui.Register)
				return
			}
			// Данный пользователь уже зарегистрирован в системе
			if !ok {
				logger.ClientLog.Error("user already register", zap.String("login", authData.Login))
				printer.Error(app, "user already register")

				// Переключаю пользователя на приветственную страницу
				app.SwitchTo(tui.Home)
				return
			}
			// Успешная регистрация пользователя
			logger.ClientLog.Info("new user successfully register", zap.String("login", authData.Login))
			printer.Message(app, "new user successfully register")

			// Переключаю пользователя на страницу авторизации
			app.SwitchTo(tui.Login)
		})

		form.AddButton("Назад", func() { app.SwitchTo(tui.Home) })
		form.AddButton("Выход", func() { app.App.Stop() })

		form.SetBorder(true).SetTitle("Регистрация").SetTitleAlign(tview.AlignCenter)

		return tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(form, 20, 1, true)
	}
}
