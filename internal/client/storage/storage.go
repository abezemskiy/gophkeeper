package storage

import (
	"context"
	"gophkeeper/internal/repositories/data"
	repoStorage "gophkeeper/internal/repositories/storage"
)

type (
	// EncryptedDataGetterByStatus - интерфейс для получения всех данных пользователя с нужным статусом.
	EncryptedDataGetterByStatus interface {
		GetEncryptedDataByStatus(ctx context.Context, userID string, status int) ([][]data.EncryptedData, error) // Возвращает зашифрованные данные с указанным статусом.
	}

	// IEncryptedClientStorage - интерфейс клиента для хранения зашифрованных данных.
	IEncryptedClientStorage interface {
		repoStorage.IEncryptedStorage
		EncryptedDataGetterByStatus
		ChangeStatusOfEncryptedData(ctx context.Context, userID, dataName string, newStatus int) (ok bool, err error) // Изменяет статус существующих данных.
	}
)
