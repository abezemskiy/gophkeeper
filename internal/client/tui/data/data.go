package data

import (
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

// DataPage создаёт экран с данными пользователя.
func DataPage(app *app.App) tview.Primitive {
	list := tview.NewList().
		AddItem("Добавить данные", "", 'a', func() { app.SwitchTo(tui.Add) }).
		AddItem("Посмотреть данные", "", 'a', func() { app.SwitchTo(tui.View) }).
		AddItem("Выйти", "", 'q', func() { app.SwitchTo(tui.Login) })

	list.SetBorder(true).SetTitle("Ваши данные")

	return list
}
