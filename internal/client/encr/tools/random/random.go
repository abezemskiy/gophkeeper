package random

import "crypto/rand"

// GenerateCryptoRandom - функция для генерации криптостойкой случайной последовательности байт длины size.
func GenerateCryptoRandom(size int) ([]byte, error) {
	// генерируем криптостойкие случайные байты в b
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
