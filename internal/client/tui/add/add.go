package add

import (
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

// AddData позволяет выбрать тип данных для добавления.
func AddData(app *app.App) tview.Primitive {
	form := tview.NewForm()

	form.AddDropDown("Тип данных", []string{"PASSWORD", "TEXT", "BINARY", "BANKCARD"}, 0, func(option string, index int) {
		switch option {
		case "PASSWORD":
			app.SwitchTo(tui.AddPassword)
		case "TEXT":
			app.SwitchTo(tui.AdddText)
		case "BINARY":
			app.SwitchTo(tui.AddBinary)
		case "BANKCARD":
			app.SwitchTo(tui.AddBankCard)
		}
	})

	form.AddButton("Отмена", func() { app.SwitchTo(tui.Data) })

	form.SetBorder(true).SetTitle("Добавить данные")

	return form
}
