package storage

import (
	"context"
	"gophkeeper/internal/repositories/data"
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

	// Starter - интерфейс для инициализации хранилища.
	Starter interface {
		Bootstrap(context.Context) error
	}

	// IStorage - интерфейс хранения данных пользователей в незашифрованном виде.
	IStorage interface {
		DataWriter
		DataReader
		Starter
	}
)
