package tui

import (
	"gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func LoginPage(app *app.App) tview.Primitive {
	form := tview.NewForm()
	authData := &Auth{}

	form.AddInputField("Логин", "", 20, nil, func(text string) { authData.Login = text })
	form.AddPasswordField("Пароль", "", 20, '*', func(text string) { authData.Password = text })

	form.AddButton("Войти", func() {
		// Теперь authData содержит введенные логин и пароль
		if authData.Login == "" || authData.Password == "" {
			modal := tview.NewModal().
				SetText("Ошибка: логин и пароль не могут быть пустыми").
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					app.Pages.RemovePage("error")
				})
			app.Pages.AddPage("error", modal, true, true)
		} else {
			app.SwitchTo("home")
		}
	})

	form.AddButton("Регистрация", func() { app.SwitchTo("home") })
	form.AddButton("Выход", func() { app.App.Stop() })

	form.SetBorder(true).SetTitle("Авторизация").SetTitleAlign(tview.AlignCenter)

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 10, 1, true)
}
