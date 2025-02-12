package home

import (
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

// Page - приветственное окно для входа в приложение.
func Page(app *app.App) tview.Primitive {
	list := tview.NewList().
		AddItem("Регистрация", "", 'a', func() { app.SwitchTo(tui.Register) }).
		AddItem("Авторизация", "", 'q', func() { app.SwitchTo(tui.Login) })

	list.SetBorder(true).SetTitle("Добро пожаловать в gophkeeper")

	return list
}
