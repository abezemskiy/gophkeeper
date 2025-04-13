package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/abezemskiy/gophkeeper/internal/client/encr"
	"github.com/abezemskiy/gophkeeper/internal/client/identity"
	"github.com/abezemskiy/gophkeeper/internal/client/storage"
	"github.com/abezemskiy/gophkeeper/internal/repositories/data"
)

// updatingPeriod - период обновления расщифрованных данных пользователя в секундах.
var updatingPeriod int

// SetUpdatingPeriod - функция для установки периода обновления расшифрованных данных пользователя во временном хранилище.
func SetUpdatingPeriod(sec int) {
	updatingPeriod = sec
}

// GetUpdatingPeriod - функция для получения периода обновления расшифрованных данных пользователя.
func GetUpdatingPeriod() time.Duration {
	return time.Second * time.Duration(updatingPeriod)
}

// DecryptedData - потокобезопасное хранилище расшифрованных данных пользователя в оперативной памяти.
type DecryptedData struct {
	mu   sync.RWMutex
	data [][]data.Data
}

// Update - метод для актуализации данных пользователя.
// Актуальные данные берутся из постоянного хранилища.
func (d *DecryptedData) Update(ctx context.Context, stor storage.IEncryptedClientStorage, info identity.IUserInfoStorage) error {
	// Получаю данные пользователя
	authData, id := info.Get()

	// Извлекаю зашифрованные данные пользователя из хранилища
	encrData, err := stor.GetAllEncryptedData(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get encrypted user data from storage, %w", err)
	}

	// Переменная для сохранения расшифрованных данных
	decrData := make([][]data.Data, len(encrData))

	for i, dataEncrVirsions := range encrData {
		// переменная для хранения всех расшифрованных версий одних данных
		dataDecrVersions := make([]data.Data, len(dataEncrVirsions))

		// Итерируюсь по всем версиям одних данных
		for j, d := range dataEncrVirsions {
			// Расшифровываю данные
			decr, err := encr.DecryptData(authData.Password, &d)
			if err != nil {
				return fmt.Errorf("failed to decrypt data, %w", err)
			}
			// устанавливаю расшифрованную версию данных в переменную
			dataDecrVersions[j] = *decr
		}

		decrData[i] = dataDecrVersions
	}

	// Сохраняю расшифрованные данные в хранилище
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data = decrData

	return nil
}

// GetAll - метод для получения расшифрованных данных пользователя.
func (d *DecryptedData) GetAll() [][]data.Data {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Копирую оригинальный слайс данных для обеспечения потокобезопасности
	copiedData := make([][]data.Data, len(d.data))
	for i := range d.data {
		copiedData[i] = make([]data.Data, len(d.data[i]))
		copy(copiedData[i], d.data[i])
	}

	return copiedData
}

// NewDecryptedData - фабричная функция для создания временного хранилища расшифрованных данных пользователя.
func NewDecryptedData() *DecryptedData {
	return &DecryptedData{}
}
