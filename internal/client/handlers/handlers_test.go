package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abezemskiy/gophkeeper/internal/client/identity"
	"github.com/abezemskiy/gophkeeper/internal/client/storage"
	"github.com/abezemskiy/gophkeeper/internal/common/identity/tools/hasher"
	"github.com/abezemskiy/gophkeeper/internal/repositories/data"
	"github.com/abezemskiy/gophkeeper/internal/repositories/mocks"

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
		return func(res http.ResponseWriter, _ *http.Request) {
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
		return func(res http.ResponseWriter, _ *http.Request) {

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

func TestAuthorize(t *testing.T) {
	// регистрирую мок хранилища идентификационных данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ident := mocks.NewMockClientIdentifier(ctrl)

	// регистрирую мок хранилища информации о пользователе
	info := mocks.NewMockIUserInfoStorage(ctrl)

	// Успешная авторизация --------------------------------------------------------------------
	successAuthData := identity.AuthData{
		Login:    "success login",
		Password: "success password",
	}
	successHash, err := hasher.CalkHash(successAuthData.Login + successAuthData.Password)
	require.NoError(t, err)
	successID := "success id"
	ident.EXPECT().Authorize(gomock.Any(), successAuthData.Login).Return(identity.UserInfo{
		ID:    successID,
		Token: "success token",
		Hash:  successHash,
	}, true, nil)
	info.EXPECT().Set(successAuthData, successID)

	// Возвращение ошибки из хранилища аутентификационных данных --------------------------------------------------------------------
	errorAuthData := identity.AuthData{
		Login:    "error login",
		Password: "error password",
	}
	require.NoError(t, err)
	ident.EXPECT().Authorize(gomock.Any(), errorAuthData.Login).Return(identity.UserInfo{}, true, errors.New("some error"))

	// Пользователь не зарегистрирован --------------------------------------------------------------------
	notRegisterAuthData := identity.AuthData{
		Login:    "not register login",
		Password: "not register password",
	}
	require.NoError(t, err)
	ident.EXPECT().Authorize(gomock.Any(), notRegisterAuthData.Login).Return(identity.UserInfo{}, false, nil)

	// Неверный пароль --------------------------------------------------------------------
	wrongPassAuthData := identity.AuthData{
		Login:    "wrong password login",
		Password: "wrong password password",
	}
	ident.EXPECT().Authorize(gomock.Any(), wrongPassAuthData.Login).Return(identity.UserInfo{
		ID:    "wrong pass id",
		Token: "success token",
		Hash:  "wrong hash",
	}, true, nil)

	type request struct {
		authData *identity.AuthData
		ident    identity.ClientIdentifier
		info     identity.IUserInfoStorage
	}
	type want struct {
		err           bool
		registered    bool
		passIsCorrect bool
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success authorize",
			req: request{
				authData: &successAuthData,
				ident:    ident,
				info:     info,
			},
			want: want{
				err:           false,
				registered:    true,
				passIsCorrect: true,
			},
		},
		{
			name: "bad login",
			req: request{
				authData: &identity.AuthData{
					Login:    "",
					Password: "some password",
				},
				ident: ident,
				info:  info,
			},
			want: want{
				err:           true,
				registered:    false,
				passIsCorrect: false,
			},
		},
		{
			name: "bad password",
			req: request{
				authData: &identity.AuthData{
					Login:    "some login",
					Password: "",
				},
				ident: ident,
				info:  info,
			},
			want: want{
				err:           true,
				registered:    false,
				passIsCorrect: false,
			},
		},
		{
			name: "error from auth data storage",
			req: request{
				authData: &notRegisterAuthData,
				ident:    ident,
				info:     nil,
			},
			want: want{
				err:           false,
				registered:    false,
				passIsCorrect: false,
			},
		},
		{
			name: "user not register",
			req: request{
				authData: &errorAuthData,
				ident:    ident,
				info:     nil,
			},
			want: want{
				err:           true,
				registered:    false,
				passIsCorrect: false,
			},
		},
		{
			name: "wrong password",
			req: request{
				authData: &wrongPassAuthData,
				ident:    ident,
				info:     nil,
			},
			want: want{
				err:           false,
				registered:    true,
				passIsCorrect: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passIsCorrect, registered, err := Authorize(context.Background(), tt.req.authData, tt.req.ident, tt.req.info)

			if tt.want.err {
				require.Error(t, err)
				return
			}
			if !tt.want.registered {
				assert.Equal(t, false, registered)
				return
			}
			assert.Equal(t, tt.want.passIsCorrect, passIsCorrect)
		})
	}
}

func TestDeleteEncryptedDataFromLocalStorage(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedClientStorage(ctrl)

	{
		// Тест с успешным удалением данных из локального хранилища
		userID := "success user id"
		dataName := "success data name"
		m.EXPECT().DeleteEncryptedData(gomock.Any(), userID, dataName).Return(true, nil)

		ok, err := DeleteEncryptedDataFromLocalStorage(context.Background(), userID, dataName, m)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
	}
	{
		// Данных не существует
		userID := "data not exists user id"
		dataName := "data not exists data name"
		m.EXPECT().DeleteEncryptedData(gomock.Any(), userID, dataName).Return(false, nil)

		ok, err := DeleteEncryptedDataFromLocalStorage(context.Background(), userID, dataName, m)
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
	{
		// Ошибка из хранилища
		userID := "error user id"
		dataName := "error data name"
		m.EXPECT().DeleteEncryptedData(gomock.Any(), userID, dataName).Return(false, errors.New("some error"))

		_, err := DeleteEncryptedDataFromLocalStorage(context.Background(), userID, dataName, m)
		require.Error(t, err)
	}
}

func TestDeleteEncryptedData(t *testing.T) {
	// вспомогательная функция
	testHandler := func(status int, wantDataName string) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// устанавливаю нужный статус в ответ
			res.WriteHeader(status)

			// Проверяю имя данных в запросе к серверу
			if status == http.StatusOK {
				var dataMetaInfo data.MetaInfo
				dec := json.NewDecoder(req.Body)
				err := dec.Decode(&dataMetaInfo)
				require.NoError(t, err)
				assert.Equal(t, wantDataName, dataMetaInfo.Name)
			}
		}
	}

	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedClientStorage(ctrl)

	// Тест успешного удаления, статус 200 от сервера ----------------------------------------------
	userID := "success user id"
	dataName := "success data name"
	m.EXPECT().DeleteEncryptedData(gomock.Any(), userID, dataName).Return(true, nil)

	// Тест успешного удаления, статус 404 от сервера ----------------------------------------------
	notExistsUserID := "not exists user id"
	notExistsDataName := "not exists data name"
	m.EXPECT().DeleteEncryptedData(gomock.Any(), notExistsUserID, notExistsDataName).Return(true, nil)

	type request struct {
		userID      string
		dataName    string
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
				dataName:    dataName,
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
			name: "data does not exists on server",
			req: request{
				userID:      notExistsUserID,
				dataName:    notExistsDataName,
				startServer: true,
				stor:        m,
				httpStatus:  404,
			},
			want: want{
				ok:  true,
				err: false,
			},
		},
		{
			name: "error, bad status from server",
			req: request{
				userID:      "bad status",
				dataName:    "bad status",
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
			name: "not connection",
			req: request{
				userID:      "user id",
				dataName:    "data name",
				startServer: false,
				stor:        m,
				httpStatus:  200,
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
			r.Delete("/test", testHandler(tt.req.httpStatus, tt.req.dataName))

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

			ok, err := DeleteEncryptedData(context.Background(), tt.req.userID, url, tt.req.dataName, resty.New(), tt.req.stor)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ok, ok)
			}
		})
	}
}

func TestReplaceEncryptedDataToLocalStorage(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedClientStorage(ctrl)

	{
		// Тест с успешной заменой данных в локальном хранилище
		userID := "success user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("success test data"),
			Name:          "success test data name",
		}
		status := data.NEW
		m.EXPECT().ReplaceEncryptedData(gomock.Any(), userID, encrData, status).Return(true, nil)

		ok, err := ReplaceEncryptedDataToLocalStorage(context.Background(), userID, m, encrData, status)
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
		m.EXPECT().ReplaceEncryptedData(gomock.Any(), userID, encrData, status).Return(false, errors.New("some error"))
		ok, err := ReplaceEncryptedDataToLocalStorage(context.Background(), userID, m, encrData, status)
		require.Error(t, err)
		assert.Equal(t, false, ok)
	}
	{
		// Тест с попыткой изменить несуществующие данные
		userID := "not exists user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("not exists test data"),
			Name:          "not exists test data name",
		}
		status := data.NEW
		m.EXPECT().ReplaceEncryptedData(gomock.Any(), userID, encrData, status).Return(false, nil)
		ok, err := ReplaceEncryptedDataToLocalStorage(context.Background(), userID, m, encrData, status)
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
}

func TestOfflineReplaceEncryptedData(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedClientStorage(ctrl)

	{
		// Успешная замена старых данных на новые со статусом NEW
		userID := "success new user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("success new test data"),
			Name:          "success new test data name",
		}
		m.EXPECT().GetStatus(gomock.Any(), userID, encrData.Name).Return(data.NEW, true, nil)
		m.EXPECT().ReplaceEncryptedData(gomock.Any(), userID, encrData, data.NEW).Return(true, nil)

		ok, err := OfflineReplaceEncryptedData(context.Background(), userID, m, &encrData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
	}
	{
		// Успешная замена старых данных на новые со статусом CHANGED
		userID := "success saved user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("success saved test data"),
			Name:          "success saved test data name",
		}
		m.EXPECT().GetStatus(gomock.Any(), userID, encrData.Name).Return(data.SAVED, true, nil)
		m.EXPECT().ReplaceEncryptedData(gomock.Any(), userID, encrData, data.CHANGED).Return(true, nil)

		ok, err := OfflineReplaceEncryptedData(context.Background(), userID, m, &encrData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
	}
	{
		// Ошибка при попытке получить статус данных
		userID := "error user id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("error test data"),
			Name:          "error data name",
		}
		m.EXPECT().GetStatus(gomock.Any(), userID, encrData.Name).Return(data.NEW, false, errors.New("some error"))

		_, err := OfflineReplaceEncryptedData(context.Background(), userID, m, &encrData)
		require.Error(t, err)
	}
	{
		// Данных не существует в локальном хранилище
		userID := "not exists id"
		encrData := data.EncryptedData{
			EncryptedData: []byte("not exists data"),
			Name:          "not exists data name",
		}
		m.EXPECT().GetStatus(gomock.Any(), userID, encrData.Name).Return(data.NEW, false, nil)

		ok, err := OfflineReplaceEncryptedData(context.Background(), userID, m, &encrData)
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
}

func TestReplaceEncryptedData(t *testing.T) {
	// вспомогательная функция
	testHandler := func(status int, wantDataName string) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// устанавливаю нужный статус в ответ
			res.WriteHeader(status)

			// Проверяю имя данных в запросе к серверу
			if status == http.StatusOK {
				var dataMetaInfo data.EncryptedData
				dec := json.NewDecoder(req.Body)
				err := dec.Decode(&dataMetaInfo)
				require.NoError(t, err)
				assert.Equal(t, wantDataName, dataMetaInfo.Name)
			}
		}
	}

	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockIEncryptedClientStorage(ctrl)

	// Тест успешной передачи ----------------------------------------------------------------------
	userID := "success user id"
	encrData := data.EncryptedData{
		EncryptedData: []byte("success ecnrypted dat"),
		Name:          "success data name",
	}
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), userID, encrData, data.SAVED).Return(true, nil)

	// Тест с пользователем в статусе офлайн -----------------------------------------------------------
	offlineUserID := "offline user id"
	offlineEncrData := data.EncryptedData{
		EncryptedData: []byte("offline ecnrypted dat"),
		Name:          "offline data name",
	}
	m.EXPECT().GetStatus(gomock.Any(), offlineUserID, offlineEncrData.Name).Return(data.CHANGED, true, nil)
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), offlineUserID, offlineEncrData, data.CHANGED).Return(true, nil)

	// Сервер вернул статус внутренней ошибки ---------------------------------------------------------
	internalErrorUserID := "internal server error user id"
	internalErrorEncrData := data.EncryptedData{
		EncryptedData: []byte("internal server error ecnrypted dat"),
		Name:          "internal server error data name",
	}
	m.EXPECT().GetStatus(gomock.Any(), internalErrorUserID, internalErrorEncrData.Name).Return(data.NEW, true, nil)
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), internalErrorUserID, internalErrorEncrData, data.NEW).Return(true, nil)

	// Статус сервера 404 - данные не найдены на сервере ---------------------------------------------------------
	serverNotFoundUserID := "server not found user id"
	serverNotFoundEncrData := data.EncryptedData{
		EncryptedData: []byte("server not found ecnrypted dat"),
		Name:          "server not found data name",
	}
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), serverNotFoundUserID, serverNotFoundEncrData, data.NEW).Return(false, nil)

	// Ошибка из внутреннего хранилища ----------------------------------------------------------------------
	errorUserID := "error user id"
	errorEncrData := data.EncryptedData{
		EncryptedData: []byte("error ecnrypted dat"),
		Name:          "error data name",
	}
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), errorUserID, errorEncrData, data.SAVED).Return(false, errors.New("some error"))

	// Ошибка из при сохранении данных в локальном хранилище ----------------------------------------------------------------------
	saveSuccessUserID := "save success user id"
	saveSuccessEncrData := data.EncryptedData{
		EncryptedData: []byte("save success ecnrypted dat"),
		Name:          "save success data name",
	}
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), saveSuccessUserID, saveSuccessEncrData, data.SAVED).Return(false, nil)
	m.EXPECT().AddEncryptedData(gomock.Any(), saveSuccessUserID, saveSuccessEncrData, data.SAVED).Return(true, nil)

	// Ошибка из при сохранении данных в локальном хранилище ----------------------------------------------------------------------
	saveErrorUserID := "save error user id"
	saveErrorEncrData := data.EncryptedData{
		EncryptedData: []byte("save error ecnrypted dat"),
		Name:          "save error data name",
	}
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), saveErrorUserID, saveErrorEncrData, data.SAVED).Return(false, nil)
	m.EXPECT().AddEncryptedData(gomock.Any(), saveErrorUserID, saveErrorEncrData, data.SAVED).Return(false, errors.New("some error"))

	// Данные уже существуют при сохранении данных в локальном хранилище ----------------------------------------------------------------------
	saveAlreadyExistsUserID := "save already exists user id"
	saveAlreadyExistsEncrData := data.EncryptedData{
		EncryptedData: []byte("save already exists ecnrypted dat"),
		Name:          "save already exists data name",
	}
	m.EXPECT().ReplaceEncryptedData(gomock.Any(), saveAlreadyExistsUserID, saveAlreadyExistsEncrData, data.SAVED).Return(false, nil)
	m.EXPECT().AddEncryptedData(gomock.Any(), saveAlreadyExistsUserID, saveAlreadyExistsEncrData, data.SAVED).Return(false, nil)

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
			name: "not connection",
			req: request{
				userID:      offlineUserID,
				encrData:    offlineEncrData,
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
			name: "status internal server error",
			req: request{
				userID:      internalErrorUserID,
				encrData:    internalErrorEncrData,
				startServer: false,
				stor:        m,
				httpStatus:  500,
			},
			want: want{
				ok:  true,
				err: false,
			},
		},
		{
			name: "server not found data",
			req: request{
				userID:      serverNotFoundUserID,
				encrData:    serverNotFoundEncrData,
				startServer: true,
				stor:        m,
				httpStatus:  404,
			},
			want: want{
				ok:  false,
				err: false,
			},
		},
		{
			name: "local storage error",
			req: request{
				userID:      errorUserID,
				encrData:    errorEncrData,
				startServer: true,
				stor:        m,
				httpStatus:  200,
			},
			want: want{
				ok:  false,
				err: true,
			},
		},
		{
			name: "success save new data",
			req: request{
				userID:      saveSuccessUserID,
				encrData:    saveSuccessEncrData,
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
			name: "error while save new data",
			req: request{
				userID:      saveErrorUserID,
				encrData:    saveErrorEncrData,
				startServer: true,
				stor:        m,
				httpStatus:  200,
			},
			want: want{
				ok:  false,
				err: true,
			},
		},
		{
			name: "data is already exists while save new data",
			req: request{
				userID:      saveAlreadyExistsUserID,
				encrData:    saveAlreadyExistsEncrData,
				startServer: true,
				stor:        m,
				httpStatus:  200,
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
			r.Post("/test", testHandler(tt.req.httpStatus, tt.req.encrData.Name))

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

			ok, err := ReplaceEncryptedData(context.Background(), tt.req.userID, url, resty.New(), tt.req.stor, &tt.req.encrData)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ok, ok)
			}
		})
	}
}
