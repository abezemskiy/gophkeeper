package encr

import (
	"testing"
	"time"

	"github.com/abezemskiy/gophkeeper/internal/repositories/data"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CompareData(lData *data.Data, rData *data.Data) bool {
	dLenth := len(lData.Data) == len(rData.Data)
	d := string(lData.Data) == string(rData.Data)
	t := lData.Type == rData.Type
	n := lData.Name == rData.Name
	m := lData.Metainfo == rData.Metainfo
	s := lData.Status == rData.Status
	c := lData.CreateDate.Equal(rData.CreateDate)
	e := lData.EditDate.Equal(rData.EditDate)
	return dLenth && d && t && n && m && s && c && e
}

func TestEncryptData(t *testing.T) {
	{
		// Тест с успешным шифрованием данных
		testData := data.Data{
			Data:       []byte("some strong pair of login and password"),
			Type:       data.PASSWORD,
			Name:       "test password",
			Metainfo:   "some metainfo",
			Status:     data.NEW,
			CreateDate: time.Now(),
			EditDate:   time.Now(),
		}
		testPass := "some strong master password of user"

		// Шифрую данные
		testEncrData, err := EncryptData(testPass, &testData)
		require.NoError(t, err)

		// расшифровываю данные
		testDecrData, err := DecryptData(testPass, testEncrData)
		require.NoError(t, err)

		assert.Equal(t, true, CompareData(&testData, testDecrData))
	}
}

func TestDecryptData(t *testing.T) {
	{
		// Тест с успешным шифрованием и расшифровыванием данных
		testData := data.Data{
			Data:       []byte("some strong pair of login and password"),
			Type:       data.PASSWORD,
			Name:       "test password",
			Metainfo:   "some metainfo",
			Status:     data.NEW,
			CreateDate: time.Now(),
			EditDate:   time.Now(),
		}
		testPass := "some strong master password of user"

		// Шифрую данные
		testEncrData, err := EncryptData(testPass, &testData)
		require.NoError(t, err)

		// расшифровываю данные
		testDecrData, err := DecryptData(testPass, testEncrData)
		require.NoError(t, err)

		assert.Equal(t, true, CompareData(&testData, testDecrData))
	}
	{
		// Тест с попыткой использовать неправильный ключ для расшифровывания данных
		testData := data.Data{
			Data:       []byte("some strong pair of login and password"),
			Type:       data.PASSWORD,
			Name:       "test password",
			Metainfo:   "some metainfo",
			Status:     data.NEW,
			CreateDate: time.Now(),
			EditDate:   time.Now(),
		}
		testPass := "some strong master password of user"

		// Шифрую данные
		testEncrData, err := EncryptData(testPass, &testData)
		require.NoError(t, err)

		// расшифровываю данные
		_, err = DecryptData("wrong strong password", testEncrData)
		require.Error(t, err)

	}
}
