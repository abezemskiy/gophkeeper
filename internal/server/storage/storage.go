package storage

import (
	"context"
	"gophkeeper/internal/repositories/data"
	repoStorage "gophkeeper/internal/repositories/storage"
)

type (
	// EncryptedDataAppender - интерфейс для сохранения дополнительной версии существующих данных в случае конфликта.
	EncryptedDataAppender interface {
		AppendEncryptedData(ctx context.Context, idUser, data data.EncryptedData) (bool, error) // Для добавления зашифрованныч данных по id
	}

	// IEncryptedServerStorage - интерфейс сервера для хранения зашифрованных данных пользователей.
	IEncryptedServerStorage interface {
		repoStorage.IEncryptedStorage
		EncryptedDataAppender
	}
)
