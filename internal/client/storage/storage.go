package storage

import (
	"context"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/repositories/data"
	repoStorage "gophkeeper/internal/repositories/storage"
)

// Описание интерфейса для постоянно хранилища зашифрованных данных клиента.
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

		ReplaceDataWithMultiVersionData(ctx context.Context, idUser string, data []data.EncryptedData,
			status int) (bool, error) // Для замены существующих в хранилище на данные с несколькими версиями
	}
)

// Описание интерфейса для временного хранилища расшифрованных данных клиента.
type (
	// DataWriter - интерфейс для добавления данных во временное хранилище.
	DataWriter interface {
		Update(ctx context.Context, stor IEncryptedClientStorage, info identity.IUserInfoStorage) error // обновляю данные пользователя из постоянного хранилища.
	}

	// DataReader - интерфейс для выгрузки данных у конкретного пользователя по его id.
	DataReader interface {
		GetAll() [][]data.Data // Возвращает слайс расшифрованных данных.
	}

	// IStorage - интерфейс хранения данных пользователей в незашифрованном виде.
	IStorage interface {
		DataWriter
		DataReader
	}
)
