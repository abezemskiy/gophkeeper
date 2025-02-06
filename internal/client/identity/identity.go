package identity

import (
	"context"
)

// ClientIdentifier - интерфейс для реализации процедур регистрации и авторизации пользователя.
type ClientIdentifier interface {
	Register(ctx context.Context, login, hash, id, token string) (bool, error)       // Метод для регистрации пользователя.
	Authorize(ctx context.Context, login string) (data UserInfo, ok bool, err error) // Метод для авторизации пользователя.
	SetToken(ctx context.Context, login, token string) error                         // Метод для установки токена для определенного пользователя.
}

// UserInfo - структура для авторизационных данных пользователя.
type UserInfo struct {
	ID    string
	Token string
	Hash  string
}

//-----------------------------------------------------------------------------------------------------------------------------

// IUserInfoStorage - потокобезопасный интерфейс для сохранения и получения информации о пользователе.
type IUserInfoStorage interface {
	Set(authData AuthData, id string) // метод для установки данных пользователя.
	Get() (AuthData, string)           // метод для получения данных пользователя.
}

// AuthData - структура для получения и передачи идентификационных данных пользователя.
type AuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
