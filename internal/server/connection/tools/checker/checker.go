package checker

import (
	"context"
	"fmt"
	"gophkeeper/internal/repositories/connection"
	"gophkeeper/internal/repositories/synchronization"
	"time"
)

// IsOnline - функция, которая проверяет, находится ли пользователь online.
func IsOnline(ctx context.Context, userID string, now time.Time, connInfo connection.ConnectionInfoKeeper) (bool, error) {
	// дата последнего подключения
	getLastVisit, err := connInfo.GetDateOfLastVisit(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get connection info from storage, %W", err)
	}
	// промежуток времени между последним подключением и настоящим моментом
	duration := now.Sub(getLastVisit)

	// Добавляю пять секунд к периоду синхронизации для компенсации проблем с подключением.
	period := synchronization.GetPeroidOfSynchr() + time.Second*5
	return duration < period, nil
}
