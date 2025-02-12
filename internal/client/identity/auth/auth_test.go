package auth

import (
	"encoding/json"
	"errors"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/common/identity/tools/hasher"
	"gophkeeper/internal/common/identity/tools/header"
	repoIdent "gophkeeper/internal/repositories/identity"
	"gophkeeper/internal/repositories/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnBeforeMiddleware(t *testing.T) {
	// вспомогательная функция
	testHandler := func(token string) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// Проверяю токен, установленный в заголовок
			getToken, err := header.GetTokenFromHeader(req)
			require.NoError(t, err)
			assert.Equal(t, token, getToken)

			// устанавливаю нужный статус в ответ
			res.WriteHeader(200)
		}
	}

	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// мок хранилища идентификационных данных пользователей
	ident := mocks.NewMockClientIdentifier(ctrl)

	// Тест с успешной установкой заголовка ---------------------------------------------------
	successInfo := mocks.NewMockIUserInfoStorage(ctrl)
	successLogin := "success login"
	successToken := "success-token"
	successInfo.EXPECT().Get().Return(identity.AuthData{
		Login: successLogin,
	}, "some id")
	ident.EXPECT().Authorize(gomock.Any(), successLogin).Return(identity.UserInfo{
		Token: successToken,
	}, true, nil)

	// Тест с возвращением ошибки из хранилища------------------------------------------------------
	errorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	errorLogin := "error login"
	errorInfo.EXPECT().Get().Return(identity.AuthData{
		Login: errorLogin,
	}, "some id")
	ident.EXPECT().Authorize(gomock.Any(), errorLogin).Return(identity.UserInfo{}, false, errors.New("some error"))

	// Тест с внезарегистрированным пользователем ------------------------------------------------------
	notRegisterInfo := mocks.NewMockIUserInfoStorage(ctrl)
	notRegisterLogin := "not register login"
	notRegisterInfo.EXPECT().Get().Return(identity.AuthData{
		Login: notRegisterLogin,
	}, "some id")
	ident.EXPECT().Authorize(gomock.Any(), notRegisterLogin).Return(identity.UserInfo{}, false, nil)

	type request struct {
		info  identity.IUserInfoStorage
		ident identity.ClientIdentifier
	}
	type want struct {
		err   bool
		token string
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success request",
			req: request{
				info:  successInfo,
				ident: ident,
			},
			want: want{
				err:   false,
				token: successToken,
			},
		},
		{
			name: "error from auth data storage",
			req: request{
				info:  errorInfo,
				ident: ident,
			},
			want: want{
				err:   true,
				token: "some token",
			},
		},
		{
			name: "not register user",
			req: request{
				info:  notRegisterInfo,
				ident: ident,
			},
			want: want{
				err:   true,
				token: "some token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/test", testHandler(tt.want.token))

			// Запускаю тестовый сервер
			ts := httptest.NewServer(r)
			defer ts.Close()

			// создаю корректный url
			url := ts.URL + "/test"

			// Создаю новый resty клиент
			client := resty.New()

			// Устанавливаю мидлварь на клиента
			client.OnBeforeRequest(OnBeforeMiddleware(tt.req.info, tt.req.ident))

			// Выполняю запрос к серверу
			resp, err := client.R().
				Get(url)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode())
			}
		})
	}
}

func TestOnAfterMiddleware(t *testing.T) {
	// Хэндлер для имитации ответа сервера, что клиент не авторизирован
	testHandler := func() http.HandlerFunc {
		return func(res http.ResponseWriter, _ *http.Request) {
			// устанавливаю нужный статус в ответ
			res.WriteHeader(http.StatusUnauthorized)
		}
	}

	// Хэндлер для тестовой обработки запроса клиента на авторизацию на сервере
	testHandlerAuth := func(status int, login, hash, token string) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			if status == http.StatusOK {
				// Извлекаю логин и хэш из запроса пользователя на авторизацию
				var regData repoIdent.Data
				err := json.NewDecoder(req.Body).Decode(&regData)
				require.NoError(t, err)

				// Сравниваю полученные логин и хэш с ожидаемыми
				assert.Equal(t, login, regData.Login)
				assert.Equal(t, hash, regData.Hash)

				// Устанавливаю новый токен в заголовок ответа
				res.Header().Set("Authorization", "Bearer "+token)
			}
			// устанавливаю нужный статус в ответ
			res.WriteHeader(status)
		}
	}

	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// мок хранилища идентификационных данных пользователей
	ident := mocks.NewMockClientIdentifier(ctrl)

	// Тест с успешной установкой заголовка ---------------------------------------------------
	successInfo := mocks.NewMockIUserInfoStorage(ctrl)
	successLogin := "success login"
	successPassword := "success password"
	hash, err := hasher.CalkHash(successLogin + successPassword)
	require.NoError(t, err)

	successToken := "success-token"
	successInfo.EXPECT().Get().Return(identity.AuthData{
		Login:    successLogin,
		Password: successPassword,
	}, "some id")
	ident.EXPECT().SetToken(gomock.Any(), successLogin, successToken).Return(true, nil)

	// Тест с неправильным адресом к хэндлеру аутентификации на сервере ---------------------------------------------------
	wrongAuthURLInfo := mocks.NewMockIUserInfoStorage(ctrl)
	wrongAuthURLInfo.EXPECT().Get().Return(identity.AuthData{
		Login:    "wrongAuthUR login",
		Password: "wrongAuthUR password",
	}, "some id")

	// Тест, когда сервер возвращает статус 500 при попытке авторизации ---------------------------------------------------
	status500Info := mocks.NewMockIUserInfoStorage(ctrl)
	status500Info.EXPECT().Get().Return(identity.AuthData{
		Login:    "status500 login",
		Password: "status500 password",
	}, "some id")

	// Тест с возвращением ошибки из хранилища аутентификационных данных  ---------------------------------------------------
	errorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	errorLogin := "error login"
	errorPassword := "error password"
	errorHash, err := hasher.CalkHash(errorLogin + errorPassword)
	require.NoError(t, err)

	errorToken := "error-token"
	errorInfo.EXPECT().Get().Return(identity.AuthData{
		Login:    errorLogin,
		Password: errorPassword,
	}, "some id")
	ident.EXPECT().SetToken(gomock.Any(), errorLogin, errorToken).Return(true, errors.New("some error"))

	// Тест попыткой обновления данных неавторизированного полльзователя  ---------------------------------------------------
	notRegisterInfo := mocks.NewMockIUserInfoStorage(ctrl)
	notRegisterLogin := "notRegister login"
	notRegisterPassword := "notRegister password"
	notRegisterHash, err := hasher.CalkHash(notRegisterLogin + notRegisterPassword)
	require.NoError(t, err)

	notRegisterToken := "notRegister-token"
	notRegisterInfo.EXPECT().Get().Return(identity.AuthData{
		Login:    notRegisterLogin,
		Password: notRegisterPassword,
	}, "some id")
	ident.EXPECT().SetToken(gomock.Any(), notRegisterLogin, notRegisterToken).Return(false, nil)

	type request struct {
		info              identity.IUserInfoStorage
		ident             identity.ClientIdentifier
		correctURLForAuth bool
	}
	type want struct {
		err    bool
		token  string
		status int
		login  string
		hash   string
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success request",
			req: request{
				info:              successInfo,
				ident:             ident,
				correctURLForAuth: true,
			},
			want: want{
				err:    false,
				token:  successToken,
				status: 200,
				login:  successLogin,
				hash:   hash,
			},
		},
		{
			name: "wrong auth url",
			req: request{
				info:              wrongAuthURLInfo,
				ident:             ident,
				correctURLForAuth: false,
			},
			want: want{
				err:    true,
				token:  successToken,
				status: 200,
				login:  successLogin,
				hash:   hash,
			},
		},
		{
			name: "status 500 from server in auth request",
			req: request{
				info:              status500Info,
				ident:             ident,
				correctURLForAuth: true,
			},
			want: want{
				err:    false,
				token:  successToken,
				status: 500,
				login:  successLogin,
				hash:   hash,
			},
		},
		{
			name: "error from storage",
			req: request{
				info:              errorInfo,
				ident:             ident,
				correctURLForAuth: true,
			},
			want: want{
				err:    true,
				token:  errorToken,
				status: 200,
				login:  errorLogin,
				hash:   errorHash,
			},
		},
		{
			name: "not register user",
			req: request{
				info:              notRegisterInfo,
				ident:             ident,
				correctURLForAuth: true,
			},
			want: want{
				err:    true,
				token:  notRegisterToken,
				status: 200,
				login:  notRegisterLogin,
				hash:   notRegisterHash,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/test", testHandler())
			r.Post("/auth", testHandlerAuth(tt.want.status, tt.want.login, tt.want.hash, tt.want.token))

			// Запускаю тестовый сервер
			ts := httptest.NewServer(r)
			defer ts.Close()

			// создаю корректный url
			url := ts.URL + "/test"

			// Создаю url для доступа к аутентификации на сервере
			var authURL string
			if tt.req.correctURLForAuth {
				authURL = ts.URL + "/auth"
			} else {
				authURL = "http://worg.server.address"
			}

			// Создаю новый resty клиент
			client := resty.New()

			// Устанавливаю мидлварь на клиента
			client.OnAfterResponse(OnAfterMiddleware(tt.req.info, tt.req.ident, authURL))

			// Выполняю запрос к серверу
			resp, err := client.R().
				Get(url)

			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Если при попытке авторизации на сервере от сервера ожидается статус иной чем 200, то ожидаю,
				// что орининальный ответ сервера будет заменен на ответ сервера после попытки авторизации.
				if tt.want.status != http.StatusOK {
					assert.Equal(t, tt.want.status, resp.StatusCode())
				} else {
					assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())
				}

			}
		})
	}
}
