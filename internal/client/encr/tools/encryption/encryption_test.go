package encryption

import (
	"testing"

	"github.com/abezemskiy/gophkeeper/internal/client/encr/tools/random"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptAES256(t *testing.T) {
	{
		// Успешное шифрование данных
		// создаю тестовые данные для шифрования
		testData, err := random.GenerateCryptoRandom(100)
		require.NoError(t, err)

		// создаю тестовый ключ для шифрования
		key, err := random.GenerateCryptoRandom(32)
		require.NoError(t, err)

		encrData, err := EncryptAES256(key, testData)
		require.NoError(t, err)

		// Расшифровываю данные для проверки
		decrData, err := DecryptAES256(key, encrData)
		require.NoError(t, err)

		assert.Equal(t, testData, decrData)
	}
	{
		// Тест с ключем неподходящей длины
		key, err := random.GenerateCryptoRandom(16)
		require.NoError(t, err)

		_, err = EncryptAES256(key, []byte("some test data"))
		require.Error(t, err)
	}
}

func TestDecryptAES256(t *testing.T) {
	{
		// Успешное шифрование данных
		// создаю тестовые данные для шифрования
		testData, err := random.GenerateCryptoRandom(100)
		require.NoError(t, err)

		// создаю тестовый ключ для шифрования
		key, err := random.GenerateCryptoRandom(32)
		require.NoError(t, err)

		encrData, err := EncryptAES256(key, testData)
		require.NoError(t, err)

		// Расшифровываю данные для проверки
		decrData, err := DecryptAES256(key, encrData)
		require.NoError(t, err)

		assert.Equal(t, testData, decrData)
	}
	{
		// Тест с ключем неподходящей длины
		key, err := random.GenerateCryptoRandom(16)
		require.NoError(t, err)

		_, err = DecryptAES256(key, []byte("some test data"))
		require.Error(t, err)
	}
}
