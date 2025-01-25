package storage

import "context"

type (
	// DataWriter - интерфейс для добавления данных хранилище.
	DataWriter interface {
		AddData(ctx context.Context, idUser string, uplink []byte) error // загружаю зашифрованные данные по идентификатору
	}

	// DataReader - интерфейс для выгрузки данных у конкретного пользователя по его id.
	DataReader interface {
		GetData(ctx context.Context, idUser string) ([][]byte, error) // Возвращает слайс данных. Данные в виде зашифрованного слайса байт
	}

	// Starter - интерфейс для инициализации хранилища.
	Starter interface {
		Bootstrap(context.Context) error
	}

	// IStorage - интерфейс хранения uplink сообщений с LoRaWAN сервера.
	IStorage interface {
		DataWriter
		DataReader
		Starter
	}
)
