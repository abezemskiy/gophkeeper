// hasher - пакет со вспомогательными функция для хэширования данных.
package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Hash - функция, которая хэширует переданную строку и возвращает хэш в виде строки.
func CalkHash(data string) (string, error) {
	src := []byte(data)

	// создаю новый hash.Hash, вычисляющий контрольную сумму SHA-256
	h := sha256.New()
	n, err := h.Write(src)
	if err != nil {
		return "", fmt.Errorf("conveing bytes for hashing error, %w", err)
	}
	if n != len(src) {
		return "", fmt.Errorf("count of wrote bytes not equal initial count of bytes")
	}
	// вычисляю хэш
	dst := h.Sum(nil)

	// кодирую хэш в виде слайса байт в строку
	return hex.EncodeToString(dst), nil
}
