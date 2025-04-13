package storage

import (
	"context"

	"github.com/abezemskiy/gophkeeper/internal/repositories/data"
)

// Интерфейсы для хранения данных пользователей.
type (
	// EncryptedDataWriter - интерфейс для добавления зашифрованных данных хранилище.
	EncryptedDataWriter interface {
		AddEncryptedData(ctx context.Context, idUser string, data data.EncryptedData, status int) (bool, error)     // Для загрузки зашифрованныч данных по id
		ReplaceEncryptedData(ctx context.Context, idUser string, data data.EncryptedData, status int) (bool, error) // Для замены существующих в хранилищеданные
	}

	// EncryptedDataReader - интерфейс для выгрузки зашифрованных данных у конкретного пользователя по его id.
	EncryptedDataReader interface {
		GetAllEncryptedData(ctx context.Context, idUser string) ([][]data.EncryptedData, error) // Возвращает все зашифрованные данные по id
	}

	// EncryptedDataDeleter - интерфейс для удаления зашифрованных данных по id пользователя и имени данных.
	EncryptedDataDeleter interface {
		DeleteEncryptedData(ctx context.Context, idUser, dataName string) (bool, error)
	}

	// IEncryptedStorage - интерфейс хранения зашифрованных данных пользователей.
	IEncryptedStorage interface {
		EncryptedDataWriter
		EncryptedDataReader
		EncryptedDataDeleter
	}
)
