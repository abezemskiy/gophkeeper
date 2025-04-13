package binary

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/abezemskiy/gophkeeper/internal/client/storage/data"
)

// GetFileType - функция для определения типа файла.
func GetFileType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file, %w", err)
	}
	defer file.Close()

	// проверяю размер файла
	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file stat, %w", err)
	}
	// Если файл пустой, возвращаю ошибку
	if stat.Size() == 0 {
		return "", fmt.Errorf("file is empty")
	}

	// Читаем первые 512 байт (этого достаточно для определения MIME-типа)
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file, %w", err)
	}

	return http.DetectContentType(buffer), nil
}

// RestoreFile - функция для восстановления файла из слайса байт.
func RestoreFile(rData data.Binary, dataName, outputDir string) error {
	ext := ""
	// Устанавливаю расширение итогового файла.
	switch rData.Type {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpg"
	case "application/pdf":
		ext = ".pdf"
	default:
		ext = ".bin" // Неизвестный формат
	}

	filename := filepath.Join(outputDir, dataName+ext)
	return os.WriteFile(filename, rData.Binary, 0644)
}
