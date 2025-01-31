package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gophkeeper/internal/common/identity/tools/header"
	"gophkeeper/internal/common/identity/tools/token"
	"gophkeeper/internal/repositories/data"
	"gophkeeper/internal/repositories/identity"
	"gophkeeper/internal/repositories/mocks"
	"gophkeeper/internal/server/identity/auth"
	"gophkeeper/internal/server/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIdentifier(ctrl)

	// Test. success register---------------------------------------------------------
	testHash := "success hash"
	regData := identity.IdentityData{
		Login: "success login",
		Hash:  testHash,
	}
	successBody, err := json.Marshal(regData)
	require.NoError(t, err)
	require.NoError(t, err)
	m.EXPECT().Register(gomock.Any(), regData.Login, regData.Hash, gomock.Any()).Return(nil)

	// Test. user already register------------------------------------------------------------
	alreadyData := identity.IdentityData{
		Login: "already login",
		Hash:  "already hash",
	}
	alreadyBody, err := json.Marshal(alreadyData)
	require.NoError(t, err)
	require.NoError(t, err)
	m.EXPECT().Register(gomock.Any(), alreadyData.Login, alreadyData.Hash, gomock.Any()).Return(&pgconn.PgError{Code: "23505"})

	// Test. register error (internal server error) ------------------------------------------------------------
	internalData := identity.IdentityData{
		Login: "internal login",
		Hash:  "internal hash",
	}
	internalBody, err := json.Marshal(internalData)
	require.NoError(t, err)
	require.NoError(t, err)
	m.EXPECT().Register(gomock.Any(), internalData.Login, internalData.Hash, gomock.Any()).Return(errors.New("some error"))

	// Test. bad login ------------------------------------------------------------------------------------------
	badloginData := identity.IdentityData{
		Login: "",
		Hash:  "hash",
	}
	badloginBody, err := json.Marshal(badloginData)
	require.NoError(t, err)

	// Test. bad hash ------------------------------------------------------------------------------------------
	badpasswordData := identity.IdentityData{
		Login: "",
		Hash:  "hash",
	}
	badpasswordBody, err := json.Marshal(badpasswordData)
	require.NoError(t, err)

	type request struct {
		body []byte
		stor identity.Identifier
	}
	type want struct {
		status int
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success register",
			req: request{
				body: successBody,
				stor: m,
			},
			want: want{
				status: 200,
			},
		},
		{
			name: "user already register",
			req: request{
				body: alreadyBody,
				stor: m,
			},
			want: want{
				status: 409,
			},
		},
		{
			name: "internal server server while register",
			req: request{
				body: internalBody,
				stor: m,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "bad body",
			req: request{
				body: []byte("bad body"),
				stor: nil,
			},
			want: want{
				status: 400,
			},
		},
		{
			name: "bad login",
			req: request{
				body: badloginBody,
				stor: nil,
			},
			want: want{
				status: 400,
			},
		},
		{
			name: "bad hash",
			req: request{
				body: badpasswordBody,
				stor: nil,
			},
			want: want{
				status: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// устанавливаю секретный ключ для подписи токена
			token.SetSecretKey("test key")
			// устанавливаю время жизни токена
			token.SerExpireHour(1)

			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", func(res http.ResponseWriter, req *http.Request) {
				Register(res, req, tt.req.stor)
			})

			// создаю тестовый запрос
			request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(tt.req.body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close() // закрываю тело ответа
			assert.Equal(t, tt.want.status, res.StatusCode)

			// если ожидается успешная регистрация, то проверяю корректность JWT в заголовке
			if tt.want.status == 200 {
				getToken, err := header.GetTokenFromResponseHeader(res)
				require.NoError(t, err)
				getId, err := token.GetIDFromToken(getToken)
				require.NoError(t, err)
				assert.NotEqual(t, "", getId)
			}
		})
	}
}

func TestAuthorize(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIdentifier(ctrl)

	// Test. success authorization ---------------------------------------------------------
	testHash := "success hash"
	authData := identity.IdentityData{
		Login: "success login",
		Hash:  testHash,
	}
	successBody, err := json.Marshal(authData)
	require.NoError(t, err)

	// тестовые данные
	wantID := "2362362"
	wantData := identity.AuthorizationData{
		Hash: testHash,
		ID:   wantID,
	}
	m.EXPECT().Authorize(gomock.Any(), authData.Login).Return(wantData, true, nil)

	// Test. authorization error, user not register ---------------------------------------------------------
	notRegisterData := identity.IdentityData{
		Login: "not register login",
		Hash:  "not register hash",
	}
	notRegisterBody, err := json.Marshal(notRegisterData)
	require.NoError(t, err)
	m.EXPECT().Authorize(gomock.Any(), notRegisterData.Login).Return(identity.AuthorizationData{}, false, nil)

	// Test. login is invalid ---------------------------------------------------------
	invalidLoginData := identity.IdentityData{
		Login: "",
		Hash:  "hash",
	}
	invalidLoginBody, err := json.Marshal(invalidLoginData)
	require.NoError(t, err)

	// Test. hash is invalid ---------------------------------------------------------
	invalidPasswordData := identity.IdentityData{
		Login: "login",
		Hash:  "",
	}
	invalidPasswordBody, err := json.Marshal(invalidPasswordData)
	require.NoError(t, err)

	// Test. authorization error, get auth data form storage error ---------------------------------------------------------
	errorData := identity.IdentityData{
		Login: "error login",
		Hash:  "error hash",
	}
	errorBody, err := json.Marshal(errorData)
	require.NoError(t, err)
	m.EXPECT().Authorize(gomock.Any(), errorData.Login).Return(identity.AuthorizationData{}, false, errors.New("get data error"))

	// Test. wrong hash ----------------------------------------------------------------------------------------------
	wrongPaswordData := identity.IdentityData{
		Login: "wrong password test login",
		Hash:  "wrong hash test hash",
	}
	wrongPaswordBody, err := json.Marshal(wrongPaswordData)
	require.NoError(t, err)

	// тестовые данные
	wrongPaswordWantData := identity.AuthorizationData{
		Hash: "want hash",
		ID:   "",
	}
	m.EXPECT().Authorize(gomock.Any(), wrongPaswordData.Login).Return(wrongPaswordWantData, true, nil)

	// Test. wrong login ----------------------------------------------------------------------------------------------
	wrongHash := "wrong hash test login"
	wrongHashData := identity.IdentityData{
		Login: wrongHash,
		Hash:  "wrong hash test hash",
	}
	wrongHashBody, err := json.Marshal(wrongHashData)
	require.NoError(t, err)

	// тестовые данные
	wrongHashWantData := identity.AuthorizationData{
		Hash: wrongHash,
		ID:   "",
	}
	m.EXPECT().Authorize(gomock.Any(), wrongHashData.Login).Return(wrongHashWantData, true, nil)

	type request struct {
		body []byte
		stor identity.Identifier
	}
	type want struct {
		id     string
		status int
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success authorization",
			req: request{
				body: successBody,
				stor: m,
			},
			want: want{
				id:     wantID,
				status: 200,
			},
		},
		{
			name: "authorization error, user not register",
			req: request{
				body: notRegisterBody,
				stor: m,
			},
			want: want{
				id:     "",
				status: 400,
			},
		},
		{
			name: "login is invalid",
			req: request{
				body: invalidLoginBody,
				stor: m,
			},
			want: want{
				id:     "",
				status: 400,
			},
		},
		{
			name: "password is invalid",
			req: request{
				body: invalidPasswordBody,
				stor: m,
			},
			want: want{
				id:     "",
				status: 400,
			},
		},
		{
			name: "get auth data form storage error",
			req: request{
				body: errorBody,
				stor: m,
			},
			want: want{
				id:     "",
				status: 500,
			},
		},
		{
			name: "wrong password",
			req: request{
				body: wrongPaswordBody,
				stor: m,
			},
			want: want{
				id:     "",
				status: 400,
			},
		},
		{
			name: "wrong hash",
			req: request{
				body: wrongHashBody,
				stor: m,
			},
			want: want{
				id:     "",
				status: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// устанавливаю секретный ключ для подписи токена
			token.SetSecretKey("test key")
			// устанавливаю время жизни токена
			token.SerExpireHour(1)

			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", func(res http.ResponseWriter, req *http.Request) {
				Authorize(res, req, tt.req.stor)
			})

			// создаю тестовый запрос
			request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(tt.req.body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close() // закрываю тело ответа
			assert.Equal(t, tt.want.status, res.StatusCode)

			// если ожидается успешная регистрация, то проверяю корректность JWT в заголовке
			if tt.want.status == 200 {
				getToken, err := header.GetTokenFromResponseHeader(res)
				require.NoError(t, err)
				getId, err := token.GetIDFromToken(getToken)
				require.NoError(t, err)
				assert.Equal(t, tt.want.id, getId)
			}
		})
	}
}

func TestAddEncryptedData(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedServerStorage(ctrl)

	// Тест с успешным добавлением данных в хранилище
	idSuccessful := "succesful data user id"
	succesfulData := data.EncryptedData{
		EncryptedData: []byte("some encrypted data"),
		Name:          "succesfulData",
	}
	successBody, err := json.Marshal(succesfulData)
	require.NoError(t, err)
	m.EXPECT().AddEncryptedData(gomock.Any(), idSuccessful, succesfulData).Return(true, nil)

	// Тест с возращением ошибки из хранилища
	idError := "error data user id"
	errorData := data.EncryptedData{
		EncryptedData: []byte("error encrypted data"),
		Name:          "error Data",
	}
	errorBody, err := json.Marshal(errorData)
	require.NoError(t, err)
	m.EXPECT().AddEncryptedData(gomock.Any(), idError, errorData).Return(false, fmt.Errorf("add data error"))

	// Тест с конфликтом данные. Попытка добавить данные, которые уже есть в хранилище.
	idConflict := "conflict data user id"
	conflictData := data.EncryptedData{
		EncryptedData: []byte("conflict encrypted data"),
		Name:          "conflict Data",
	}
	conflictBody, err := json.Marshal(conflictData)
	require.NoError(t, err)
	m.EXPECT().AddEncryptedData(gomock.Any(), idConflict, conflictData).Return(false, nil)

	type request struct {
		body  []byte
		stor  storage.IEncryptedServerStorage
		setID bool
		id    string
	}
	type want struct {
		status int
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "succesful data addition",
			req: request{
				body:  successBody,
				stor:  m,
				setID: true,
				id:    idSuccessful,
			},
			want: want{
				status: 200,
			},
		},
		{
			name: "don't set context",
			req: request{
				body:  successBody,
				stor:  m,
				setID: false,
				id:    idSuccessful,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "bad data",
			req: request{
				body:  []byte("bad data"),
				stor:  m,
				setID: true,
				id:    idSuccessful,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "error from storage",
			req: request{
				body:  errorBody,
				stor:  m,
				setID: true,
				id:    idError,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "data already exist",
			req: request{
				body:  conflictBody,
				stor:  m,
				setID: true,
				id:    idConflict,
			},
			want: want{
				status: 409,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", func(res http.ResponseWriter, req *http.Request) {
				AddEncryptedData(res, req, tt.req.stor)
			})

			// создаю тестовый запрос
			request := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(tt.req.body))
			if tt.req.setID {
				// устанавливаю id пользователя в контекст
				ctx := context.WithValue(request.Context(), auth.UserIDKey, tt.req.id)
				request = request.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close() // закрываю тело ответа
			assert.Equal(t, tt.want.status, res.StatusCode)
		})
	}
}
