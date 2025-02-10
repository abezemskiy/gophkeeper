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
		ReplaceDataWithMultiVersionData(ctx context.Context, idUser string, data []data.EncryptedData,
			status int) (bool, error) // Для замены существующих в хранилище на данные с несколькими версиями
	}
)

type (
	// DataWriter - интерфейс для добавления данных во временное хранилище.
	DataWriter interface {
		AddData(ctx context.Context, idUser, data data.Data) error // загружаю зашифрованные данные по идентификатору
	}

	// DataReader - интерфейс для выгрузки данных у конкретного пользователя по его id.
	DataReader interface {
		GetAllData(ctx context.Context, idUser string) ([][]data.Data, error) // Возвращает слайс данных.
	}

	// IStorage - интерфейс хранения данных пользователей в незашифрованном виде.
	IStorage interface {
		DataWriter
		DataReader
	}
)
