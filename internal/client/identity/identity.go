package identity

import "sync"

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
