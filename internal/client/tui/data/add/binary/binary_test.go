package binary

import (
	"context"
	"testing"

	"gophkeeper/internal/client/tui/app"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestAddBinaryPage(t *testing.T) {
	// Создаем тестовое приложение
	testApp := &app.App{}

	// Создаем страницу ввода пароля
	passwordPage := AddBinaryPage(context.Background(), "some/url", nil, nil, nil)(testApp)

	// Проверяем, что это форма
	form, ok := passwordPage.(*tview.Form)
	assert.True(t, ok, "TestAddBinaryPage must return *tview.Form")

	// Проверяем количество полей в форме (3 поля)
	assert.Equal(t, 3, form.GetFormItemCount(), "Form must containe 3 fields and 2 buttons")

	// Проверяю названия элементов--------------------------------------------------------------
	label := form.GetFormItem(0).GetLabel()
	assert.Equal(t, "Имя данных", label, "First element of field must be named as Имя данных")

	label = form.GetFormItem(1).GetLabel()
	assert.Equal(t, "Путь к файлу", label)

	label = form.GetFormItem(2).GetLabel()
	assert.Equal(t, "Описание", label, "First element of field must be named as Описание")

	// Симулирую ввод данных в поля---------------------------------------------------------------
	field0 := form.GetFormItem(0).(*tview.InputField)
	message0 := "some data name"
	field0.SetText(message0)
	assert.Equal(t, message0, field0.GetText())

	field1 := form.GetFormItem(1).(*tview.InputField)
	message1 := "some path"
	field1.SetText(message1)
	assert.Equal(t, message1, field1.GetText())

	field2 := form.GetFormItem(2).(*tview.InputField)
	message2 := "some description"
	field2.SetText(message2)
	assert.Equal(t, message2, field2.GetText())

	// Получаем кнопки
	saveButton := form.GetButton(0)
	cancelButton := form.GetButton(1)

	assert.Equal(t, "Сохранить", saveButton.GetLabel(), "Первая кнопка должна быть 'Сохранить'")
	assert.Equal(t, "Отмена", cancelButton.GetLabel(), "Вторая кнопка должна быть 'Отмена'")
}
