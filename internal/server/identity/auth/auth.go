// auth - пакет, который реализует middleware для аутентификации пользователя.
package auth

import (
	"context"
	"fmt"
	"gophkeeper/internal/common/identity/tools/header"
	"gophkeeper/internal/common/identity/tools/token"
	"gophkeeper/internal/server/logger"
	"net/http"

	"go.uber.org/zap"
)

type contextKey string

// UserIDKey - ключ для установки ID пользователя в контекст.
const UserIDKey = contextKey("userID")

// Middleware - проверяет JWT входящих запросов к серверу.
// Позволит установить доступ к ресурсам только для аутентифицированных пользователей.
// Из полученного токена извлекается ID пользователя и устанавливается в контекст.
func Middleware(h http.Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		getToken, err := header.GetTokenFromHeader(req)
		// В случае ошибки получения токена возвращаю статус 401 - пользователь не аутентифицирован.
		if err != nil {
			logger.ServerLog.Error("failed to get token from request", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
			http.Error(res, fmt.Errorf("failed to get token from request, %w", err).Error(), http.StatusUnauthorized)
			return
		}
		id, err := token.GetIDFromToken(getToken)
		if err != nil {
			logger.ServerLog.Error("failed to get user id from token", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
			http.Error(res, fmt.Errorf("failed to get user id from token, %w", err).Error(), http.StatusUnauthorized)
			return
		}

		// В случае успешного получения id пользователя устанавливаю идентификатор в контекст для дальнейшей обработки.
		ctx := context.WithValue(req.Context(), UserIDKey, id)

		// вызываю основной обработчик
		h.ServeHTTP(res, req.WithContext(ctx))
	}
}
