package synchronization

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/repositories/data"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// SynchronizeNewLocalData - функция для сохранения локальных данных со статусом NEW на сервере.
// URL представляет собой адрес до хэндлера сервера для добавления новых данных.
func SynchronizeNewLocalData(ctx context.Context, stor storage.IEncryptedClientStorage, info identity.IUserInfoStorage,
	client *resty.Client, url string) error {
	// Извлекаю данные пользователя
	authData, id := info.Get()

	// Извлекаю локальные изменения пользователя, которые не были сохранены на сервере и произвожу повторную попытку их сохранения.
	encrData, err := stor.GetEncryptedDataByStatus(ctx, id, data.NEW)
	if err != nil {
		return fmt.Errorf("failed to get encrypted data from storage with status NEW of user %s, %w", authData.Login, err)
	}

	// Итерируюсь по слайсу зашифрованных данных
	for _, d := range encrData {
		if len(d) == 0 {
			return fmt.Errorf("no version of data with status NEW exist")
		}
		// Убеждаюсь, что существует лишь единственная версия данных со статусом NEW
		if len(d) > 1 {
			return fmt.Errorf("only one version of data with status NEW can be exists")
		}

		// Отправляю данные на сервер
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(d[0]).
			Post(url)

		if err != nil {
			return fmt.Errorf("failed to post request to server for adding new client data, %w", err)
		}

		// Обновляю статус данных в хранилище --------------------
		var newStatus int
		if resp.StatusCode() == http.StatusConflict {
			newStatus = data.CHANGE
		} else if resp.StatusCode() == http.StatusOK {
			newStatus = data.SAVED
		}

		if resp.StatusCode() == http.StatusConflict || resp.StatusCode() == http.StatusOK {
			ok, err := stor.ChangeStatusOfEncryptedData(ctx, id, d[0].Name, newStatus)
			if err != nil {
				return fmt.Errorf("failed to change status from data %s of user %s, %w", authData.Login, d[0].Name, err)
			}
			if !ok {
				return fmt.Errorf("user %s or data %s not exist", authData.Login, d[0].Name)
			}
			// В случае корректной обработки запроса продолжанию отправку данных на сервер
			continue
		}

		return fmt.Errorf("failed to save data in server with status %d", resp.StatusCode())
	}
	return nil
}
