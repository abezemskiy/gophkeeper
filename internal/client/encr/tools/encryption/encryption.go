// Пакет для шифрования и расшифровывания данных пользователя.
// Для шифрования данных в приложении используется мастер пароль.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"

	"github.com/abezemskiy/gophkeeper/internal/client/encr/tools/random"
)

// EncryptAES256 - функция для шифрования данных с помощью алгоритма AES256.
func EncryptAES256(key []byte, data []byte) ([]byte, error) {
	// Проверка длины ключа для соответстия алгоритму AES256
	if len(key) < 32 {
		return nil, errors.New("lenth of key is not equal 32 for AES256")
	}

	// NewCipher создает и возвращает новый cipher.Block.
	// Ключевым аргументом должен быть ключ AES, 16, 24 или 32 байта
	// для выбора AES-128, AES-192 или AES-256.
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher.Block, %w", err)
	}

	// NewGCM возвращает заданный 128-битный блочный шифр
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gcm, %w", err)
	}

	// создаём вектор инициализации
	nonce, err := random.GenerateCryptoRandom(aesgcm.NonceSize())
	if err != nil {
		return nil, fmt.Errorf("failed to create new initialization vector, %w", err)
	}

	encrData := aesgcm.Seal(nil, nonce, data, nil) // зашифровываем
	result := append(nonce, encrData...)

	return result, nil
}

// DecryptAES256 - функция для расшифровывания данных с помощью алгоритма AES256.
func DecryptAES256(key []byte, encrData []byte) ([]byte, error) {
	// Проверка длины ключа для соответстия алгоритму AES256
	if len(key) < 32 {
		return nil, errors.New("lenth of key is not equal 32 for AES256")
	}

	// NewCipher создает и возвращает новый cipher.Block.
	// Ключевым аргументом должен быть ключ AES, 16, 24 или 32 байта
	// для выбора AES-128, AES-192 или AES-256.
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher.Block, %w", err)
	}

	// NewGCM возвращает заданный 128-битный блочный шифр
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gcm, %w", err)
	}

	// извлекаю вектор инициализации из полученных данных
	nonce := encrData[:aesgcm.NonceSize()]

	data, err := aesgcm.Open(nil, nonce, encrData[aesgcm.NonceSize():], nil) // расшифровываем
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data, %w", err)
	}
	return data, nil
}
