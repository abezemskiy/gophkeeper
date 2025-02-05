package handlers

import (
	"context"
	"errors"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/common/identity/tools/hasher"
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

			// Запускаю тестовый сервер
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

func TestRegister(t *testing.T) {
	// Вспомогательная функция---------------------------------------
	testHandler := func(status int, token string) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {

			// Если ожидается успешный запрос, то устанавливаю токен в заголовок
			if status == http.StatusOK {
				res.Header().Set("Authorization", "Bearer "+token)
			}
			// устанавливаю нужный статус в ответ
			res.WriteHeader(status)
		}
	}

	// регистрирую мок хранилища хранилища аутентификационных данных клиента
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockClientIdentifier(ctrl)

	// Тест с успешной регистрацией пользователя--------------------------------------------------------------------------
	successAuthData := identity.AuthData{
		Login:    "success login",
		Password: "success password",
	}
	// Вычисляю хэш
	hash, err := hasher.CalkHash(successAuthData.Login + successAuthData.Password)
	require.NoError(t, err)
	// Создаю тестовый токен
	successToken := "success-token"
	// устанавливаю мок
	m.EXPECT().Register(gomock.Any(), successAuthData.Login, hash, gomock.Any(), successToken).Return(true, nil)

	// Тест с возвращением ошибки из хранилища --------------------------------------------------------------------------
	errorAuthData := identity.AuthData{
		Login:    "error login",
		Password: "error password",
	}
	errHash, err := hasher.CalkHash(errorAuthData.Login + errorAuthData.Password)
	require.NoError(t, err)
	errorToken := "error-token"
	m.EXPECT().Register(gomock.Any(), errorAuthData.Login, errHash, gomock.Any(), errorToken).Return(false, errors.New("some error"))

	// Тест попыткой зарегистрировать существующего пользователя ---------------------------------------------------------------------
	alreadyExistAuthData := identity.AuthData{
		Login:    "already exist login",
		Password: "laredy exist password",
	}
	alreadyExistHash, err := hasher.CalkHash(alreadyExistAuthData.Login + alreadyExistAuthData.Password)
	require.NoError(t, err)
	alredyExistToken := "already-exist-token"
	m.EXPECT().Register(gomock.Any(), alreadyExistAuthData.Login, alreadyExistHash, gomock.Any(),
		alredyExistToken).Return(false, nil)

	type request struct {
		wrongURL bool
		authData identity.AuthData
		ident    identity.ClientIdentifier
		status   int
		token    string
	}
	type want struct {
		err bool
		ok  bool
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success test",
			req: request{
				wrongURL: false,
				authData: successAuthData,
				ident:    m,
				status:   200,
				token:    successToken,
			},
			want: want{
				err: false,
				ok:  true,
			},
		},
		{
			name: "invalid login test",
			req: request{
				wrongURL: false,
				authData: identity.AuthData{
					Login:    "",
					Password: "some password",
				},
				ident:  m,
				status: 200,
				token:  successToken,
			},
			want: want{
				err: true,
				ok:  false,
			},
		},
		{
			name: "invalid password test",
			req: request{
				wrongURL: false,
				authData: identity.AuthData{
					Login:    "some login",
					Password: "",
				},
				ident:  m,
				status: 200,
				token:  successToken,
			},
			want: want{
				err: true,
				ok:  false,
			},
		},
		{
			name: "wrong url test",
			req: request{
				wrongURL: true,
				authData: successAuthData,
				ident:    m,
				status:   200,
				token:    successToken,
			},
			want: want{
				err: true,
				ok:  false,
			},
		},
		{
			name: "wrong token test",
			req: request{
				wrongURL: false,
				authData: successAuthData,
				ident:    m,
				status:   200,
				token:    "wrong token because has spaces",
			},
			want: want{
				err: true,
				ok:  false,
			},
		},
		{
			name: "error from auth data storage",
			req: request{
				wrongURL: false,
				authData: errorAuthData,
				ident:    m,
				status:   200,
				token:    errorToken,
			},
			want: want{
				err: true,
				ok:  false,
			},
		},
		{
			name: "user alredy exists",
			req: request{
				wrongURL: false,
				authData: alreadyExistAuthData,
				ident:    m,
				status:   200,
				token:    alredyExistToken,
			},
			want: want{
				err: false,
				ok:  false,
			},
		},
		{
			name: "user alredy exists on server",
			req: request{
				wrongURL: false,
				authData: successAuthData,
				ident:    m,
				status:   409,
				token:    successToken,
			},
			want: want{
				err: false,
				ok:  false,
			},
		},
		{
			name: "bad server status",
			req: request{
				wrongURL: false,
				authData: successAuthData,
				ident:    m,
				status:   500,
				token:    successToken,
			},
			want: want{
				err: true,
				ok:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", testHandler(tt.req.status, tt.req.token))

			// Запускаю тестовый сервер
			ts := httptest.NewServer(r)
			defer ts.Close()

			var url string
			if !tt.req.wrongURL {
				// Усанвливаю корректный адрес
				url = ts.URL + "/test"
			} else {
				// устанавливаю невалидный url, иммитирую недоступность сервера
				url = "http://wrong.address.com" + "/test"
			}

			// Вызываю тестируемый хэндлер
			ok, err := Register(context.Background(), url, &tt.req.authData, resty.New(), tt.req.ident)
			// Сравниваю полученный ответ с ожидаемым
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ok, ok)
			}
		})
	}
}
