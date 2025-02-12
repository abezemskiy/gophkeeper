package id

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateID - функция для гененерации ID пользователя. В качестве id будет использоваться сгенерированный UUID (Universally Unique Identifier)
func GenerateID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate id, %w", err)
	}
	return id.String(), nil
}
