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
	idSuccessful := "successful data user id"
	succesfulData := data.EncryptedData{
		EncryptedData: []byte("some encrypted data"),
		Name:          "successfulData",
	}
	successBody, err := json.Marshal(succesfulData)
	require.NoError(t, err)
	m.EXPECT().AddEncryptedData(gomock.Any(), idSuccessful, succesfulData, data.SAVED).Return(true, nil)

	// Тест с возращением ошибки из хранилища
	idError := "error data user id"
	errorData := data.EncryptedData{
		EncryptedData: []byte("error encrypted data"),
		Name:          "error Data",
	}
	errorBody, err := json.Marshal(errorData)
	require.NoError(t, err)
	m.EXPECT().AddEncryptedData(gomock.Any(), idError, errorData, data.SAVED).Return(false, fmt.Errorf("add data error"))

	// Тест с конфликтом данные. Попытка добавить данные, которые уже есть в хранилище.
	idConflict := "conflict data user id"
	conflictData := data.EncryptedData{
		EncryptedData: []byte("conflict encrypted data"),
		Name:          "conflict Data",
	}
	conflictBody, err := json.Marshal(conflictData)
	require.NoError(t, err)
	m.EXPECT().AddEncryptedData(gomock.Any(), idConflict, conflictData, data.SAVED).Return(false, nil)

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
			name: "successful data addition",
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

func TestReplaceEncryptedData(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedServerStorage(ctrl)

	// Тест с успешным редактированием данных в хранилище
	idSuccessful := "successful edit data user id"
	succesfulData := data.EncryptedData{
		EncryptedData: []byte("some encrypted data"),
		Name:          "successfulData",
	}
	successBody, err := json.Marshal(succesfulData)
	require.NoError(t, err)
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), idSuccessful, succesfulData, data.SAVED).Return(true, nil)

	// Тест с попыткой редактирования данных, которых нет в хранилище.
	doesNotExistID := "does not exist user id"
	doesNotExistData := data.EncryptedData{
		EncryptedData: []byte("does not exist data"),
		Name:          "succedoes not exist datafulData",
	}
	doesNotExistBody, err := json.Marshal(doesNotExistData)
	require.NoError(t, err)
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), doesNotExistID, doesNotExistData, data.SAVED).Return(false, nil)

	// Тест с возвратом ошибки из хранилища.
	errorID := "error user id"
	errorData := data.EncryptedData{
		EncryptedData: []byte("error data"),
		Name:          "error data",
	}
	errorBody, err := json.Marshal(errorData)
	require.NoError(t, err)
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), errorID, errorData, data.SAVED).Return(false, fmt.Errorf("some storage error"))

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
			name: "successful data addition",
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
			name: "data does not exist",
			req: request{
				body:  doesNotExistBody,
				stor:  m,
				setID: true,
				id:    doesNotExistID,
			},
			want: want{
				status: 404,
			},
		},
		{
			name: "error from storage",
			req: request{
				body:  errorBody,
				stor:  m,
				setID: true,
				id:    errorID,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "bad data",
			req: request{
				body:  []byte("some bad data"),
				stor:  m,
				setID: true,
				id:    idSuccessful,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "id doesn't set in context",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", func(res http.ResponseWriter, req *http.Request) {
				ReplaceEncryptedData(res, req, tt.req.stor)
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

func TestGetAllEncryptedData(t *testing.T) {
	dataIsEqual := func(want, get [][]data.EncryptedData) bool {
		if len(want) != len(get) {
			return false
		}
		for i := range len(want) {
			if len(want[i]) != len(get[i]) {
				return false
			}
			for j := range want[i] {
				if len(want[i][j].EncryptedData) != len(get[i][j].EncryptedData) {
					return false
				}
				wantStr := string(want[i][j].EncryptedData)
				getStr := string(get[i][j].EncryptedData)
				if wantStr != getStr {
					return false
				}
				if want[i][j].Name != get[i][j].Name {
					return false
				}
			}
		}
		return true
	}
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedServerStorage(ctrl)

	// Тест с успешным получением данных из хранилища
	idSuccessful := "successful user id"
	successData := [][]data.EncryptedData{
		{{Name: "first data", EncryptedData: []byte("first payload")}, {Name: "first data", EncryptedData: []byte("second payload")}},
		{{Name: "second data", EncryptedData: []byte("first payload")}, {Name: "second data", EncryptedData: []byte("second payload")}},
	}
	m.EXPECT().GetAllEncryptedData(gomock.Any(), idSuccessful).Return(successData, nil)

	// Тест с возвращением ошибки из хранилища
	errorID := "error user id"
	m.EXPECT().GetAllEncryptedData(gomock.Any(), errorID).Return(nil, fmt.Errorf("some error"))

	type request struct {
		stor  storage.IEncryptedServerStorage
		setID bool
		id    string
	}
	type want struct {
		status int
		data   [][]data.EncryptedData
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "successful getting data",
			req: request{
				stor:  m,
				setID: true,
				id:    idSuccessful,
			},
			want: want{
				status: 200,
				data:   successData,
			},
		},
		{
			name: "erro from storage",
			req: request{
				stor:  m,
				setID: true,
				id:    errorID,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "id doesn't set in context",
			req: request{
				stor:  m,
				setID: false,
				id:    idSuccessful,
			},
			want: want{
				status: 500,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Get("/test", func(res http.ResponseWriter, req *http.Request) {
				GetAllEncryptedData(res, req, tt.req.stor)
			})

			// создаю тестовый запрос
			request := httptest.NewRequest(http.MethodGet, "/test", nil)
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

			// Проверяю тело запроса с ожидаемым
			if tt.want.status == http.StatusOK {
				var getData [][]data.EncryptedData
				dec := json.NewDecoder(res.Body)
				err := dec.Decode(&getData)
				require.NoError(t, err)

				// проверяю данные отправленные сервером
				assert.Equal(t, true, dataIsEqual(tt.want.data, getData))
			}
		})
	}
}

func TestDeleteEncryptedData(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedServerStorage(ctrl)

	// Тест с успешным удалением данных из хранилища
	idSuccessful := "successful delete data user id"
	successfulDataName := "successful delete data name"
	successfulBody, err := json.Marshal(data.MetaInfo{
		Name: successfulDataName,
	})
	require.NoError(t, err)
	m.EXPECT().DeleteEncryptedData(gomock.Any(), idSuccessful, successfulDataName).Return(true, nil)

	// Тест с возвращением ошибки из хранилища
	errorID := "error delete data user id"
	errorDataName := "error delete data name"
	errorBody, err := json.Marshal(data.MetaInfo{
		Name: errorDataName,
	})
	require.NoError(t, err)
	m.EXPECT().DeleteEncryptedData(gomock.Any(), errorID, errorDataName).Return(false, errors.New("some storage error"))

	// Тест с попыткой удалить несуществующие данные
	doesNotID := "does not exist data user id"
	doesNotDataName := "does not exist delete data name"
	doesNotBody, err := json.Marshal(data.MetaInfo{
		Name: doesNotDataName,
	})
	require.NoError(t, err)
	m.EXPECT().DeleteEncryptedData(gomock.Any(), doesNotID, doesNotDataName).Return(false, nil)

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
			name: "successful data deletion",
			req: request{
				body:  successfulBody,
				stor:  m,
				setID: true,
				id:    idSuccessful,
			},
			want: want{
				status: 200,
			},
		},
		{
			name: "bad meta info",
			req: request{
				body:  []byte("bad meta info"),
				stor:  m,
				setID: true,
				id:    idSuccessful,
			},
			want: want{
				status: 400,
			},
		},
		{
			name: "error from storage",
			req: request{
				body:  errorBody,
				stor:  m,
				setID: true,
				id:    errorID,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "data doesn't exist",
			req: request{
				body:  doesNotBody,
				stor:  m,
				setID: true,
				id:    doesNotID,
			},
			want: want{
				status: 404,
			},
		},
		{
			name: "id does not set in context",
			req: request{
				body:  successfulBody,
				stor:  m,
				setID: false,
				id:    idSuccessful,
			},
			want: want{
				status: 500,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Delete("/test", func(res http.ResponseWriter, req *http.Request) {
				DeleteEncryptedData(res, req, tt.req.stor)
			})

			// создаю тестовый запрос
			request := httptest.NewRequest(http.MethodDelete, "/test", bytes.NewBuffer(tt.req.body))
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

func TestHandleConflictData(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedServerStorage(ctrl)

	// Тест с успешным добавление новой версии данных в хранилище
	idSuccessful := "successful append data user id"
	succesfulData := data.EncryptedData{
		EncryptedData: []byte("some encrypted data"),
		Name:          "successfulData",
	}
	successBody, err := json.Marshal(succesfulData)
	require.NoError(t, err)
	m.EXPECT().AppendEncryptedData(gomock.Any(), idSuccessful, succesfulData).Return(true, nil)

	// Тест с возвращением ошибки из хранилища
	errorID := "error from storage while append data user id"
	errorData := data.EncryptedData{
		EncryptedData: []byte("some error encrypted data"),
		Name:          "error data name",
	}
	errorBody, err := json.Marshal(errorData)
	require.NoError(t, err)
	m.EXPECT().AppendEncryptedData(gomock.Any(), errorID, errorData).Return(false, errors.New("some storage error"))

	// Тест с попыткой изменить данные, которых нет в хранилище
	doesNotExistID := "does not exist data in append handler user id"
	doesNotExistData := data.EncryptedData{
		EncryptedData: []byte("some does not exist data"),
		Name:          "does not exist data name",
	}
	doesNotExistBody, err := json.Marshal(doesNotExistData)
	require.NoError(t, err)
	m.EXPECT().AppendEncryptedData(gomock.Any(), doesNotExistID, doesNotExistData).Return(false, nil)

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
			name: "successful data appending",
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
			name: "bad data",
			req: request{
				body:  []byte("some bad data"),
				stor:  m,
				setID: true,
				id:    idSuccessful,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "error in storage",
			req: request{
				body:  errorBody,
				stor:  m,
				setID: true,
				id:    errorID,
			},
			want: want{
				status: 500,
			},
		},
		{
			name: "does not exist data",
			req: request{
				body:  doesNotExistBody,
				stor:  m,
				setID: true,
				id:    doesNotExistID,
			},
			want: want{
				status: 404,
			},
		},
		{
			name: "id does not set in context",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Post("/test", func(res http.ResponseWriter, req *http.Request) {
				HandleConflictData(res, req, tt.req.stor)
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

func TestHandleOtherRequest(t *testing.T) {
	{
		r := chi.NewRouter()
		r.Post("/test", HandleOtherRequest())
		// создаю тестовый запрос
		request := httptest.NewRequest(http.MethodPost, "/test", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, request)

		res := w.Result()
		defer res.Body.Close() // закрываю тело ответа
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	}
	{
		r := chi.NewRouter()
		r.Get("/test", HandleOtherRequest())
		// создаю тестовый запрос
		request := httptest.NewRequest(http.MethodGet, "/test", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, request)

		res := w.Result()
		defer res.Body.Close() // закрываю тело ответа
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	}
}
