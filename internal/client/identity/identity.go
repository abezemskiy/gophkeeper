package identity

import (
	"context"
	"sync"
)

// ClientIdentifier - интерфейс для реализации процедур регистрации и авторизации пользователя.
type ClientIdentifier interface {
	Register(ctx context.Context, login, hash, id, token string) (bool, error)                 // Метод для регистрации пользователя.
	Authorize(ctx context.Context, login, password string) (data UserInfo, ok bool, err error) // Метод для авторизации пользователя.
	SetToken(ctx context.Context, login, token string) error                                   // Метод для установки токена для определенного пользователя.
}

// UserInfo - структура для авторизационных данных пользователя.
type UserInfo struct {
	ID    string
	Token string
}

//-----------------------------------------------------------------------------------------------------------------------------

// PasswordStorage - структура для хранения мастер пароля в оперативной памяти.
// Предоставляет методы для потокобезопасного использования.
type PasswordStorage struct {
	mu       sync.RWMutex
	password string
}

// Установка мастер-пароля
func (s *PasswordStorage) Set(password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.password = password
}

// Получение мастер-пароля
func (s *PasswordStorage) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.password
}

// IPasswordStorage - интерфейс для сохранения и получения мастер пароля из оперативной памяти.
type IPasswordStorage interface {
	Set(string)  // метод для установки мастер пароля.
	Get() string // метод для получения мастер пароля.
}

// AuthData - структура для получения и передачи идентификационных данных пользователя.
type AuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
