package inmemory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/abezemskiy/gophkeeper/internal/client/encr"
	"github.com/abezemskiy/gophkeeper/internal/client/identity"
	"github.com/abezemskiy/gophkeeper/internal/repositories/data"
	"github.com/abezemskiy/gophkeeper/internal/repositories/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func DataIsEqual(want, get [][]data.Data) bool {
	if len(want) != len(get) {
		return false
	}

	for i, versions := range want {
		if len(versions) != len(get[i]) {
			return false
		}
		for j, d := range versions {
			if d.Name != get[i][j].Name {
				return false
			}
			if string(d.Data) != string(get[i][j].Data) {
				return false
			}
		}
	}
	return true
}

func TestSetUpdatingPeriod(t *testing.T) {
	test := 10
	SetUpdatingPeriod(test)
	assert.Equal(t, test, updatingPeriod)
}

func TestGetUpdatingPeriod(t *testing.T) {
	test := 15
	updatingPeriod = test
	get := GetUpdatingPeriod()
	assert.Equal(t, time.Second*time.Duration(test), get)
}

func TestUpdate(t *testing.T) {
	// регистрирую мок хранилища идентификационных данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stor := mocks.NewMockIEncryptedClientStorage(ctrl)

	{
		// Успешное обновление и получение данных
		inmemo := NewDecryptedData()

		info := mocks.NewMockIUserInfoStorage(ctrl)
		pass := "success password"
		authData := identity.AuthData{
			Login:    "success login",
			Password: pass,
		}
		id := "success id"
		info.EXPECT().Get().Return(authData, id)

		// создаю тестовые данные
		testData := make([][]data.Data, 2)
		testData[0] = []data.Data{{Data: []byte("1 first version"), Name: "first data"}}
		testData[1] = []data.Data{{Data: []byte("1 second version"), Name: "second data"}, {Data: []byte("2 second version"), Name: "second data"}}

		// Шифрую данные
		testEncrData := make([][]data.EncryptedData, 2)
		for i, testVersions := range testData {
			encrForSave := make([]data.EncryptedData, len(testVersions))
			for j, d := range testVersions {
				e, err := encr.EncryptData(pass, &d)
				require.NoError(t, err)
				encrForSave[j] = *e
			}
			testEncrData[i] = encrForSave
		}

		stor.EXPECT().GetAllEncryptedData(gomock.Any(), id).Return(testEncrData, nil)

		// Сохраняю данные в хранилище
		err := inmemo.Update(context.Background(), stor, info)
		require.NoError(t, err)

		// Извлекаю данные из хранилища
		getData := inmemo.GetAll()

		// Проверяю на равенство данные полученные из хранилища с теми, которые изначально туда загружались
		assert.Equal(t, true, DataIsEqual(testData, getData))
	}
	{
		// Возвращение пустых данных
		inmemo := NewDecryptedData()

		info := mocks.NewMockIUserInfoStorage(ctrl)
		pass := "empty password"
		authData := identity.AuthData{
			Login:    "empty login",
			Password: pass,
		}
		id := "empty id"
		info.EXPECT().Get().Return(authData, id)

		stor.EXPECT().GetAllEncryptedData(gomock.Any(), id).Return([][]data.EncryptedData{}, nil)

		err := inmemo.Update(context.Background(), stor, info)
		require.NoError(t, err)

		// Извлекаю данные из хранилища
		getData := inmemo.GetAll()
		assert.Equal(t, 0, len(getData))
	}
	{
		// Ошибка из постоянного хранилища данных
		inmemo := NewDecryptedData()

		info := mocks.NewMockIUserInfoStorage(ctrl)
		pass := "error password"
		authData := identity.AuthData{
			Login:    "error login",
			Password: pass,
		}
		id := "error id"
		info.EXPECT().Get().Return(authData, id)

		stor.EXPECT().GetAllEncryptedData(gomock.Any(), id).Return(nil, errors.New("some error"))

		err := inmemo.Update(context.Background(), stor, info)
		require.Error(t, err)
	}
}
