package pg

import (
	"context"
	"database/sql"
	"errors"
	"gophkeeper/internal/repositories/data"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDatabaseName = "TEST_GOPHKEEPER_DATABASE_URL"

// Вспомогательная функция для очистки данных в базе
func cleanBD(t *testing.T, dsn string, stor *Store) {
	conn, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer conn.Close()

	// проверка соединения с БД
	ctx := context.Background()
	err = conn.PingContext(ctx)
	require.NoError(t, err)

	// Вызываю метод для очистки данных в хранилище
	err = stor.Disable(ctx)
	require.NoError(t, err)
}

func TestRegister(t *testing.T) {
	// беру адрес тестовой БД из переменной окружения
	databaseDsn := os.Getenv(envDatabaseName)
	assert.NotEqual(t, "", databaseDsn)

	// создаю соединение с базой данных
	conn, err := sql.Open("pgx", databaseDsn)
	require.NoError(t, err)
	defer conn.Close()

	// Проверка соединения с БД
	ctx := context.Background()
	err = conn.PingContext(ctx)
	require.NoError(t, err)

	// создаю экземпляр хранилища
	stor := NewStore(conn)
	err = stor.Bootstrap(ctx)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	// Попытка зарегистрировать пользователя когда контекст уже завершен
	{
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := stor.Register(ctx, "login", "hash", "id")
		require.Error(t, err)
	}
	// Попытка повторно зарегистрировать пользователя
	{
		ctx := context.Background()
		err := stor.Register(ctx, "login", "hash", "id")
		require.NoError(t, err)

		err = stor.Register(ctx, "login", "new hash", "new id")
		require.Error(t, err)

		// проверяю, что полученная ошибка соответствует ошибке при попытке установить повторяющеся поле типа "PRIMARY KEY"
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			require.NoError(t, nil)
		} else {
			require.Error(t, nil)
		}
	}

}

func TestAuthorize(t *testing.T) {
	// беру адрес тестовой БД из переменной окружения
	databaseDsn := os.Getenv(envDatabaseName)
	assert.NotEqual(t, "", databaseDsn)

	// создаю соединение с базой данных
	conn, err := sql.Open("pgx", databaseDsn)
	require.NoError(t, err)
	defer conn.Close()

	// Проверка соединения с БД
	ctx := context.Background()
	err = conn.PingContext(ctx)
	require.NoError(t, err)

	// создаю экземпляр хранилища
	stor := NewStore(conn)
	err = stor.Bootstrap(ctx)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	{
		// Test. successful authorization--------------------------------
		// регистрирую пользователя
		sLogin := "login"
		sHash := "hash"
		sId := "id"
		err = stor.Register(ctx, sLogin, sHash, sId)
		require.NoError(t, err)

		// получаю данные пользователя для авторизации по его логину
		//var data identity.AuthorizationData
		data, ok, err := stor.Authorize(ctx, sLogin)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		assert.Equal(t, sHash, data.Hash)
		assert.Equal(t, sId, data.ID)
	}
	{
		// Test. context is exceeded--------------------------------
		// регистрирую пользователя
		sLogin := "exceeded login"
		sHash := "hash"
		sId := "id"
		err = stor.Register(ctx, sLogin, sHash, sId)
		require.NoError(t, err)

		ctxExc, cancel := context.WithCancel(context.Background())
		cancel()

		// попытка получить данные пользователя для авторизации по его логину, хотя контекст уже отменен.
		_, _, err := stor.Authorize(ctxExc, sLogin)
		require.Error(t, err)
	}
	{
		// Test. error authorization. User not register --------------------------------
		// пытаюсь получить данные пользователя для авторизации по его логину
		_, ok, err := stor.Authorize(ctx, "not register user")
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
}

func TestAddEncryptedData(t *testing.T) {
	// беру адрес тестовой БД из переменной окружения
	databaseDsn := os.Getenv(envDatabaseName)
	assert.NotEqual(t, "", databaseDsn)

	// создаю соединение с базой данных
	conn, err := sql.Open("pgx", databaseDsn)
	require.NoError(t, err)
	defer conn.Close()

	// Проверка соединения с БД
	ctx := context.Background()
	err = conn.PingContext(ctx)
	require.NoError(t, err)

	// создаю экземпляр хранилища
	stor := NewStore(conn)
	err = stor.Bootstrap(ctx)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	{
		// Test. successful add data--------------------------------
		encryptedData := []byte("some encrypted data")
		userID := "test user id"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          "first data",
		}

		// добавляю новые данные в хранилище
		ok, err := stor.AddEncryptedData(ctx, userID, userData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// попытка добавить уже существующие данные
		ok, err = stor.AddEncryptedData(ctx, userID, userData)
		require.Error(t, err)
		assert.Equal(t, false, ok)

		// добавляю данные с тем-же именем, но для другого пользователя
		antoherUserID := "another test user id"
		ok, err = stor.AddEncryptedData(ctx, antoherUserID, userData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		// проверяю, что данные успешно добавились, ведь теперь не получится их повторно добавить
		ok, err = stor.AddEncryptedData(ctx, antoherUserID, userData)
		require.Error(t, err)
		assert.Equal(t, false, ok)

		// Проверка хранящихся в БД данных
		data, err := stor.GetAllEncryptedData(ctx, userID)
		require.NoError(t, err)
		checkData := data[0][0]
		assert.Equal(t, userData.EncryptedData, checkData.EncryptedData)
		assert.Equal(t, userData.Name, checkData.Name)
	}
	{
		// Test. Context exceeded
		encryptedData := []byte("some encrypted data")
		userID := "test user id"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          "first data",
		}

		ctx, cancel := context.WithCancel(context.Background())
		// отменяю контекст
		cancel()
		// добавляю новые данные в хранилище
		_, err := stor.AddEncryptedData(ctx, userID, userData)
		require.Error(t, err)
	}
}

func TestReplaceEncryptedData(t *testing.T) {
	// беру адрес тестовой БД из переменной окружения
	databaseDsn := os.Getenv(envDatabaseName)
	assert.NotEqual(t, "", databaseDsn)

	// создаю соединение с базой данных
	conn, err := sql.Open("pgx", databaseDsn)
	require.NoError(t, err)
	defer conn.Close()

	// Проверка соединения с БД
	ctx := context.Background()
	err = conn.PingContext(ctx)
	require.NoError(t, err)

	// создаю экземпляр хранилища
	stor := NewStore(conn)
	err = stor.Bootstrap(ctx)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	{
		// Test. successful add data--------------------------------
		encryptedData := []byte("some encrypted data")
		userID := "test user id"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          "first data",
		}

		// добавляю новые данные в хранилище
		ok, err := stor.AddEncryptedData(ctx, userID, userData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// изменяю уже сохраненные данные
		anotherUserData := data.EncryptedData{
			EncryptedData: []byte("another test data"),
			Name:          "first data",
		}

		ok, err = stor.ReplaceEncryptedData(ctx, userID, anotherUserData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// Проверка хранящихся в БД данных
		data, err := stor.GetAllEncryptedData(ctx, userID)
		require.NoError(t, err)
		checkData := data[0][0]
		assert.Equal(t, anotherUserData.EncryptedData, checkData.EncryptedData)
		assert.Equal(t, userData.Name, checkData.Name)
	}
	{
		// Test. Context exceeded
		ctx, cancel := context.WithCancel(context.Background())
		// отменяю контекст
		cancel()
		// пытаюсь изменить данные в хранилище
		_, err = stor.ReplaceEncryptedData(ctx, "some user id", data.EncryptedData{
			EncryptedData: []byte("some encrypted data"),
			Name:          "some data name",
		})
		require.Error(t, err)
	}
	{
		// Test. Data does not exist
		// Попытка исправить данные, которых нет в хранилище
		encryptedData := []byte("some not exist encrypted data")
		userID := "test user id"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          "not exist data",
		}
		ok, err := stor.ReplaceEncryptedData(ctx, userID, userData)
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
}

func TestGetAllEncryptedData(t *testing.T) {
	// беру адрес тестовой БД из переменной окружения
	databaseDsn := os.Getenv(envDatabaseName)
	assert.NotEqual(t, "", databaseDsn)

	// создаю соединение с базой данных
	conn, err := sql.Open("pgx", databaseDsn)
	require.NoError(t, err)
	defer conn.Close()

	// Проверка соединения с БД
	ctx := context.Background()
	err = conn.PingContext(ctx)
	require.NoError(t, err)

	// создаю экземпляр хранилища
	stor := NewStore(conn)
	err = stor.Bootstrap(ctx)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	{
		// Test. successful add data--------------------------------
		encryptedData := []byte("some encrypted data")
		userID := "test user id"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          "first data",
		}

		// добавляю новые данные в хранилище
		ok, err := stor.AddEncryptedData(ctx, userID, userData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// Проверка хранящихся в БД данных
		data, err := stor.GetAllEncryptedData(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(data))
		assert.Equal(t, 1, len(data[0]))

		checkData := data[0][0]
		assert.Equal(t, userData.EncryptedData, checkData.EncryptedData)
		assert.Equal(t, userData.Name, checkData.Name)

		// Добавляю такие-же данные, но для другого пользователя и проверяю их наличие в хранилище
		anotherUserID := "another test user id"
		ok, err = stor.AddEncryptedData(ctx, anotherUserID, userData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// Проверка хранящихся в БД данных
		anotherData, err := stor.GetAllEncryptedData(ctx, anotherUserID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(anotherData))
		assert.Equal(t, 1, len(anotherData[0]))

		anotherCheckData := anotherData[0][0]
		assert.Equal(t, userData.EncryptedData, anotherCheckData.EncryptedData)
		assert.Equal(t, userData.Name, anotherCheckData.Name)
	}
}
