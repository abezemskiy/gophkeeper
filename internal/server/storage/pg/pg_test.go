//go:build integration_tests
// +build integration_tests

package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/abezemskiy/gophkeeper/internal/repositories/data"

	"math/rand"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	code, err := runMain(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

const (
	testDBName       = "test"
	testUserName     = "test"
	testUserPassword = "test"
)

var (
	getDSN          func() string
	getSUConnection func() (*pgx.Conn, error)
)

func initGetDSN(hostAndPort string) {
	getDSN = func() string {
		return fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			testUserName,
			testUserPassword,
			hostAndPort,
			testDBName,
		)
	}
}

func initGetSUConnection(hostPort string) error {
	host, port, err := getHostPort(hostPort)
	if err != nil {
		return fmt.Errorf("failed to extract the host and port parts from the string %s: %w", hostPort, err)
	}
	getSUConnection = func() (*pgx.Conn, error) {
		conn, err := pgx.Connect(pgx.ConnConfig{
			Host:     host,
			Port:     port,
			Database: "postgres",
			User:     "postgres",
			Password: "postgres",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get a super user connection: %w", err)
		}
		return conn, nil
	}
	return nil
}

func runMain(m *testing.M) (int, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return 1, fmt.Errorf("failed to initialize a pool: %w", err)
	}

	pg, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "17.2",
			Name:       "server-migrations-integration-tests",
			Env: []string{
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=postgres",
			},
			ExposedPorts: []string{"5432/tcp"},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return 1, fmt.Errorf("failed to run the postgres container: %w", err)
	}

	defer func() {
		if err := pool.Purge(pg); err != nil {
			log.Printf("failed to purge the postgres container: %v", err)
		}
	}()

	hostPort := pg.GetHostPort("5432/tcp")
	initGetDSN(hostPort)
	if err := initGetSUConnection(hostPort); err != nil {
		return 1, err
	}

	pool.MaxWait = 10 * time.Second
	var conn *pgx.Conn
	if err := pool.Retry(func() error {
		conn, err = getSUConnection()
		if err != nil {
			return fmt.Errorf("server: failed to connect to the DB: %w", err)
		}
		return nil
	}); err != nil {
		return 1, err
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to correctly close the connection: %v", err)
		}
	}()

	if err := createTestDB(conn); err != nil {
		return 1, fmt.Errorf("failed to create a test DB: %w", err)
	}

	exitCode := m.Run()

	return exitCode, nil
}

func createTestDB(conn *pgx.Conn) error {
	_, err := conn.Exec(
		fmt.Sprintf(
			`CREATE USER %s PASSWORD '%s'`,
			testUserName,
			testUserPassword,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create a test user: %w", err)
	}

	_, err = conn.Exec(
		fmt.Sprintf(`
			CREATE DATABASE %s
				OWNER '%s'
				ENCODING 'UTF8'
				LC_COLLATE = 'en_US.utf8'
				LC_CTYPE = 'en_US.utf8'
			`, testDBName, testUserName,
		),
	)

	if err != nil {
		return fmt.Errorf("failed to create a test DB: %w", err)
	}

	return nil
}

func getHostPort(hostPort string) (string, uint16, error) {
	hostPortParts := strings.Split(hostPort, ":")
	if len(hostPortParts) != 2 {
		return "", 0, fmt.Errorf("got an invalid host-port string: %s", hostPort)
	}

	portStr := hostPortParts[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("failed to cast the port %s to an int: %w", portStr, err)
	}
	return hostPortParts[0], uint16(port), nil
}

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
	// беру адрес тестовой БД
	databaseDsn := getDSN()

	// создаю экземпляр хранилища
	stor, err := NewStore(context.Background(), databaseDsn)
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
	// беру адрес тестовой БД
	databaseDsn := getDSN()

	ctx := context.Background()
	// создаю экземпляр хранилища
	stor, err := NewStore(ctx, databaseDsn)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	{
		// Test. successful authorization--------------------------------
		// регистрирую пользователя
		sLogin := "login"
		sHash := "hash"
		sID := "id"
		err = stor.Register(ctx, sLogin, sHash, sID)
		require.NoError(t, err)

		// получаю данные пользователя для авторизации по его логину
		//var data identity.AuthorizationData
		data, ok, err := stor.Authorize(ctx, sLogin)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		assert.Equal(t, sHash, data.Hash)
		assert.Equal(t, sID, data.ID)
	}
	{
		// Test. context is exceeded--------------------------------
		// регистрирую пользователя
		sLogin := "exceeded login"
		sHash := "hash"
		sID := "id"
		err = stor.Register(ctx, sLogin, sHash, sID)
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
	// беру адрес тестовой БД
	databaseDsn := getDSN()

	ctx := context.Background()
	// создаю экземпляр хранилища
	stor, err := NewStore(ctx, databaseDsn)
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
		ok, err := stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// попытка добавить уже существующие данные
		ok, err = stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.Error(t, err)
		assert.Equal(t, false, ok)

		// добавляю данные с тем-же именем, но для другого пользователя
		antoherUserID := "another test user id"
		ok, err = stor.AddEncryptedData(ctx, antoherUserID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
		// проверяю, что данные успешно добавились, ведь теперь не получится их повторно добавить
		ok, err = stor.AddEncryptedData(ctx, antoherUserID, userData, data.SAVED)
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
		_, err := stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.Error(t, err)
	}
}

func TestReplaceEncryptedData(t *testing.T) {
	// беру адрес тестовой БД
	databaseDsn := getDSN()

	ctx := context.Background()
	// создаю экземпляр хранилища
	stor, err := NewStore(ctx, databaseDsn)
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
		ok, err := stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// изменяю уже сохраненные данные
		anotherUserData := data.EncryptedData{
			EncryptedData: []byte("another test data"),
			Name:          "first data",
		}

		ok, err = stor.ReplaceEncryptedData(ctx, userID, anotherUserData, data.SAVED)
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
		}, data.SAVED)
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
		ok, err := stor.ReplaceEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
}

func TestGetAllEncryptedData(t *testing.T) {
	// Набор символов для генерации
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	generateRandomString := func(seed, length int) string {
		r := rand.New(rand.NewSource(int64(seed))) // Инициализация генератора случайных чисел

		result := make([]byte, length)
		for i := range result {
			result[i] = charset[r.Intn(len(charset))]
		}
		return string(result)
	}

	// беру адрес тестовой БД
	databaseDsn := getDSN()

	ctx := context.Background()
	// создаю экземпляр хранилища
	stor, err := NewStore(ctx, databaseDsn)
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
		ok, err := stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// Проверка хранящихся в БД данных
		getData, err := stor.GetAllEncryptedData(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(getData))
		assert.Equal(t, 1, len(getData[0]))

		checkData := getData[0][0]
		assert.Equal(t, userData.EncryptedData, checkData.EncryptedData)
		assert.Equal(t, userData.Name, checkData.Name)

		// Добавляю такие-же данные, но для другого пользователя и проверяю их наличие в хранилище
		anotherUserID := "another test user id"
		ok, err = stor.AddEncryptedData(ctx, anotherUserID, userData, data.SAVED)
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
	{
		// add data in cycle
		userID := "cycle user id"

		// добавляю ещё данные для того-же пользователя и проверяю их наличие
		i := 10
		for j := range i {
			name := generateRandomString(j, 14)
			ok, err := stor.AddEncryptedData(ctx, userID, data.EncryptedData{
				EncryptedData: []byte("some data"),
				Name:          name,
			}, data.SAVED)
			require.NoError(t, err)
			assert.Equal(t, true, ok)

			data, err := stor.GetAllEncryptedData(ctx, userID)
			require.NoError(t, err)
			assert.Equal(t, j+1, len(data))
			assert.Equal(t, 1, len(data[j]))
			assert.Equal(t, name, data[j][0].Name)
		}

	}
	{
		// Test context exceeded
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = stor.GetAllEncryptedData(ctx, "context exceede user id")
		require.Error(t, err)
	}
	{
		// Test empty data
		res, err := stor.GetAllEncryptedData(ctx, "empty data user id")
		require.NoError(t, err)
		assert.Equal(t, 0, len(res))
	}
}

func TestGetEncryptedDataByStatus(t *testing.T) {
	// Набор символов для генерации
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	generateRandomString := func(seed, length int) string {
		r := rand.New(rand.NewSource(int64(seed))) // Инициализация генератора случайных чисел

		result := make([]byte, length)
		for i := range result {
			result[i] = charset[r.Intn(len(charset))]
		}
		return string(result)
	}

	// беру адрес тестовой БД
	databaseDsn := getDSN()

	ctx := context.Background()
	// создаю экземпляр хранилища
	stor, err := NewStore(ctx, databaseDsn)
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
		ok, err := stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// Проверка хранящихся в БД данных
		getData, err := stor.GetEncryptedDataByStatus(ctx, userID, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, 1, len(getData))
		assert.Equal(t, 1, len(getData[0]))

		checkData := getData[0][0]
		assert.Equal(t, userData.EncryptedData, checkData.EncryptedData)
		assert.Equal(t, userData.Name, checkData.Name)

		// Добавляю такие-же данные, но для другого пользователя и проверяю их наличие в хранилище
		anotherUserID := "another test user id"
		ok, err = stor.AddEncryptedData(ctx, anotherUserID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// Проверка хранящихся в БД данных
		anotherData, err := stor.GetEncryptedDataByStatus(ctx, anotherUserID, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, 1, len(anotherData))
		assert.Equal(t, 1, len(anotherData[0]))

		anotherCheckData := anotherData[0][0]
		assert.Equal(t, userData.EncryptedData, anotherCheckData.EncryptedData)
		assert.Equal(t, userData.Name, anotherCheckData.Name)

		// Попытка извлечь данные со статусом, которого нет в хранилище
		conflictData, err := stor.GetEncryptedDataByStatus(ctx, anotherUserID, data.CONFLICT)
		require.NoError(t, err)
		assert.Equal(t, 0, len(conflictData))
	}
	{
		// add data in cycle
		userID := "cycle user id"

		// добавляю ещё данные для того-же пользователя и проверяю их наличие
		i := 10
		for j := range i {
			name := generateRandomString(j, 14)
			ok, err := stor.AddEncryptedData(ctx, userID, data.EncryptedData{
				EncryptedData: []byte("some data"),
				Name:          name,
			}, data.SAVED)
			require.NoError(t, err)
			assert.Equal(t, true, ok)

			data, err := stor.GetEncryptedDataByStatus(ctx, userID, data.SAVED)
			require.NoError(t, err)
			assert.Equal(t, j+1, len(data))
			assert.Equal(t, 1, len(data[j]))
			assert.Equal(t, name, data[j][0].Name)
		}

	}
	{
		// Test context exceeded
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = stor.GetEncryptedDataByStatus(ctx, "context exceede user id", data.SAVED)
		require.Error(t, err)
	}
	{
		// Test empty data
		res, err := stor.GetEncryptedDataByStatus(ctx, "empty data user id", data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, 0, len(res))
	}
}

func TestDeleteEncryptedData(t *testing.T) {
	// беру адрес тестовой БД
	databaseDsn := getDSN()

	ctx := context.Background()
	// создаю экземпляр хранилища
	stor, err := NewStore(ctx, databaseDsn)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	{
		// Test. successful delete data--------------------------------
		encryptedData := []byte("some encrypted data")
		userID := "test user id"
		nameData := "first data"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          nameData,
		}

		// добавляю новые данные в хранилище
		ok, err := stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		ok, err = stor.DeleteEncryptedData(ctx, userID, nameData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// попытка извлечь данные, которых не существует
		res, err := stor.GetAllEncryptedData(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(res))

		// добавляю заново ранее удаленные данные и проверяю, что запрос выполнен успешно
		ok, err = stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)
	}
	{
		// Attempting to delete does not existing data
		ok, err := stor.DeleteEncryptedData(ctx, "does not existing user id", "does not existing data name")
		require.NoError(t, err)
		assert.Equal(t, false, ok)

		// Добавляю данные пользователя, чтобы попытаться удалить данные у существующего полязователя
		// но с несуществующим имененм данных.
		encryptedData := []byte("some encrypted data")
		userID := "some existing user id"
		nameData := "existing data"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          nameData,
		}

		// добавляю новые данные в хранилище
		ok, err = stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		ok, err = stor.DeleteEncryptedData(ctx, userID, "different name data")
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
	{
		// Test. Context exceeded
		ctx, cancel := context.WithCancel(context.Background())
		// отменяю контекст
		cancel()
		// добавляю новые данные в хранилище
		_, err := stor.DeleteEncryptedData(ctx, "context exceeded id", "user data name")
		require.Error(t, err)
	}
}

func TestAppendEncryptedData(t *testing.T) {
	// беру адрес тестовой БД
	databaseDsn := getDSN()

	ctx := context.Background()
	// создаю экземпляр хранилища
	stor, err := NewStore(ctx, databaseDsn)
	require.NoError(t, err)

	// очищаю данные в БД от предыдущих запусков
	cleanBD(t, databaseDsn, stor)
	defer cleanBD(t, databaseDsn, stor)

	{
		// Test. successful add data--------------------------------
		encryptedData := []byte("some encrypted data")
		userID := "test user id"
		dataName := "first data"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          dataName,
		}

		// добавляю новые данные в хранилище
		ok, err := stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// добавляю новую версию данных в хранилище
		additionEncryptedData := []byte("addition encrypted data")
		additionData := data.EncryptedData{
			EncryptedData: additionEncryptedData,
			Name:          dataName,
		}
		ok, err = stor.AppendEncryptedData(ctx, userID, additionData)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		// извлекаю данные пользователя и убеждаюсь хранится пара версий данных с одним имененем
		data, err := stor.GetAllEncryptedData(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(data))
		assert.Equal(t, 2, len(data[0]))
		// проверка корректности первой версии данных
		assert.Equal(t, dataName, data[0][0].Name)
		assert.Equal(t, encryptedData, data[0][0].EncryptedData)

		// проверка корректности второй версии данных
		assert.Equal(t, dataName, data[0][1].Name)
		assert.Equal(t, additionEncryptedData, data[0][1].EncryptedData)
	}
	{
		// Test. Context exceeded
		ctx, cancel := context.WithCancel(context.Background())
		// отменяю контекст
		cancel()
		// добавляю новые данные в хранилище
		_, err = stor.AppendEncryptedData(ctx, "context exceeded user id", data.EncryptedData{})
		require.Error(t, err)
	}
	{
		// Тест с попыткой добавить данные, которых не существует
		ok, err := stor.AppendEncryptedData(ctx, "not exist user id", data.EncryptedData{})
		require.NoError(t, err)
		assert.Equal(t, false, ok)

		// Тест с попыткой добавить данные, которых не существует, но для существующего пользователя
		encryptedData := []byte("some encrypted data")
		userID := "not existing data test user id"
		dataName := "first data"
		userData := data.EncryptedData{
			EncryptedData: encryptedData,
			Name:          dataName,
		}

		// добавляю новые данные в хранилище
		ok, err = stor.AddEncryptedData(ctx, userID, userData, data.SAVED)
		require.NoError(t, err)
		assert.Equal(t, true, ok)

		ok, err = stor.AppendEncryptedData(ctx, userID, data.EncryptedData{
			Name: "different data name",
		})
		require.NoError(t, err)
		assert.Equal(t, false, ok)
	}
}
