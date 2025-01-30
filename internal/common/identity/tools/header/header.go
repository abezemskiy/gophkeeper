package header

import (
	"fmt"
	"net/http"
	"strings"
)

// GetTokenFromHeader - функция для получения токена из заголовка запроса.
func GetTokenFromHeader(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Проверяю, что заголовок начинается с "Bearer "
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	jwtToken := parts[1]
	return jwtToken, nil
}

// GetTokenFromResponseHeader извлекает JWT-токен из заголовка в ответе сервера
// необходима для тестирования хэндлеров сервера. Имитирую работу клиента и получение им токена из заголовка.
func GetTokenFromResponseHeader(res *http.Response) (string, error) {
	authHeader := res.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Проверяю, что заголовок начинается с "Bearer "
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	jwtToken := parts[1]
	return jwtToken, nil
}
