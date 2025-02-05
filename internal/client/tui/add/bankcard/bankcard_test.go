package bankcard

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/client/storage/data"
	"gophkeeper/internal/client/tui/app"
	repoData "gophkeeper/internal/repositories/data"
	"gophkeeper/internal/repositories/mocks"

	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddPasswordPage(t *testing.T) {
	// Создаем тестовое приложение
	testApp := app.App{}

	// Создаем страницу ввода пароля
	passwordPage := AddPasswordPage(context.Background(), "some id", "some/url", nil, nil, nil, testApp)

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

func TestSave(t *testing.T) {
	// вспомогательная функция
	testHandler := func(status int) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// устанавливаю нужный статус в ответ
			res.WriteHeader(status)
		}
	}

	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedClientStorage(ctrl)

	// Тест с успешным добавлением данных--------------------------------------------------------
	userID := "success user id"
	info := dataInfo{
		bank: data.Bank{
			Number: 123456789,
		},
		metaInfo:   "some metainfo",
		name:       "success data name",
		createDate: time.Now(),
		editDate:   time.Now(),
	}
	encrData := repoData.EncryptedData{}
	succesMasterPassword := "succes strong master password"

	m.EXPECT().AddEncryptedData(gomock.Any(), userID, gomock.Any(), repoData.SAVED).Return(true, nil)

	// Тест с возвращением ошибки из хранилища --------------------------------------------------------
	errorID := "error from storage user id"
	errorInfo := dataInfo{
		bank: data.Bank{
			Number: 23151371719419249,
		},
		metaInfo:   "error metainfo",
		name:       "error data name",
		createDate: time.Now(),
		editDate:   time.Now(),
	}
	errorEncrData := repoData.EncryptedData{}

	m.EXPECT().AddEncryptedData(gomock.Any(), errorID, gomock.Any(), repoData.SAVED).Return(false, errors.New("some error"))

	// Тест с попыткой добавить уже существующие данные --------------------------------------------------------
	alreadyExistID := "data is already exist user id"
	alreadyExistInfo := dataInfo{
		bank: data.Bank{
			Number: 458572347238,
		},
		metaInfo:   "data is already exist metainfo",
		name:       "data is already exist data name",
		createDate: time.Now(),
		editDate:   time.Now(),
	}
	alreadyExistData := repoData.EncryptedData{}

	m.EXPECT().AddEncryptedData(gomock.Any(), alreadyExistID, gomock.Any(), repoData.SAVED).Return(false, nil)

	type request struct {
		userID         string
		encrData       repoData.EncryptedData
		info           dataInfo
		startServer    bool
		stor           storage.IEncryptedClientStorage
		httpStatus     int
		masterPassword string
	}
	type want struct {
		ok  bool
		err bool
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success test",
			req: request{
				userID:         userID,
				encrData:       encrData,
				info:           info,
				startServer:    true,
				stor:           m,
				httpStatus:     200,
				masterPassword: succesMasterPassword,
			},
			want: want{
				ok:  true,
				err: false,
			},
		},
		{
			name: "error from storage test",
			req: request{
				userID:         errorID,
				encrData:       errorEncrData,
				info:           errorInfo,
				startServer:    true,
				stor:           m,
				httpStatus:     200,
				masterPassword: succesMasterPassword,
			},
			want: want{
				ok:  false,
				err: true,
			},
		},
		{
			name: "data is already exist",
			req: request{
				userID:         alreadyExistID,
				encrData:       alreadyExistData,
				info:           alreadyExistInfo,
				startServer:    true,
				stor:           m,
				httpStatus:     200,
				masterPassword: succesMasterPassword,
			},
			want: want{
				ok:  false,
				err: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", testHandler(tt.req.httpStatus))

			// Создаю тестовый сервер
			ts := httptest.NewServer(r)
			defer ts.Close()

			var url string
			if tt.req.startServer {
				// Усанвливаю корректный адрес
				url = ts.URL + "/test"
			} else {
				// устанавливаю невалидный url, иммитирую недоступность сервера
				url = "http://wrong.address.com" + "/test"
			}

			ok, err := save(context.Background(), tt.req.userID, url, resty.New(), tt.req.stor, tt.req.info, tt.req.masterPassword)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ok, ok)
			}
		})
	}
}
