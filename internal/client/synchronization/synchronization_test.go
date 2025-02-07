package synchronization

import (
	"context"
	"encoding/json"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/repositories/data"
	"gophkeeper/internal/repositories/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSynchronizeNewLocalData(t *testing.T) {
	// Хэндлер для тестовой обработки запроса клиента на авторизацию на сервере
	testHandler := func(status int) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// Извлекаю данные из запроса клиента
			var encrData data.EncryptedData
			err := json.NewDecoder(req.Body).Decode(&encrData)
			require.NoError(t, err)

			// Проверяю корректность полученных данных
			assert.NotEqual(t, "", encrData.Name)
			assert.NotEqual(t, 0, len(encrData.EncryptedData))

			// устанавливаю нужный статус в ответ
			res.WriteHeader(status)
		}
	}

	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stor := mocks.NewMockIEncryptedClientStorage(ctrl)

	// Тест с успешным сохранением локальных данных со статусом NEW на сервере
	successID := "success id"
	successInfo := mocks.NewMockIUserInfoStorage(ctrl)
	successInfo.EXPECT().Get().Return(identity.AuthData{}, successID)
	wantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data"), Name: "first encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
		{{EncryptedData: []byte("third encr data"), Name: "third encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), successID, data.NEW).Return(wantData, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), successID, "first encr data name", data.SAVED).Return(true, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), successID, "second encr data name", data.SAVED).Return(true, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), successID, "third encr data name", data.SAVED).Return(true, nil)

	type request struct {
		stor        storage.IEncryptedClientStorage
		info        identity.IUserInfoStorage
		setValidURL bool
		status      int
	}
	type want struct {
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
				stor:        stor,
				info:        successInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", testHandler(tt.req.status))

			// Запускаю тестовый сервер
			ts := httptest.NewServer(r)
			defer ts.Close()

			var url string
			if tt.req.setValidURL {
				// Усанвливаю корректный адрес
				url = ts.URL + "/test"
			} else {
				// устанавливаю невалидный url, иммитирую недоступность сервера
				url = "http://wrong.address.com" + "/test"
			}

			err := SynchronizeNewLocalData(context.Background(), tt.req.stor, tt.req.info, resty.New(), url)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
