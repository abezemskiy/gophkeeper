package add

import (
	"gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

// AddPage позволяет выбрать тип данных для добавления.
func AddPage(app *app.App) tview.Primitive {
	form := tview.NewForm()

	form.AddDropDown("Тип данных", []string{"PASSWORD", "TEXT", "BINARY", "BANKCARD"}, 0, func(option string, index int) {
		switch option {
		case "PASSWORD":
			app.SwitchTo("add_password")
		case "TEXT":
			app.SwitchTo("add_text")
		case "BINARY":
			app.SwitchTo("add_binary")
		case "BANKCARD":
			app.SwitchTo("add_bankcard")
		}
	})

	form.AddButton("Отмена", func() { app.SwitchTo("home") })

	form.SetBorder(true).SetTitle("Добавить данные")

	return form
}
