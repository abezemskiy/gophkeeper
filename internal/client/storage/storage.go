package storage

import (
	"context"

	"github.com/abezemskiy/gophkeeper/internal/client/identity"
	"github.com/abezemskiy/gophkeeper/internal/repositories/data"
	repoStorage "github.com/abezemskiy/gophkeeper/internal/repositories/storage"
)

// Описание интерфейса для постоянно хранилища зашифрованных данных клиента.
type (
	// EncryptedDataGetterByStatus - интерфейс для получения всех данных пользователя с нужным статусом.
	EncryptedDataGetterByStatus interface {
		GetEncryptedDataByStatus(ctx context.Context, userID string, status int) ([][]data.EncryptedData, error) // Возвращает зашифрованные данные с указанным статусом.
	}

	// EncryptedDataStatusChecker - интерфес для проверки статуса данных пользователя по ID и имени данных.
	EncryptedDataStatusChecker interface {
		GetStatus(ctx context.Context, userID, dataName string) (status int, ok bool, err error) // Метод для получения текущего статуса данных.
	}

	// IEncryptedClientStorage - интерфейс клиента для хранения зашифрованных данных.
	IEncryptedClientStorage interface {
		repoStorage.IEncryptedStorage
		EncryptedDataGetterByStatus
		EncryptedDataStatusChecker
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
