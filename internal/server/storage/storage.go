package storage

import (
	"context"

	"github.com/abezemskiy/gophkeeper/internal/repositories/data"
	repoStorage "github.com/abezemskiy/gophkeeper/internal/repositories/storage"
)

// Интерфейсы для хранения зашифрованных данных пользователей на сервере.
type (
	// EncryptedDataAppender - интерфейс для сохранения дополнительной версии существующих данных в случае конфликта.
	EncryptedDataAppender interface {
		AppendEncryptedData(ctx context.Context, idUser string, data data.EncryptedData) (bool, error) // Для добавления зашифрованныч данных по id
	}

	// IEncryptedServerStorage - интерфейс сервера для хранения зашифрованных данных пользователей.
	IEncryptedServerStorage interface {
		repoStorage.IEncryptedStorage
		EncryptedDataAppender
	}
)
