package bankcard

import (
	"context"
	"testing"

	"github.com/abezemskiy/gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestAddPasswordPage(t *testing.T) {
	// Создаем тестовое приложение
	testApp := &app.App{}

	// Создаем страницу ввода пароля
	passwordPage := AddBankcardPage(context.Background(), "some/url", nil, nil, nil)(testApp)

	// Проверяем, что это форма
	form, ok := passwordPage.(*tview.Form)
	assert.True(t, ok, "AddPasswordPage must return *tview.Form")

	// Проверяем количество полей в форме (6 полей)
	assert.Equal(t, 7, form.GetFormItemCount(), "Form must containe 6 fields and 2 buttons")

	// Проверяю названия элементов--------------------------------------------------------------
	label := form.GetFormItem(0).GetLabel()
	assert.Equal(t, "Имя данных", label, "First element of field must be named as Имя данных")

	label = form.GetFormItem(1).GetLabel()
	assert.Equal(t, "Номер карты", label)

	label = form.GetFormItem(2).GetLabel()
	assert.Equal(t, "Месяц", label)

	label = form.GetFormItem(3).GetLabel()
	assert.Equal(t, "Год", label)

	label = form.GetFormItem(4).GetLabel()
	assert.Equal(t, "CVV", label)

	label = form.GetFormItem(5).GetLabel()
	assert.Equal(t, "Имя владельца", label)

	label = form.GetFormItem(6).GetLabel()
	assert.Equal(t, "Описание", label)

	// Симулирую ввод данных в поля---------------------------------------------------------------
	{
		// устанавливаю кооректное имя данных
		field0 := form.GetFormItem(0).(*tview.InputField)
		message0 := "some data name"
		field0.SetText(message0)
		assert.Equal(t, message0, field0.GetText())
	}
	{
		// устанавливаю кооректный номер карты
		field1 := form.GetFormItem(1).(*tview.InputField)
		message1 := "1234567891234567"
		field1.SetText(message1)
		// Проверяю, что номер карты сохранился в поле
		assert.Equal(t, message1, field1.GetText())
	}
	{
		// устанавливаю корректный месяц
		field2 := form.GetFormItem(2).(*tview.InputField)
		message2 := "02"
		field2.SetText(message2)
		assert.Equal(t, message2, field2.GetText())
	}
	{
		// устанавливаю корректный год
		field := form.GetFormItem(3).(*tview.InputField)
		message := "35"
		field.SetText(message)
		assert.Equal(t, message, field.GetText())
	}
	{
		// устанавливаю корректный CVV
		field := form.GetFormItem(4).(*tview.InputField)
		message := "333"
		field.SetText(message)
		assert.Equal(t, message, field.GetText())
	}
	{
		// устанавливаю корректное имя владельца
		field := form.GetFormItem(5).(*tview.InputField)
		message := "SOMEOWNER NAME"
		field.SetText(message)
		assert.Equal(t, message, field.GetText())
	}
	{
		// устанавливаю корректное описание
		field := form.GetFormItem(6).(*tview.InputField)
		message := "some description"
		field.SetText(message)
		assert.Equal(t, message, field.GetText())
	}

	// Получаем кнопки
	saveButton := form.GetButton(0)
	cancelButton := form.GetButton(1)

	assert.Equal(t, "Сохранить", saveButton.GetLabel(), "Первая кнопка должна быть 'Сохранить'")
	assert.Equal(t, "Отмена", cancelButton.GetLabel(), "Вторая кнопка должна быть 'Отмена'")
}
