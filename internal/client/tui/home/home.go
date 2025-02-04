package tui

import (
	"gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

// HomePage создаёт экран с данными пользователя.
func HomePage(app *app.App) tview.Primitive {
	list := tview.NewList().
		AddItem("Добавить данные", "", 'a', func() { app.SwitchTo("add") }).
		AddItem("Выйти", "", 'q', func() { app.SwitchTo("login") })

	list.SetBorder(true).SetTitle("Ваши данные")

	return list
}
