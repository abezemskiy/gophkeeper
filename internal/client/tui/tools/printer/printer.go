package printer

import (
	"github.com/abezemskiy/gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
)

// Error - функция для вывода ошибок на экран пользователя.
func Error(app *app.App, message string) {
	go func() {
		app.App.QueueUpdateDraw(func() {
			modal := tview.NewModal().
				SetText("Ошибка: " + message).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(_ int, _ string) {
					app.Pages.RemovePage("error")
				})
			app.Pages.AddPage("error", modal, true, true)
		})
	}()
}

// Message - функция для вывода сообщения на экран пользователя.
func Message(app *app.App, message string) {
	go func() {
		app.App.QueueUpdateDraw(func() {
			modal := tview.NewModal().
				SetText("Сообщение: " + message).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(_ int, _ string) {
					app.Pages.RemovePage("message")
				})
			app.Pages.AddPage("message", modal, true, true)
		})
	}()
}
