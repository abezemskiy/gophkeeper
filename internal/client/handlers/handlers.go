package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/repositories/data"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// SaveEncryptedData - функция для сохранения новых данных. Новые данные сохраняются в локальном хранилище
// и происходит попытка отправки данных на сервер.
func SaveEncryptedData(serverAddress, action string, client *resty.Client, encryptedData data.EncryptedData) error {

	// сериализую зашифрованные данные в json-представление  в виде слайса байт
	var bufEncode bytes.Buffer
	enc := json.NewEncoder(&bufEncode)
	if err := enc.Encode(encryptedData); err != nil {
		logger.AgentLog.Error("Encode encrypted data error", zap.String("error", error.Error(err)))
		return fmt.Errorf("encode encrypted data error, %w", err)
	}

	// Конструирую полный адрес отправки данных
	url := fmt.Sprintf("%s/%s", serverAddress, action)

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

		return nil
	}

	// Успешная отправка данных на сервер
	if resp.StatusCode() == http.StatusOK {
		logger.AgentLog.Debug("successful pushing encrypted data to server", zap.String("data name", encryptedData.Name))

		// Сохранение данных в локальном хранилище со статусом SAVED

		return nil
	}

	// Обработка случаю, когда на сервере произошла внутренняя ошибка
	if resp.StatusCode() == http.StatusInternalServerError {
		// сохранение зашифрованных данных в локальном хранилище со статусом NEW

		return nil
	}

	// Сервер вернул иной статус
	logger.AgentLog.Error("push json encrypted to server error", zap.String("status", fmt.Sprintf("%d", resp.StatusCode())))
	return fmt.Errorf("push json encrypted to server error, status %d", resp.StatusCode())
}
