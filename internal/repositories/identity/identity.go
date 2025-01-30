package identity

import "context"

// Identifier - интерфейс для реализации процедур регистрации и авторизации пользователя.
type Identifier interface {
	Register(ctx context.Context, login, hash, id string) error                               // Метод для регистрации пользователя.
	Authorize(ctx context.Context, login string) (data AuthorizationData, ok bool, err error) // Метод для авторизации пользователя.
}

// IdentityData - структура данных для аутентификации пользователя.
type IdentityData struct {
	Login string `json:"login"` // логин пользователя
	Hash  string `json:"hash"`  //  хэш от суммы логин+пароль
}

// AuthorizationData - структуоа для авторизационных данных пользователя.
type AuthorizationData struct {
	Hash string
	ID   string
}
