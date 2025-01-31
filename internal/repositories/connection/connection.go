package connection

import (
	"context"
	"time"
)

// ConnectionInfoKeeper - интерфейс для хранения и получения информации о соединении пользователя.
type ConnectionInfoKeeper interface {
	AddDateOfLastVisit(ctx context.Context, id string, date time.Time) error // Метод для добавления даты последнего визита пользователя.
	GetDateOfLastVisit(ctx context.Context, id string) (time.Time, error)    // Метод для выгрузки даты последнего визита пользователя.
}
