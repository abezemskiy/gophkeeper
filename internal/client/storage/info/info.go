package info

import (
	"sync"

	"github.com/abezemskiy/gophkeeper/internal/client/identity"
)

// UserInfoStorage - потокобезопасная структура для хранения информации о пользователе (логин, мастер пароль, id) в оперативной памяти.
// Предоставляет методы для потокобезопасного использования.
type UserInfoStorage struct {
	mu       sync.RWMutex
	authData identity.AuthData
	id       string
}

// Установка информации о пользователе.
func (s *UserInfoStorage) Set(authData identity.AuthData, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authData = authData
	s.id = id
}

// Получение информации о пользователе.
func (s *UserInfoStorage) Get() (identity.AuthData, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.authData, s.id
}

// NewUserInfoStorage - фабричная функция структуры хранения информации пользователя UserInfoStorage.
func NewUserInfoStorage() *UserInfoStorage {
	return &UserInfoStorage{}
}
