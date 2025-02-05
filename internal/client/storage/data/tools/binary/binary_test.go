package binary

import (
	"gophkeeper/internal/client/storage/data"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Создает файл с заданными байтами.
func createFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func TestGetFileType(t *testing.T) {
	{
		// попытка определения типа пустого файла
		pngName := "test-png.png"
		_, err := os.Create(pngName)
		require.NoError(t, err)
		defer os.Remove(pngName)
		// определяю тип пустого файла
		_, err = GetFileType(pngName)
		require.Error(t, err)
	}
	{
		// определяю тип файла png
		pngName := "test-png.png"

		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		err := createFile(pngName, pngData)
		require.NoError(t, err)

		defer os.Remove(pngName)
		// определяю тип файла
		typeFile, err := GetFileType(pngName)
		require.NoError(t, err)
		assert.Equal(t, "image/png", typeFile)
	}
	{
		// определяю тип файла pdf
		name := "test.pdf"

		data := []byte("%PDF-1.4\n")
		err := createFile(name, data)
		require.NoError(t, err)

		defer os.Remove(name)
		// определяю тип файла
		typeFile, err := GetFileType(name)
		require.NoError(t, err)
		assert.Equal(t, "application/pdf", typeFile)
	}
	{
		// определяю тип файла jpeg
		name := "test.jpeg"

		data := []byte{0xFF, 0xD8, 0xFF}
		err := createFile(name, data)
		require.NoError(t, err)

		defer os.Remove(name)
		// определяю тип файла
		typeFile, err := GetFileType(name)
		require.NoError(t, err)
		assert.Equal(t, "image/jpeg", typeFile)
	}
	{
		// определяю тип файла с неизвестным расширением
		name := "test.bin"

		data := []byte("some contetnt")
		err := createFile(name, data)
		require.NoError(t, err)

		defer os.Remove(name)
		// определяю тип файла
		typeFile, err := GetFileType(name)
		require.NoError(t, err)
		assert.Equal(t, "application/octet-stream", typeFile)
	}
	{
		// попытка определить тип несущетсвующего файла
		_, err := GetFileType("wrong-file")
		require.Error(t, err)
	}
}

func TestRestoreFile(t *testing.T) {
	{
		// создаю тестовый png файл
		pngName := "test-png"
		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

		b := data.Binary{
			Binary: pngData,
			Type:   "image/png",
		}

		// восстанавливаю исходный файл
		dir := "./"
		err := RestoreFile(b, pngName, dir)
		require.NoError(t, err)

		// Проверяю тип восстановленного файла
		getType, err := GetFileType(dir + pngName + ".png")
		require.NoError(t, err)
		assert.Equal(t, "image/png", getType)

		// удаляю тестовый файл
		defer os.Remove(dir + pngName + ".png")
	}
	{
		// создаю тестовый png файл
		pngName := "test-png"
		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

		err := createFile(pngName+".png", pngData)
		require.NoError(t, err)

		// преобразую содержимое файла в слайс байт
		byteOfData, err := os.ReadFile(pngName + ".png")
		require.NoError(t, err)
		err = os.Remove(pngName + ".png")
		require.NoError(t, err)

		// создаю структуру хранения и передачи бинарных данных
		b := data.Binary{
			Binary: byteOfData,
			Type:   "image/png",
		}

		// восстанавливаю исходный файл
		dir := "./"
		err = RestoreFile(b, pngName, dir)
		require.NoError(t, err)

		// Проверяю тип восстановленного файла
		getType, err := GetFileType(dir + pngName + ".png")
		require.NoError(t, err)
		assert.Equal(t, "image/png", getType)

		// удаляю восстановленный тестовый файл
		defer os.Remove(dir + pngName + ".png")
	}
	{
		// создаю тестовый jpg файл
		name := "test"

		b := data.Binary{
			Binary: []byte{0xFF, 0xD8, 0xFF},
			Type:   "image/jpeg",
		}

		// восстанавливаю исходный файл
		dir := "./"
		err := RestoreFile(b, name, dir)
		require.NoError(t, err)

		// Проверяю тип восстановленного файла
		getType, err := GetFileType(dir + name + ".jpg")
		require.NoError(t, err)
		assert.Equal(t, "image/jpeg", getType)

		// удаляю тестовый файл
		defer os.Remove(dir + name + ".jpg")
	}
	{
		// создаю тестовый pdf файл
		name := "test"

		b := data.Binary{
			Binary: []byte("%PDF-1.4\n"),
			Type:   "application/pdf",
		}

		// восстанавливаю исходный файл
		dir := "./"
		err := RestoreFile(b, name, dir)
		require.NoError(t, err)

		// Проверяю тип восстановленного файла
		getType, err := GetFileType(dir + name + ".pdf")
		require.NoError(t, err)
		assert.Equal(t, "application/pdf", getType)

		// удаляю тестовый файл
		defer os.Remove(dir + name + ".pdf")
	}
	{
		// создаю тестовый файл типа bin
		name := "test"

		b := data.Binary{
			Binary: []byte("come content"),
			Type:   "application/octet-stream",
		}

		// восстанавливаю исходный файл
		dir := "./"
		err := RestoreFile(b, name, dir)
		require.NoError(t, err)

		// Проверяю тип восстановленного файла
		getType, err := GetFileType(dir + name + ".bin")
		require.NoError(t, err)
		assert.Equal(t, "application/octet-stream", getType)

		// удаляю тестовый файл
		defer os.Remove(dir + name + ".bin")
	}
}
