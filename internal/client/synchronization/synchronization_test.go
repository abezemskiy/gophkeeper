package synchronization

import (
	"context"
	"encoding/json"

	"errors"

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

	// Тест с успешным сохранением локальных данных со статусом NEW на сервере------------------------------------------
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

	// Тест с ошибкой из хранилища при попытке извлечь все данные пользователя ------------------------------------------
	getErrorID := "get error id"
	getErrorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	getErrorInfo.EXPECT().Get().Return(identity.AuthData{}, getErrorID)
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), getErrorID, data.NEW).Return(nil, errors.New("some error"))

	// Тест с ошибкой из обработки невалидных данных полученных из хранилища ------------------------------------------
	wrongDataID := "wrong data id"
	wrongDataInfo := mocks.NewMockIUserInfoStorage(ctrl)
	wrongDataInfo.EXPECT().Get().Return(identity.AuthData{}, wrongDataID)
	wantWrongData := [][]data.EncryptedData{
		{},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), wrongDataID, data.NEW).Return(wantWrongData, nil)

	// Тест с ошибкой из обработки невалидных данных полученных из хранилища ------------------------------------------
	wrongDataTooMuchVersionID := "wrong data too much versionid"
	wrongDataTooMuchInfo := mocks.NewMockIUserInfoStorage(ctrl)
	wrongDataTooMuchInfo.EXPECT().Get().Return(identity.AuthData{}, wrongDataTooMuchVersionID)
	wantWrongDataTooMuchVersion := [][]data.EncryptedData{
		{{EncryptedData: []byte("first version encr data"), Name: "first wrong encr data name"},
			{EncryptedData: []byte("second version encr data"), Name: "first wrong encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), wrongDataTooMuchVersionID, data.NEW).Return(wantWrongDataTooMuchVersion, nil)

	// Тест с попыткой отправить данные на сервер по неверному адресу -----------------------------------------------
	badURLID := "bad url id"
	badURLInfo := mocks.NewMockIUserInfoStorage(ctrl)
	badURLInfo.EXPECT().Get().Return(identity.AuthData{}, badURLID)
	badURLData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first version encr data"), Name: "first encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), badURLID, data.NEW).Return(badURLData, nil)

	// Тест - сервер возращает статус конфликта данных -----------------------------------------------
	dataConflictID := "data conflict id"
	dataConflictInfo := mocks.NewMockIUserInfoStorage(ctrl)
	dataConflictInfo.EXPECT().Get().Return(identity.AuthData{}, dataConflictID)
	conflictData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first version encr data"), Name: "first data conflict encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second data conflict encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), dataConflictID, data.NEW).Return(conflictData, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), dataConflictID, "first data conflict encr data name", data.CHANGED).Return(true, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), dataConflictID, "second data conflict encr data name", data.CHANGED).Return(true, nil)

	// Тест - ошибка изменения статуса данных -----------------------------------------------
	changeStatusErrorID := "change status error id"
	changeStatusErrorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	changeStatusErrorInfo.EXPECT().Get().Return(identity.AuthData{}, changeStatusErrorID)
	changeStatusErrorData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first change status version encr data"), Name: "change status error data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), changeStatusErrorID, data.NEW).Return(changeStatusErrorData, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), changeStatusErrorID, "change status error data name", data.SAVED).
		Return(true, errors.New("some error"))

	// Тест - пользователь или данные не найдены  ------------------------------------------------------------------------
	notFoundID := "not found id"
	notFoundInfo := mocks.NewMockIUserInfoStorage(ctrl)
	notFoundInfo.EXPECT().Get().Return(identity.AuthData{}, notFoundID)
	notFoundData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first change status version encr data"), Name: "not found data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), notFoundID, data.NEW).Return(notFoundData, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), notFoundID, "not found data name", data.SAVED).Return(false, nil)

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
		{
			name: "get all encrypted data error",
			req: request{
				stor:        stor,
				info:        getErrorInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad data",
			req: request{
				stor:        stor,
				info:        wrongDataInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad data, too much version",
			req: request{
				stor:        stor,
				info:        wrongDataTooMuchInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad url",
			req: request{
				stor:        stor,
				info:        badURLInfo,
				setValidURL: false,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "data conflict",
			req: request{
				stor:        stor,
				info:        dataConflictInfo,
				setValidURL: true,
				status:      409,
			},
			want: want{
				err: false,
			},
		},
		{
			name: "change status error id",
			req: request{
				stor:        stor,
				info:        changeStatusErrorInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "not found error",
			req: request{
				stor:        stor,
				info:        notFoundInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
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

func TestSynchronizeChangedLocalData(t *testing.T) {
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

	// Тест с успешным сохранением локальных данных со статусом NEW на сервере------------------------------------------
	successID := "success id"
	successInfo := mocks.NewMockIUserInfoStorage(ctrl)
	successInfo.EXPECT().Get().Return(identity.AuthData{}, successID)
	wantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data"), Name: "first encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
		{{EncryptedData: []byte("third encr data"), Name: "third encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), successID, data.CHANGED).Return(wantData, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), successID, "first encr data name", data.SAVED).Return(true, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), successID, "second encr data name", data.SAVED).Return(true, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), successID, "third encr data name", data.SAVED).Return(true, nil)

	// Тест с ошибкой из хранилища при попытке извлечь все данные пользователя ------------------------------------------
	getErrorID := "get error id"
	getErrorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	getErrorInfo.EXPECT().Get().Return(identity.AuthData{}, getErrorID)
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), getErrorID, data.CHANGED).Return(nil, errors.New("some error"))

	// Тест с ошибкой из обработки невалидных данных полученных из хранилища ------------------------------------------
	wrongDataID := "wrong data id"
	wrongDataInfo := mocks.NewMockIUserInfoStorage(ctrl)
	wrongDataInfo.EXPECT().Get().Return(identity.AuthData{}, wrongDataID)
	wantWrongData := [][]data.EncryptedData{
		{},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), wrongDataID, data.CHANGED).Return(wantWrongData, nil)

	// Тест с ошибкой из обработки невалидных данных полученных из хранилища ------------------------------------------
	wrongDataTooMuchVersionID := "wrong data too much versionid"
	wrongDataTooMuchInfo := mocks.NewMockIUserInfoStorage(ctrl)
	wrongDataTooMuchInfo.EXPECT().Get().Return(identity.AuthData{}, wrongDataTooMuchVersionID)
	wantWrongDataTooMuchVersion := [][]data.EncryptedData{
		{{EncryptedData: []byte("first version encr data"), Name: "first wrong encr data name"},
			{EncryptedData: []byte("second version encr data"), Name: "first wrong encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), wrongDataTooMuchVersionID, data.CHANGED).Return(wantWrongDataTooMuchVersion, nil)

	// Тест с попыткой отправить данные на сервер по неверному адресу -----------------------------------------------
	badURLID := "bad url id"
	badURLInfo := mocks.NewMockIUserInfoStorage(ctrl)
	badURLInfo.EXPECT().Get().Return(identity.AuthData{}, badURLID)
	badURLData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first version encr data"), Name: "first encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), badURLID, data.CHANGED).Return(badURLData, nil)

	// Тест - сервер возращает статус конфликта данных -----------------------------------------------
	dataConflictID := "data conflict id"
	dataConflictInfo := mocks.NewMockIUserInfoStorage(ctrl)
	dataConflictInfo.EXPECT().Get().Return(identity.AuthData{}, dataConflictID)
	conflictData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first version encr data"), Name: "first data conflict encr data name"}},
		{{EncryptedData: []byte("second encr data"), Name: "second data conflict encr data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), dataConflictID, data.CHANGED).Return(conflictData, nil)

	// Тест - ошибка изменения статуса данных -----------------------------------------------
	changeStatusErrorID := "change status error id"
	changeStatusErrorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	changeStatusErrorInfo.EXPECT().Get().Return(identity.AuthData{}, changeStatusErrorID)
	changeStatusErrorData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first change status version encr data"), Name: "change status error data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), changeStatusErrorID, data.CHANGED).Return(changeStatusErrorData, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), changeStatusErrorID, "change status error data name", data.SAVED).
		Return(true, errors.New("some error"))

	// Тест - пользователь или данные не найдены  ------------------------------------------------------------------------
	notFoundID := "not found id"
	notFoundInfo := mocks.NewMockIUserInfoStorage(ctrl)
	notFoundInfo.EXPECT().Get().Return(identity.AuthData{}, notFoundID)
	notFoundData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first change status version encr data"), Name: "not found data name"}},
	}
	stor.EXPECT().GetEncryptedDataByStatus(gomock.Any(), notFoundID, data.CHANGED).Return(notFoundData, nil)
	stor.EXPECT().ChangeStatusOfEncryptedData(gomock.Any(), notFoundID, "not found data name", data.SAVED).Return(false, nil)

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
		{
			name: "get all encrypted data error",
			req: request{
				stor:        stor,
				info:        getErrorInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad data",
			req: request{
				stor:        stor,
				info:        wrongDataInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad data, too much version",
			req: request{
				stor:        stor,
				info:        wrongDataTooMuchInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad url",
			req: request{
				stor:        stor,
				info:        badURLInfo,
				setValidURL: false,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "data conflict",
			req: request{
				stor:        stor,
				info:        dataConflictInfo,
				setValidURL: true,
				status:      409,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "change status error id",
			req: request{
				stor:        stor,
				info:        changeStatusErrorInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "not found error",
			req: request{
				stor:        stor,
				info:        notFoundInfo,
				setValidURL: true,
				status:      200,
			},
			want: want{
				err: true,
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

			err := SynchronizeChangedLocalData(context.Background(), tt.req.stor, tt.req.info, resty.New(), url)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSynchronizeDataFromServer(t *testing.T) {
	// Хэндлер для тестовой обработки запроса клиента на авторизацию на сервере
	testHandler := func(status int, serverData [][]data.EncryptedData) http.HandlerFunc {
		return func(res http.ResponseWriter, _ *http.Request) {
			if status == http.StatusOK {
				// Устанавливаю ответ сервера
				err := json.NewEncoder(res).Encode(serverData)
				require.NoError(t, err)
			}

			// устанавливаю нужный статус в ответ
			res.WriteHeader(status)
		}
	}

	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stor := mocks.NewMockIEncryptedClientStorage(ctrl)

	// Тест с успешным изменением существующих локальных данных на актуальные от сервера ------------------------------------------
	successID := "success id"
	successInfo := mocks.NewMockIUserInfoStorage(ctrl)
	successInfo.EXPECT().Get().Return(identity.AuthData{}, successID)
	successWantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data version 1"), Name: "first encr data name"},
			{EncryptedData: []byte("first encr data version 2"), Name: "first encr data name"}},
	}
	stor.EXPECT().ReplaceDataWithMultiVersionData(gomock.Any(), successID, successWantData[0], data.CONFLICT).Return(true, nil)

	// Тест с успешным изменением существующих локальных данных на актуальные от сервера -------------------------------
	newSuccessOneVirsionID := "new success one version id"
	newSuccessOneVirsionInfo := mocks.NewMockIUserInfoStorage(ctrl)
	newSuccessOneVirsionInfo.EXPECT().Get().Return(identity.AuthData{}, newSuccessOneVirsionID)
	newSuccessOneVirsionWantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data version 1"), Name: "first new success one version encr data name"}},
	}
	stor.EXPECT().ReplaceDataWithMultiVersionData(gomock.Any(), newSuccessOneVirsionID,
		newSuccessOneVirsionWantData[0], data.SAVED).Return(true, nil)

	// Тест с добавлением новых данных в локальное хранилище, полученных от сервера -------------------------------
	newSuccessID := "new success id"
	newSuccessInfo := mocks.NewMockIUserInfoStorage(ctrl)
	newSuccessInfo.EXPECT().Get().Return(identity.AuthData{}, newSuccessID)
	newSuccessWantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data version 1"), Name: "first new success encr data name"}},
	}
	stor.EXPECT().ReplaceDataWithMultiVersionData(gomock.Any(), newSuccessID, newSuccessWantData[0], data.SAVED).Return(false, nil)
	stor.EXPECT().AddEncryptedData(gomock.Any(), newSuccessID, newSuccessWantData[0][0], data.SAVED).Return(true, nil)

	// Тест с неспешной попыткой выполнить запрос на сервер ------------------------------------------
	connectionErrorID := "connection error id"
	connectionErrorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	connectionErrorInfo.EXPECT().Get().Return(identity.AuthData{}, connectionErrorID)

	// Тест со возвращением статуса от сервера не http.StatusOk ------------------------------------------
	badStatusID := "bad status id"
	badStatusInfo := mocks.NewMockIUserInfoStorage(ctrl)
	badStatusInfo.EXPECT().Get().Return(identity.AuthData{}, badStatusID)

	// Тест со возвращением статуса от сервера не http.StatusOk ------------------------------------------
	noVersionID := "no version of data id"
	noVersionInfo := mocks.NewMockIUserInfoStorage(ctrl)
	noVersionInfo.EXPECT().Get().Return(identity.AuthData{}, noVersionID)
	noVersionData := [][]data.EncryptedData{{}, {{EncryptedData: []byte("first encr data version 1"), Name: "first encr data name"}}}

	// Тест - ошибка из метода replace ---------------------------------------------------------
	replaceErrorID := "replace error id"
	replaceErrorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	replaceErrorInfo.EXPECT().Get().Return(identity.AuthData{}, replaceErrorID)
	replaceErrorWantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data version 1"), Name: "replace error encr data name"}},
	}
	stor.EXPECT().ReplaceDataWithMultiVersionData(gomock.Any(), replaceErrorID, replaceErrorWantData[0],
		data.SAVED).Return(false, errors.New("some error"))

	// Тест - ошибка из метода AddEncryptedData ---------------------------------------------------------
	newErrorID := "new error id"
	newErrorInfo := mocks.NewMockIUserInfoStorage(ctrl)
	newErrorInfo.EXPECT().Get().Return(identity.AuthData{}, newErrorID)
	newErrorWantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data version 1"), Name: "first new error encr data name"}},
	}
	stor.EXPECT().ReplaceDataWithMultiVersionData(gomock.Any(), newErrorID, newErrorWantData[0], data.SAVED).Return(false, nil)
	stor.EXPECT().AddEncryptedData(gomock.Any(), newErrorID, newErrorWantData[0][0], data.SAVED).Return(false, errors.New("some error"))

	// Тест - попытка добавить данные, которые уже храняться в локальном хранилище,
	// хотя до этого попытка изменения этих данных закончилась неудачей ---------------------------------------------------------
	newIsAlreadyExistsID := "new data is already exists id"
	newIsAlreadyExistsInfo := mocks.NewMockIUserInfoStorage(ctrl)
	newIsAlreadyExistsInfo.EXPECT().Get().Return(identity.AuthData{}, newIsAlreadyExistsID)
	newIsAlreadyExistsWantData := [][]data.EncryptedData{
		{{EncryptedData: []byte("first encr data version 1"), Name: "first new error encr data name"}},
	}
	stor.EXPECT().ReplaceDataWithMultiVersionData(gomock.Any(), newIsAlreadyExistsID, newIsAlreadyExistsWantData[0], data.SAVED).Return(false, nil)
	stor.EXPECT().AddEncryptedData(gomock.Any(), newIsAlreadyExistsID, newIsAlreadyExistsWantData[0][0],
		data.SAVED).Return(false, nil)

	type request struct {
		stor        storage.IEncryptedClientStorage
		info        identity.IUserInfoStorage
		setValidURL bool
		status      int
		serverData  [][]data.EncryptedData
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
				serverData:  successWantData,
			},
			want: want{
				err: false,
			},
		},
		{
			name: "success one version data",
			req: request{
				stor:        stor,
				info:        newSuccessOneVirsionInfo,
				setValidURL: true,
				status:      200,
				serverData:  newSuccessOneVirsionWantData,
			},
			want: want{
				err: false,
			},
		},
		{
			name: "success add new data from server",
			req: request{
				stor:        stor,
				info:        newSuccessInfo,
				setValidURL: true,
				status:      200,
				serverData:  newSuccessWantData,
			},
			want: want{
				err: false,
			},
		},
		{
			name: "connection error",
			req: request{
				stor:        stor,
				info:        connectionErrorInfo,
				setValidURL: false,
				status:      200,
				serverData:  nil,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad status",
			req: request{
				stor:        stor,
				info:        badStatusInfo,
				setValidURL: true,
				status:      500,
				serverData:  nil,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "no version",
			req: request{
				stor:        stor,
				info:        noVersionInfo,
				setValidURL: true,
				status:      200,
				serverData:  noVersionData,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "error from replace method",
			req: request{
				stor:        stor,
				info:        replaceErrorInfo,
				setValidURL: true,
				status:      200,
				serverData:  replaceErrorWantData,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "add new data from server error",
			req: request{
				stor:        stor,
				info:        newErrorInfo,
				setValidURL: true,
				status:      200,
				serverData:  newErrorWantData,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "new data from server is already exists",
			req: request{
				stor:        stor,
				info:        newIsAlreadyExistsInfo,
				setValidURL: true,
				status:      200,
				serverData:  newIsAlreadyExistsWantData,
			},
			want: want{
				err: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаю тестовый http сервер
			r := chi.NewRouter()
			r.Get("/test", testHandler(tt.req.status, tt.req.serverData))

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

			err := SynchronizeDataFromServer(context.Background(), tt.req.stor, tt.req.info, resty.New(), url)
			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
