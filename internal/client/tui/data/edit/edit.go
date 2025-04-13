package edit

import (
	"github.com/abezemskiy/gophkeeper/internal/client/tui"
	"github.com/abezemskiy/gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

// Edit позволяет выбрать тип данных для изменения.
func Edit(app *app.App) tview.Primitive {
	form := tview.NewForm()

	form.AddDropDown("Тип данных", []string{"PASSWORD", "TEXT", "BINARY", "BANKCARD"}, 0, func(option string, _ int) {
		switch option {
		case "PASSWORD":
			app.SwitchTo(tui.EditPassword)
		case "TEXT":
			app.SwitchTo(tui.EditText)
		case "BINARY":
			app.SwitchTo(tui.EditBinary)
		case "BANKCARD":
			app.SwitchTo(tui.EditBankCard)
		}
	})

	form.AddButton("Отмена", func() { app.SwitchTo(tui.Data) })

	form.SetBorder(true).SetTitle("Изменить данные")

	return form
}
