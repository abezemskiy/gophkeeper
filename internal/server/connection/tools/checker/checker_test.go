package checker

import (
	"context"
	"errors"
	"gophkeeper/internal/repositories/mocks"
	"gophkeeper/internal/repositories/synchronization"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsOnline(t *testing.T) {
	// регистрирую мок хранилища данных пользователей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockConnectionInfoKeeper(ctrl)

	{
		// Пользователь онлайн
		ctx := context.Background()
		userID := "online user id"
		lastVisit := time.Now().Add(-(synchronization.GetPeroidOfSynchr() * 3) / 4)
		m.EXPECT().GetDateOfLastVisit(gomock.Any(), userID).Return(lastVisit, nil)

		res, err := IsOnline(ctx, userID, time.Now(), m)
		require.NoError(t, err)
		assert.Equal(t, true, res)
	}
	{
		// Пользователь офлайн
		ctx := context.Background()
		userID := "ofline user id"
		lastVisit := time.Now().Add(-(synchronization.GetPeroidOfSynchr() * 2))
		m.EXPECT().GetDateOfLastVisit(gomock.Any(), userID).Return(lastVisit, nil)

		res, err := IsOnline(ctx, userID, time.Now(), m)
		require.NoError(t, err)
		assert.Equal(t, false, res)
	}
	{
		// Хранилище возвращает ошибку
		ctx := context.Background()
		userID := "error user id"
		lastVisit := time.Now().Add(-(synchronization.GetPeroidOfSynchr() * 2))
		m.EXPECT().GetDateOfLastVisit(gomock.Any(), userID).Return(lastVisit, errors.New("some storage error"))

		_, err := IsOnline(ctx, userID, time.Now(), m)
		require.Error(t, err)
	}
}
