package synchronization

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/repositories/data"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
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
			return fmt.Errorf("no version of data with status NEW exists")
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
			newStatus = data.CHANGED
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

// SynchronizeChangedLocalData - функция для сохранения локальных данных со статусом CHANGED на сервере.
// URL представляет собой адрес до хэндлера сервера для созранения дополнительной версии уже существующих данных.
func SynchronizeChangedLocalData(ctx context.Context, stor storage.IEncryptedClientStorage, info identity.IUserInfoStorage,
	client *resty.Client, url string) error {
	// Извлекаю данные пользователя
	authData, id := info.Get()

	// Извлекаю локальные изменения пользователя, которые не были сохранены на сервере и произвожу повторную попытку их сохранения.
	encrData, err := stor.GetEncryptedDataByStatus(ctx, id, data.CHANGED)
	if err != nil {
		return fmt.Errorf("failed to get encrypted data from storage with status CHANGED of user %s, %w", authData.Login, err)
	}

	// Итерируюсь по слайсу зашифрованных данных
	for _, d := range encrData {
		if len(d) == 0 {
			return fmt.Errorf("no version of data with status CHANGED exists")
		}
		// Убеждаюсь, что существует лишь единственная версия данных со статусом CHANGED
		if len(d) > 1 {
			return fmt.Errorf("only one version of data with status CHANGED can be exists")
		}

		// Отправляю данные на сервер
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(d[0]).
			Post(url)

		if err != nil {
			return fmt.Errorf("failed to post request to server for adding changed client data, %w", err)
		}

		// Обновляю статус данных в хранилище --------------------
		if resp.StatusCode() == http.StatusOK {
			ok, err := stor.ChangeStatusOfEncryptedData(ctx, id, d[0].Name, data.SAVED)
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

// SynchronizeDataFromServer - функция для сохранения локальных данных со статусом CHANGED на сервере.
// URL представляет собой адрес до хэндлера сервера для созранения дополнительной версии уже существующих данных.
func SynchronizeDataFromServer(ctx context.Context, stor storage.IEncryptedClientStorage, info identity.IUserInfoStorage,
	client *resty.Client, url string) error {
	// Извлекаю данные текущего пользователя
	authData, id := info.Get()

	// Отправляю запрос на сервер для получения актуальных данных пользователя
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return fmt.Errorf("failed to post request to server for adding changed client data, %w", err)
	}
	// В случае, если сервер не обработал запрос со статусом 200 возвращаю ошибку
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to get actual user data from server with status %d", resp.StatusCode())
	}

	// Создаю переменную для хранения актуальных данных пользователя полученных от сервера
	dataFromServer := make([][]data.EncryptedData, 0)

	// Декодирую ответ сервера в структуру
	if err := json.NewDecoder(bytes.NewReader(resp.Body())).Decode(&dataFromServer); err != nil {
		return fmt.Errorf("failed to decode server answer from json, %w", err)
	}

	// Итерируюсь по полученным данным в случае успешного получения данных от сервера
	for _, d := range dataFromServer {
		if len(d) == 0 {
			return fmt.Errorf("no version of data exists")
		}

		// Попытка заменить старую версию данных в локальном хранилище на актуальную, полученную от сервера
		status := data.SAVED
		// Если существует несколько версий данных, то устанавливаю статус CONFLICT
		if len(d) > 1 {
			logger.ClientLog.Debug("user got data with multiply version", zap.String("login", authData.Login),
				zap.String("data name", d[0].Name))
			status = data.CONFLICT
		}

		// Меняю существующие данные в хранилище на актуальную версию сервера
		ok, err := stor.ReplaceDataWithMultiVersionData(ctx, id, d, status)
		if err != nil {
			return fmt.Errorf("failed to replace data in storage, %w", err)
		}

		// Происходит попытка заменить данные, которых нет в хранилище.
		// В таком случае, произвожу попытку добавить новые данные.
		if !ok {
			logger.ClientLog.Info("attempting to change not existing data", zap.String("login", authData.Login),
				zap.String("data name", d[0].Name))
			ok, err := stor.AddEncryptedData(ctx, id, d[0], data.SAVED)
			if err != nil || !ok {
				return fmt.Errorf("failed to add new data %s in storage, %w", d[0].Name, err)
			}
		}
	}
	return nil
}
