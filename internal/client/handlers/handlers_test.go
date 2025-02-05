package handlers

import (
	"context"
	"errors"
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

func TestSaveEncryptedDataToLocalStorage(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedClientStorage(ctrl)

	{
		// Тест с успешным добавлением данных в локальное хранилище
		userID := "success user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("success test data"),
			Name:          "success test data name",
		}
		status := data.NEW
		m.EXPECT().AddEncryptedData(gomock.Any(), userID, encrData, status).Return(true, nil)
		ok, err := SaveEncryptedDataToLocalStorage(context.Background(), userID, m, encrData, status)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
	}
	{
		// Тест с возвращением ошибки из хранилища
		userID := "failed to save data user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("failed to save test data"),
			Name:          "failed to save test data name",
		}
		status := data.NEW
		m.EXPECT().AddEncryptedData(gomock.Any(), userID, encrData, status).Return(false, errors.New("some error"))
		ok, err := SaveEncryptedDataToLocalStorage(context.Background(), userID, m, encrData, status)
		require.Error(t, err)
		assert.Equal(t, false, ok)
	}
	{
		// Тест с попыткой добавить в хранилище уже существующие данные
		userID := "data is already exist user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("data is already exist test data"),
			Name:          "data is already exist test data name",
		}
		status := data.NEW
		m.EXPECT().AddEncryptedData(gomock.Any(), userID, encrData, status).Return(false, nil)
		ok, err := SaveEncryptedDataToLocalStorage(context.Background(), userID, m, encrData, status)
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
}

func TestSaveEncryptedData(t *testing.T) {
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

	// Тест успешной передачи----------------------------------------------
	userID := "success user id"
	encrData := data.EncryptedData{
		EncryptedData: []byte("success ecnrypted dat"),
		Name:          "success data name",
	}
	m.EXPECT().AddEncryptedData(gomock.Any(), userID, encrData, data.SAVED).Return(true, nil)

	// Сервер недоступен, сохранение данных только в локальном хранилище---
	serverNotAvailableUserID := "server not available user id"
	serverNotAvailableEncrData := data.EncryptedData{
		EncryptedData: []byte("server not available ecnrypted dat"),
		Name:          "server not available data name",
	}
	m.EXPECT().AddEncryptedData(gomock.Any(), serverNotAvailableUserID, serverNotAvailableEncrData, data.NEW).Return(true, nil)

	// Внутрення ошибка сервера --------------------------------------------------
	internalServerErrorUserID := "internalServerError user id"
	internalServerErrorEncrData := data.EncryptedData{
		EncryptedData: []byte("internalServerError ecnrypted dat"),
		Name:          "internalServerError data name",
	}
	m.EXPECT().AddEncryptedData(gomock.Any(), internalServerErrorUserID, internalServerErrorEncrData, data.NEW).Return(false, errors.New("some error"))

	type request struct {
		userID      string
		encrData    data.EncryptedData
		startServer bool
		stor        storage.IEncryptedClientStorage
		httpStatus  int
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
				userID:      userID,
				encrData:    encrData,
				startServer: true,
				stor:        m,
				httpStatus:  200,
			},
			want: want{
				ok:  true,
				err: false,
			},
		},
		{
			name: "server not available",
			req: request{
				userID:      serverNotAvailableUserID,
				encrData:    serverNotAvailableEncrData,
				startServer: false,
				stor:        m,
				httpStatus:  200,
			},
			want: want{
				ok:  true,
				err: false,
			},
		},
		{
			name: "internal server error",
			req: request{
				userID:      internalServerErrorUserID,
				encrData:    internalServerErrorEncrData,
				startServer: true,
				stor:        m,
				httpStatus:  500,
			},
			want: want{
				ok:  false,
				err: true,
			},
		},
		{
			name: "uknown server address",
			req: request{
				userID:      userID,
				encrData:    encrData,
				startServer: true,
				stor:        m,
				httpStatus:  403,
			},
			want: want{
				ok:  false,
				err: true,
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

			ok, err := SaveEncryptedData(context.Background(), tt.req.userID, url, resty.New(), tt.req.stor, &tt.req.encrData)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ok, ok)
			}
		})
	}
}
