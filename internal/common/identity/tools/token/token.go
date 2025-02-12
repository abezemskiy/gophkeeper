package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Секретный ключа для генерации JWT.
var secretKey string

// SetSecretKey - функция для установки секретного ключа для генерации JWT.
func SetSecretKey(newKey string) {
	secretKey = newKey
}

// expireHour - время действия токена в часах.
var expireHour int

// SerExpireHour - функция, для установки времени действия токена в часах.
func SerExpireHour(expire int) {
	expireHour = expire
}

// Claims - структура утверждений, которая включает стандартные утверждения
// и одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// BuildJWT - создает токен и возвращает его в виде строки.
func BuildJWT(userID string) (string, error) {
	// создаю токен с алгоритмом подписи HS256 и утверждениями - Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// дата истечения токена
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHour))),
		},
		// собственное утверждение - идентификатор пользователя
		UserID: userID,
	})

	// создаю строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to signed JWT to string, %w", err)
	}
	return tokenString, nil
}

// GetIDFromToken - функция для получения id пользователя из токена с проверкой заголовка алгоритма токена.
// Заголовок должен совпадать с тем, который сервер использует для подписи и проверки токенов.
func GetIDFromToken(tokenStr string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}
