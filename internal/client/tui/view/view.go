package view

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/client/storage/data"
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"
	"gophkeeper/internal/client/tui/tools/printer"
	repoData "gophkeeper/internal/repositories/data"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Page - страница отображения данных пользователя.
func Page(_ context.Context, decrData storage.IStorage) func(app *app.App) tview.Primitive {
	return func(app *app.App) tview.Primitive {
		// Создаю элементы интерфейса
		list := tview.NewList()
		table := tview.NewTable().SetBorders(true)

		// Кнопка "Обновить" для обновления данных на странице
		updateFunc := func() {
			list.Clear()
			table.Clear()

			// Извлекаю данные пользователя из inmemory хранилища
			data := decrData.GetAll()

			if len(data) == 0 {
				go func() {
					app.App.QueueUpdateDraw(func() {
						printer.Message(app, "data not added yet")
					})
				}()
				return
			}

			// Заполнение списка имен данных
			//var firstVersions []repoData.Data
			for i, versions := range data {
				if len(versions) > 0 {
					name := versions[0].Name
					list.AddItem(name, "", rune('a'+i), func() {
						err := updateTable(table, versions)
						if err != nil {
							printer.Message(app, fmt.Errorf("failed to update table, %w", err).Error())
						}
					})
				}
			}
			// Устанавливаем фокус на список данных
			app.App.SetFocus(list)
		}

		// Кнопки
		backButton := tview.NewButton("Назад")
		updateButton := tview.NewButton("Обновить")

		// Контейнер с двумя панелями и кнопками
		flex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().AddItem(list, 30, 1, true).
				AddItem(table, 0, 2, false), 0, 1, true)

		buttons := tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(updateButton, 12, 1, true).
			AddItem(backButton, 12, 1, false)

		flex.AddItem(buttons, 3, 1, true)

		// фокус на кнопку "Обновить"
		app.App.SetFocus(updateButton)

		// цвет фона для выделенного элемента списка
		list.SetSelectedBackgroundColor(tcell.ColorBlue)

		// Переключение фокуса с помощью Tab
		flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab: // Циклический переход фокуса между элементами
				switch app.App.GetFocus() {
				case updateButton:
					app.App.SetFocus(backButton)
				case backButton:
					app.App.SetFocus(list)
				case list:
					app.App.SetFocus(updateButton)
				}
			case tcell.KeyEnter: // Обработка нажатий кнопок
				if app.App.GetFocus() == updateButton {
					updateFunc()
				} else if app.App.GetFocus() == backButton {
					app.Pages.SwitchToPage(tui.Data)
				}
			case tcell.KeyEsc: // Выход на предыдущую страницу
				app.Pages.SwitchToPage(tui.Data)
			}
			return event
		})

		return flex
	}
}

// updateTable - функция для обновления таблицы с версиями данных.
func updateTable(table *tview.Table, versions []repoData.Data) error {
	table.Clear()
	table.SetCell(0, 0, tview.NewTableCell("Версия").SetSelectable(false).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	table.SetCell(0, 1, tview.NewTableCell("Данные").SetSelectable(false).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	table.SetCell(0, 2, tview.NewTableCell("Описание").SetSelectable(false).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	table.SetCell(0, 3, tview.NewTableCell("Создано").SetSelectable(false).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	table.SetCell(0, 4, tview.NewTableCell("Изменено").SetSelectable(false).SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))

	for i, v := range versions {
		dataString, err := parseData(v)
		if err != nil {
			return fmt.Errorf("failed to parse data, %w", err)
		}
		table.SetCell(i+1, 0, tview.NewTableCell(fmt.Sprintf("v%d", i+1)).SetSelectable(true))
		table.SetCell(i+1, 1, tview.NewTableCell(dataString).SetSelectable(true))
		table.SetCell(i+1, 2, tview.NewTableCell(v.Metainfo).SetSelectable(true))
		table.SetCell(i+1, 3, tview.NewTableCell(v.CreateDate.Format("02.01.2006 15:04:05")).SetSelectable(true))
		table.SetCell(i+1, 4, tview.NewTableCell(v.EditDate.Format("02.01.2006 15:04:05")).SetSelectable(true))
	}
	return nil
}

// parseData - функция для парсинга данных по типу.
func parseData(d repoData.Data) (string, error) {
	switch d.Type {
	case repoData.PASSWORD:
		var p data.Password
		err := json.Unmarshal(d.Data, &p)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal data, %w", err)
		}
		return fmt.Sprintf("Login: %s, Password: %s", p.Login, p.Password), nil
	case repoData.TEXT:
		var t data.Text
		err := json.Unmarshal(d.Data, &t)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal data, %w", err)
		}
		return t.Text, nil
	case repoData.BINARY:
		var b data.Binary
		err := json.Unmarshal(d.Data, &b)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal data, %w", err)
		}
		return fmt.Sprintf("Binary (%s)", b.Type), nil
	case repoData.BANKCARD:
		var b data.Bank
		err := json.Unmarshal(d.Data, &b)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal data, %w", err)
		}
		return fmt.Sprintf("Card: %d, Owner: %s", b.Number, b.Owner), nil
	default:
		return "", errors.New("unknown data type")
	}
}
