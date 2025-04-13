// Пакет для шифрования и расшифровывания пользовательских данных с помощью мастер пароля.
package encr

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/abezemskiy/gophkeeper/internal/client/encr/tools/encryption"
	"github.com/abezemskiy/gophkeeper/internal/client/encr/tools/key"
	"github.com/abezemskiy/gophkeeper/internal/repositories/data"
)

// EncryptData - функция для шифрования пользовательских данных.
func EncryptData(passwrod string, userData *data.Data) (*data.EncryptedData, error) {
	// создаю ключ шифрования из мастер пароля пользователя
	key := key.DeriveKey(passwrod, 32)

	// Сериализую данные пользователя в массив байт
	var bufEncode bytes.Buffer
	if err := json.NewEncoder(&bufEncode).Encode(*userData); err != nil {
		return nil, fmt.Errorf("failed to encode data, %w", err)
	}

	// Шифрую данные
	encrDta, err := encryption.EncryptAES256(key, bufEncode.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data, %w", err)
	}

	return &data.EncryptedData{EncryptedData: encrDta, Name: userData.Name}, nil
}

// DecryptData - функция для шифрования пользовательских данных.
func DecryptData(passwrod string, encrData *data.EncryptedData) (*data.Data, error) {
	// создаю ключ шифрования из мастер пароля пользователя
	key := key.DeriveKey(passwrod, 32)

	// Расшифровываю данные
	res, err := encryption.DecryptAES256(key, encrData.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data, %w", err)
	}

	// Десериализую расшифрованные данные в структуру
	var userData data.Data
	r := bytes.NewReader(res)
	if err := json.NewDecoder(r).Decode(&userData); err != nil {
		return nil, fmt.Errorf("failed decode data from bites to structer, %w", err)
	}

	return &userData, nil
}
