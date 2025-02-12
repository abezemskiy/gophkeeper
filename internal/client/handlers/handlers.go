package handlers

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/encr"
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
	"strconv"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// SaveEncryptedDataToLocalStorage - функция для сохранения данных в локальном хранилище.
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

	logger.ClientLog.Debug("successful save encrypted data in local storage", zap.String("data name", encrData.Name))
	return true, nil
}

// SaveEncryptedData - функция для сохранения новых зашифрованных данных. Новые зашифрованные данные сохраняются в локальном хранилище
// и происходит попытка отправки данных на сервер.
func SaveEncryptedData(ctx context.Context, userID, url string, client *resty.Client, stor storage.IEncryptedClientStorage,
	encrData *data.EncryptedData) (bool, error) {

	// попытка отправить новые данные на сервер
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(*encrData).
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
		logger.ClientLog.Error("push json encrypted to server error", zap.String("status", fmt.Sprintf("%d", resp.StatusCode())))

		// сохранение зашифрованных данных в локальном хранилище со статусом NEW
		return SaveEncryptedDataToLocalStorage(ctx, userID, stor, *encrData, data.NEW)
	}

	// Сервер вернул иной статус
	logger.ClientLog.Error("push json encrypted to server error", zap.String("status", fmt.Sprintf("%d", resp.StatusCode())))
	return false, fmt.Errorf("push json encrypted to server error, status %d", resp.StatusCode())
}

// SaveData - функция для сохранения новых данных. Новые данные зашифровываются с помощью мастер пароля, сохраняются в локальном хранилище
// и происходит попытка отправки данных на сервер.
func SaveData(ctx context.Context, userID, url, masterPass string, client *resty.Client, stor storage.IEncryptedClientStorage,
	userData *data.Data) (bool, error) {

	// шифрую данные с помощью мастер пароля пользователя
	encrData, err := encr.EncryptData(masterPass, userData)
	if err != nil {
		logger.ClientLog.Error("failed to encrypt data", zap.String("error", error.Error(err)))
		return false, fmt.Errorf("failed to encrypt data, %w", err)
	}

	// Сохраняю данные в хранилище
	ok, err := SaveEncryptedData(ctx, userID, url, client, stor, encrData)
	if err != nil {
		return false, fmt.Errorf("failed to save data, %w", err)
	}
	if !ok {
		return false, nil
	}
	return true, nil
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
	regData := repoIdent.Data{
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

	// пользователь с такими данными уже зарегистрирован, возвращаю false.
	if resp.StatusCode() == http.StatusConflict {
		logger.ClientLog.Error("such user already exist", zap.String("login", authData.Login))
		return false, nil
	}

	if resp.StatusCode() != http.StatusOK {
		logger.ClientLog.Error("bad server status", zap.String("status", fmt.Sprint(resp.StatusCode())))
		return false, fmt.Errorf("bad server status %d", resp.StatusCode())
	}

	// Сервер успешно обработал запрос пользователя на регистрацию
	logger.ClientLog.Debug("successful register new user on server")

	// вычисляю уникальный идентификатор пользователя. Идентификатора пользователя отличаются на сервере и в локальном хранилище.
	id, err := id.GenerateID()
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
	ok, err = ident.Register(ctx, authData.Login, hash, id, token)
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

// Authorize - хэндлер для авторизации пользователя в системе.
// passIsCorrect - обозначает введен ли верный пароль,
// registered - зарегистрирован ли пользователь в системе.
// После успешной авторизации данные пользователя устанавливаются в хранилище на время сесси для использования в других методах.
func Authorize(ctx context.Context, authData *identity.AuthData, ident identity.ClientIdentifier,
	info identity.IUserInfoStorage) (passIsCorrect bool, registered bool, err error) {
	// проверяю корректность логина
	ok := checker.CheckLogin(authData.Login)
	if !ok {
		return false, false, fmt.Errorf("login is not valid")
	}

	// проверяю корректность пароля
	ok = checker.CheckPassword(authData.Password)
	if !ok {
		return false, false, fmt.Errorf("password is not valid")
	}

	// Извлекаю данные пользователя из хранилища
	userInfo, ok, err := ident.Authorize(ctx, authData.Login)
	if err != nil {
		logger.ClientLog.Error("failed to getting user info from storage", zap.String("error", error.Error(err)))
		return false, false, fmt.Errorf("failed to getting user info from storage, %w", err)
	}
	// Пользователь не зарегистрирован
	if !ok {
		logger.ClientLog.Error("user not register", zap.String("login", authData.Login))
		return false, false, nil
	}

	// Вычисляю хэш от пары логин пароль для сверки с тем, что содержится в хранилище
	hash, err := hasher.CalkHash(authData.Login + authData.Password)
	if err != nil {
		logger.ClientLog.Error("failed to calculate hash", zap.String("error", error.Error(err)))
		return false, false, fmt.Errorf("failed to calculate hash, %w", err)
	}

	// Если хэш полученный из хранилища не совпадает с тем, что был расчитан из полученной пары логи-пароль,
	// то пароль неверный.
	if hash != userInfo.Hash {
		logger.ClientLog.Error("wrong password", zap.String("login", authData.Login))
		return false, true, nil
	}

	// Устанавливаю данные пользователя в хранилище
	info.Set(*authData, userInfo.ID)

	// Пользователь успешно авторизирован
	logger.ClientLog.Info("user successfully authorize", zap.String("login", authData.Login))
	return true, true, nil
}

// DeleteEncryptedDataFromLocalStorage - функция для удаления данных пользователя в локальном хранилище.
func DeleteEncryptedDataFromLocalStorage(ctx context.Context, userID, dataName string, stor storage.IEncryptedClientStorage) (bool, error) {

	ok, err := stor.DeleteEncryptedData(ctx, userID, dataName)
	if err != nil {
		logger.ClientLog.Error("failed to delete data from local storage", zap.String("error", error.Error(err)), zap.String("data name", dataName))
		return false, fmt.Errorf("failed to delete data from local storage, %w", err)
	}
	// Данных не существует
	if !ok {
		logger.ClientLog.Error("failed to delete data from local storage", zap.String("reason", "data does not exists"))
		return false, nil
	}

	logger.ClientLog.Debug("successful delete data from local storage", zap.String("data name", dataName))
	return true, nil
}

// DeleteEncryptedData - хэндлер для удаления данных пользователя на сервере и из локального хранилища по имени этих данных.
// Удаление данных разрешено только в статусе онлайн.
func DeleteEncryptedData(ctx context.Context, userID, url, dataName string, client *resty.Client, stor storage.IEncryptedClientStorage) (bool, error) {

	// попытка удалить данные пользователя на сервере
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data.MetaInfo{
			Name: dataName,
		}).
		Delete(url)

	// Не удалось установить соединение сервером или другая ошибка подобного рода.
	// Удаление данных в состоянии офлайн запрещено, возвращаю ошибку.
	if err != nil {
		logger.ClientLog.Error("delete data on server error", zap.String("error", error.Error(err)), zap.String("data name", dataName))

		return false, fmt.Errorf("delete data on server error, %w", err)
	}

	// В случае, если данные на сервере успешно удалены, либо данных уже не было на сервере произвожу удаление в локальном хранилище.
	if resp.StatusCode() == http.StatusOK || resp.StatusCode() == http.StatusNotFound {
		logger.ClientLog.Debug("successful delete data from server", zap.String("data name", dataName))

		// Удаляю данные из локального хранилища
		return DeleteEncryptedDataFromLocalStorage(ctx, userID, dataName, stor)
	}

	// Сервер вернул иной статус
	logger.ClientLog.Error("failed to delete data from server", zap.String("status", strconv.Itoa(resp.StatusCode())), zap.String("data name", dataName))
	return false, fmt.Errorf("failed to delete data from server with status %d", resp.StatusCode())
}
