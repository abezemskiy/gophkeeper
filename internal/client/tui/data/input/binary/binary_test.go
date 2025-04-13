package binary

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	{
		// Тест с успешным преобразованием
		// создаю файл
		name := "data.bin"
		f, err := os.Create(name)
		require.NoError(t, err)
		defer os.Remove(name)

		// Записываю данные в файл
		d := []byte("some data")
		_, err = f.Write(d)
		require.NoError(t, err)

		testInfo := DataInfo{
			Path: name,
		}
		err = parseFile(&testInfo)
		require.NoError(t, err)

		// Проверяю содердимое структуры
		assert.Equal(t, "application/octet-stream", testInfo.Binary.Type)
		assert.Equal(t, string(d), string(testInfo.Binary.Binary))
	}
	{
		// Попытка преобразования несуществующего файла
		err := parseFile(&DataInfo{Path: "not-exist-file.txt"})
		require.Error(t, err)
	}
}
