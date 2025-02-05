package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/common/identity/tools/checker"
	"gophkeeper/internal/common/identity/tools/hasher"
	"gophkeeper/internal/common/identity/tools/header"
	"gophkeeper/internal/common/identity/tools/id"
	"gophkeeper/internal/repositories/data"
	repoIdent "gophkeeper/internal/repositories/identity"
	"net/http"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func SaveEncryptedDataToLocalStorage(ctx context.Context, userID string, stor storage.IEncryptedClientStorage,
	encrData data.EncryptedData, status int) (bool, error) {

	ok, err := stor.AddEncryptedData(ctx, userID, encrData, status)
	if err != nil {
		logger.ClientLog.Error("failed to save new encrypted data to storage", zap.String("error", error.Error(err)))
		return false, fmt.Errorf("failed to save new encrypted data to storage with error, %w", err)
	}
	// Данные уже существуют в локальном хранилище
	if !ok {
		logger.ClientLog.Error("failed to save new encrypted data to storage", zap.String("reason", "data is already exist"))
		return false, nil
	}
	return true, nil
}

// SaveEncryptedData - функция для сохранения новых зашифрованных данных. Новые зашифрованные данные сохраняются в локальном хранилище
// и происходит попытка отправки данных на сервер.
func SaveEncryptedData(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	encrData *data.EncryptedData) (bool, error) {

	// сериализую зашифрованные данные в json-представление  в виде слайса байт
	var bufEncode bytes.Buffer
	enc := json.NewEncoder(&bufEncode)
	if err := enc.Encode(encrData); err != nil {
		logger.ClientLog.Error("Encode encrypted data error", zap.String("error", error.Error(err)))
		return false, fmt.Errorf("encode encrypted data error, %w", err)
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
		logger.ClientLog.Error("push json encrypted to server error", zap.String("error", error.Error(err)))

		// сохранение зашифрованных данных в локальном хранилище со статусом NEW
		return SaveEncryptedDataToLocalStorage(ctx, userID, stor, *encrData, data.NEW)
	}

	// Успешная отправка данных на сервер
	if resp.StatusCode() == http.StatusOK {
		logger.ClientLog.Debug("successful pushing encrypted data to server", zap.String("data name", encrData.Name))

		// Сохранение данных в локальном хранилище со статусом SAVED
		return SaveEncryptedDataToLocalStorage(ctx, userID, stor, *encrData, data.SAVED)
	}

	// Обработка случаю, когда на сервере произошла внутренняя ошибка
	if resp.StatusCode() == http.StatusInternalServerError {
		// сохранение зашифрованных данных в локальном хранилище со статусом NEW
		return SaveEncryptedDataToLocalStorage(ctx, userID, stor, *encrData, data.NEW)
	}

	// Сервер вернул иной статус
	logger.ClientLog.Error("push json encrypted to server error", zap.String("status", fmt.Sprintf("%d", resp.StatusCode())))
	return false, fmt.Errorf("push json encrypted to server error, status %d", resp.StatusCode())
}

// Register - хэндлер для регистрации нового пользователя.
func Register(ctx context.Context, url string, authData *identity.AuthData, client *resty.Client, ident identity.ClientIdentifier) (bool, error) {
	// проверяю корректность логина
	ok := checker.CheckLogin(authData.Login)
	if !ok {
		return false, fmt.Errorf("login is not valid")
	}

	// проверяю корректность пароля
	ok = checker.CheckPassword(authData.Password)
	if !ok {
		return false, fmt.Errorf("password is not valid")
	}

	// вычисляю хэш на основе логина и пароля для последующей аутентификации
	hash, err := hasher.CalkHash(authData.Login + authData.Password)
	if err != nil {
		logger.ClientLog.Error("failed to calculate hash", zap.String("address", url), zap.String("error", error.Error(err)))
		return false, fmt.Errorf("failed to calculate hash, %w", err)
	}

	// Создаю тело запроса
	regData := repoIdent.IdentityData{
		Login: authData.Login,
		Hash:  hash,
	}

	// Отправляю запрос регистрации пользователя на сервер
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(regData).
		Post(url)

	// Не удалось установить соединение с сервером или другая ошибка подобного рода.
	// Регистрация пользователя в состоянии офлайн запрещена, возвращаю ошибку.
	if err != nil {
		logger.ClientLog.Error("sending registration request failed", zap.String("error", error.Error(err)))
		return false, fmt.Errorf("sending registration request failed, %w", err)
	}

	// Сервер успешно зарегистрировал нового пользователя
	if resp.StatusCode() == http.StatusOK {
		logger.ClientLog.Debug("successful register new user on server")

		// вычисляю уникальный идентификатор пользователя. Идентификатора пользователя отличаются на сервере и в локальном хранилище.
		id, err := id.GenerateId()
		if err != nil {
			logger.ClientLog.Error("failed to generate id", zap.String("error", error.Error(err)))
			return false, fmt.Errorf("failed to generate id, %w", err)
		}

		// Получаю токен из заголовка, который отправил сервер.
		token, err := header.GetTokenFromRestyResponseHeader(resp)
		if err != nil {
			logger.ClientLog.Error("failed to get JWT from server responce", zap.String("error", error.Error(err)))
			return false, fmt.Errorf("failed to get JWT from server responce, %w", err)
		}

		// Сохраняю данные пользователя в локальном хранилище
		ok, err := ident.Register(ctx, authData.Login, hash, id, token)
		if err != nil {
			logger.ClientLog.Error("failed to register user in local storage", zap.String("error", error.Error(err)))
			return false, fmt.Errorf("failed to register user in local storage, %w", err)
		}
		// Пользователь уже зарегистрирован в хранилище
		if !ok {
			logger.ClientLog.Error("such user already exist", zap.String("login", authData.Login))
			return false, nil
		}

		// Успешная регистрация нового пользователя
		logger.ClientLog.Info("new user successfully has been registered", zap.String("login", authData.Login))
		return true, nil
	}

	// пользователь с такими данными уже зарегистрирован, возвращаю false.
	if resp.StatusCode() == http.StatusConflict {
		logger.ClientLog.Error("such user already exist", zap.String("login", authData.Login))
		return false, nil
	}

	logger.ClientLog.Error("bad server status", zap.String("status", fmt.Sprint(resp.StatusCode())))
	return false, fmt.Errorf("bad server status %d", resp.StatusCode())
}
