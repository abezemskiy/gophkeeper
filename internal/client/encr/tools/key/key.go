// Пакет для хэширования мастер пароля произвольной длины в ключ для алгоритма AES.
package key

import (
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

// DeriveKey - функция для хэширования пароля в ключ длиной len с помощью алгоритма PBKDF2.
func DeriveKey(password string, len int) []byte {
	iterations := 100_000 // Количество итераций (чем больше, тем лучше защита)
	passwordByte := []byte(password)
	return pbkdf2.Key(passwordByte, passwordByte, iterations, len, sha256.New)
}
