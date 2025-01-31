package forbider

import (
	"fmt"
	"gophkeeper/internal/repositories/connection"
	"gophkeeper/internal/server/connection/tools/checker"
	"gophkeeper/internal/server/identity/auth"
	"gophkeeper/internal/server/logger"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// OfflineForbider - meddleware, которая запрещает действие, если пользователь был офлайн и только восстановил соединение.
func OfflineForbider(connInfo connection.ConnectionInfoKeeper) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// Получаю id пользователя из контекста
			id, ok := req.Context().Value(auth.UserIDKey).(string)
			if !ok {
				logger.ServerLog.Error("user ID not found in context", zap.String("address", req.URL.String()))
				http.Error(res, "user ID not found in context", http.StatusInternalServerError)
				return
			}

			// фиксирую текущий момент времени
			now := time.Now().Truncate(time.Second)

			// Выполняю проверку, пользователь онлайн или офлайн
			online, err := checker.IsOnline(req.Context(), id, now, connInfo)
			if err != nil {
				logger.ServerLog.Error("check network status error", zap.String("address", req.URL.String()), zap.String("error", err.Error()))
				http.Error(res, fmt.Errorf("check network status error, %w", err).Error(), http.StatusInternalServerError)
				return
			}

			// Запрещаю соединение, если пользователь был офлайн
			if !online {
				logger.ServerLog.Info("user was been offline", zap.String("address", req.URL.String()))
				http.Error(res, "user was been offline", http.StatusForbidden)
				return
			}
			// продолжаю выполнение запроса, если пользователь не терял соединение
			h.ServeHTTP(res, req)
		}
	}
}
