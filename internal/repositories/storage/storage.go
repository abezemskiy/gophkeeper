package storage

import (
	"context"
	"gophkeeper/internal/repositories/data"
)

type (
	// EncryptedDataWriter - интерфейс для добавления зашифрованных данных хранилище.
	EncryptedDataWriter interface {
		AddEncryptedData(ctx context.Context, idUser string, data data.EncryptedData) (bool, error)     // Для загрузки зашифрованныч данных по id
		ReplaceEncryptedData(ctx context.Context, idUser string, data data.EncryptedData) (bool, error) // Для замены существующих в хранилищеданные
	}

	// EncryptedDataReader - интерфейс для выгрузки зашифрованных данных у конкретного пользователя по его id.
	EncryptedDataReader interface {
		GetAllEncryptedData(ctx context.Context, idUser string) ([][]data.EncryptedData, error) // Возвращает все зашифрованные данные по id
	}

	// EncryptedDataDeleter - интерфейс для удаления зашифрованных данных по id пользователя и имени данных.
	EncryptedDataDeleter interface {
		DeleteEncryptedData(ctx context.Context, idUser, dataName string) (bool, error)
	}

	// Starter - интерфейс для инициализации хранилища с зашифрованными данными.
	Starter interface {
		Bootstrap(context.Context) error
	}

	// IEncryptedStorage - интерфейс хранения зашифрованных данных пользователей.
	IEncryptedStorage interface {
		EncryptedDataWriter
		EncryptedDataReader
		Starter
		EncryptedDataDeleter
	}
)
