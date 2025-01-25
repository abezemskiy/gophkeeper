package storage

import (
	"context"
)

type (
	// EncryptedDataWriter - интерфейс для добавления зашифрованных данных хранилище.
	EncryptedDataWriter interface {
		AddEncryptedData(ctx context.Context, idUser string, uplink []byte) error // загружаю зашифрованные данные по идентификатору
	}

	// EncryptedDataReader - интерфейс для выгрузки зашифрованных данных у конкретного пользователя по его id.
	EncryptedDataReader interface {
		GetEncryptedData(ctx context.Context, idUser string) ([][]byte, error) // Возвращает слайс данных. Данные в виде зашифрованного слайса байт
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
	}
)
