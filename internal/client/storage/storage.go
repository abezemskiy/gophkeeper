package storage

import (
	"context"
	"gophkeeper/internal/repositories/data"
	repoStorage "gophkeeper/internal/repositories/storage"
)

type (
	// EncryptedDataGetterByStatus - интерфейс для получения всех данных пользователя с нужным статусом.
	EncryptedDataGetterByStatus interface {
		GetEncryptedDataByStatus(ctx context.Context, idUser string, status int) ([][]data.EncryptedData, error) // Возвращает зашифрованные данные с указанным статусом.
	}

	// IEncryptedClientStorage - интерфейс клиента для хранения зашифрованных данных.
	IEncryptedClientStorage interface {
		repoStorage.IEncryptedStorage
		EncryptedDataGetterByStatus
	}
)
