package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/repositories/data"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// SaveEncryptedData - функция для сохранения новых данных. Новые данные сохраняются в локальном хранилище
// и происходит попытка отправки данных на сервер.
func SaveEncryptedData(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage, encrData data.EncryptedData) error {

	// сериализую зашифрованные данные в json-представление  в виде слайса байт
	var bufEncode bytes.Buffer
	enc := json.NewEncoder(&bufEncode)
	if err := enc.Encode(encrData); err != nil {
		logger.AgentLog.Error("Encode encrypted data error", zap.String("error", error.Error(err)))
		return fmt.Errorf("encode encrypted data error, %w", err)
	}

	// попытка отправить новые данные на сервер
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(bufEncode).
		Post(url)

	// Не удалось установить соединение сервером или другая ошибка подобного рода.
	// Сохраняю данные в локальном хранилище. Следующая попытка сохранения данных на сервере будет
	// осуществлена во время синхронизации данных.
	if err != nil {
		logger.AgentLog.Error("push json encrypted to server error", zap.String("error", error.Error(err)))

		// сохранение зашифрованных данных в локальном хранилище со статусом NEW
		return SaveEncryptedDataToLocalStorage(ctx, userID, stor, encrData, data.NEW)
	}

	// Успешная отправка данных на сервер
	if resp.StatusCode() == http.StatusOK {
		logger.AgentLog.Debug("successful pushing encrypted data to server", zap.String("data name", encrData.Name))

		// Сохранение данных в локальном хранилище со статусом SAVED
		return SaveEncryptedDataToLocalStorage(ctx, userID, stor, encrData, data.SAVED)
	}

	// Обработка случаю, когда на сервере произошла внутренняя ошибка
	if resp.StatusCode() == http.StatusInternalServerError {
		// сохранение зашифрованных данных в локальном хранилище со статусом NEW
		return SaveEncryptedDataToLocalStorage(ctx, userID, stor, encrData, data.NEW)
	}

	// Сервер вернул иной статус
	logger.AgentLog.Error("push json encrypted to server error", zap.String("status", fmt.Sprintf("%d", resp.StatusCode())))
	return fmt.Errorf("push json encrypted to server error, status %d", resp.StatusCode())
}

func SaveEncryptedDataToLocalStorage(ctx context.Context, userID string, stor storage.IEncryptedClientStorage, encrData data.EncryptedData, status int) error {
	ok, err := stor.AddEncryptedData(ctx, userID, encrData, status)
	if err != nil {
		logger.AgentLog.Error("failed to save new encrypted data to storage", zap.String("error", error.Error(err)))
		return fmt.Errorf("failed to save new encrypted data to storage with error, %w", err)
	}
	// Данные уже существуют в локальном хранилище
	if !ok {
		logger.AgentLog.Error("failed to save new encrypted data to storage", zap.String("reason", "data is already exist"))
		return fmt.Errorf("data is already exist")
	}
	return nil
}
