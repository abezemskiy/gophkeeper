package pg

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/repositories/data"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// Store - реализует интерфейс storage.IStorage и позволяет взаимодествовать с СУБД PostgreSQL.
type Store struct {
	// Поле conn содержит объект соединения с СУБД
	conn *sql.DB
}

// NewStore - применяет миграции и возвращает новый экземпляр PostgreSQL-хранилища.
func NewStore(ctx context.Context, dsn string) (*Store, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	// Подключение к базе данных
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connection to database: %v by address %s", err, dsn)
	}

	// Проверка соединения с БД
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error checking connection with database: %v", err)
	}

	return &Store{
		conn: db,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

// Disable - очищает БД, удаляя записи из таблиц.
// Метод необходим для тестирования, чтобы в процессе удалять тестовые записи.
func (s Store) Disable(ctx context.Context) error {
	// запускаем транзакцию
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction error, %w", err)
	}
	// в случае неупешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	// удаляю все записи в таблице auth----------------------
	_, err = tx.ExecContext(ctx, `
		TRUNCATE TABLE auth
	`)
	if err != nil {
		return fmt.Errorf("truncate table auth error, %w", err)
	}

	// удаляю все записи в таблице user_data----------------------
	_, err = tx.ExecContext(ctx, `
		TRUNCATE TABLE user_data
	`)
	if err != nil {
		return fmt.Errorf("truncate table user_data error, %w", err)
	}

	// коммитим транзакцию
	return tx.Commit()
}

// Register - сохраняет в базу данные нового пользователя. Если такой пользователь уже зарегистрирован, вернется false.
func (s Store) Register(ctx context.Context, login, hash, id, token string) (bool, error) {
	query := `
	INSERT INTO auth (login, hash, id, token)
	VALUES ($1, $2, $3, $4)
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, login, hash, id, token)

	if err != nil {
		// Обрабатываю полученную ошибку
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Ошибка соответствует ошибке при попытке установить повторяющеся поле типа "PRIMARY KEY".
			// Пользователь уже зарегистрирован.
			return false, nil
		}
		return false, fmt.Errorf("query execution error, %w", err)
	}

	// Пользователь успешно зарегистрирован
	return true, nil
}

// Authorize - получаю авторизационные данные пользователя (хэш) по логину.
// В случае, если пользователь с переданным логином не найден, возвращается ошибка.
func (s Store) Authorize(ctx context.Context, login string) (data identity.UserInfo, ok bool, err error) {
	query := `
		SELECT  hash,
				id,
				token
		FROM auth
		WHERE login = $1
	`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		err = fmt.Errorf("prepare context error, %w", err)
		return
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, login)

	err = row.Scan(&data.Hash, &data.ID, &data.Token)
	if err != nil {
		// пользователь не найден
		err = nil
		ok = false
		return
	}
	ok = true
	return
}

// AddEncryptedData - метод для добавления уникальных зашифрованныч данных по id в хранилище.
// В случае если данные не уникальны, возвращается false.
func (s Store) AddEncryptedData(ctx context.Context, idUser string, userData data.EncryptedData, status int) (bool, error) {
	query := `
		INSERT INTO user_data (user_id, data_name, encrypted_data, status)
		VALUES ($1, $2, $3, $4)
	`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, idUser, userData.Name, [][]byte{userData.EncryptedData}, status)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			// Код ошибки 23505 - unique_violation
			// конфликт, уже существуют данные с таким именем для данного пользователя
			return false, nil
		}
		return false, fmt.Errorf("query execution error, %w", err)
	}
	return true, nil
}

// ReplaceEncryptedData - метод для замены старых данных значениями новых.
// В случае попытки заменить данные, когда данные с текущим id полязователя и именем ещё не загружены в хранилище
// возвращается false.
func (s Store) ReplaceEncryptedData(ctx context.Context, idUser string, userData data.EncryptedData, _ int) (bool, error) {
	query := `
	UPDATE user_data
	SET encrypted_data = $3, status = $4
	WHERE user_id = $1 AND data_name = $2
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, idUser, userData.Name, [][]byte{userData.EncryptedData}, data.CHANGED)

	if err != nil {
		return false, fmt.Errorf("query execution error, %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// попытка обновить данные, которых не существует
		return false, nil
	}
	return true, nil
}

// GetAllEncryptedData - метод для выгрузки всех зашифрованных данных конкретного пользователя.
func (s Store) GetAllEncryptedData(ctx context.Context, idUser string) ([][]data.EncryptedData, error) {
	query := `
	SELECT  data_name,
			encrypted_data
	FROM user_data
	WHERE user_id = $1
	`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, idUser)
	if err != nil {
		return nil, fmt.Errorf("query execution error, %w", err)
	}

	result := make([][]data.EncryptedData, 0)
	defer rows.Close()
	for rows.Next() {
		// получаю массив байт, который представляет собой несколько версий данных в бинарном виде
		binaryData := make([][]byte, 0)

		// переменная для хранения имени данных
		var dataName string

		err = rows.Scan(&dataName, pq.Array(&binaryData))
		if err != nil {
			return nil, fmt.Errorf("scan error, %w", err)
		}
		dataVersions := make([]data.EncryptedData, 0, len(binaryData))
		for _, d := range binaryData {
			// преобразую данные из бинарного вида в структуру
			jsonData := data.EncryptedData{
				EncryptedData: d,
				Name:          dataName,
			}
			dataVersions = append(dataVersions, jsonData)
		}
		result = append(result, dataVersions)
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetEncryptedDataByStatus - метод для выгрузки всех зашифрованных данных конкретного пользователя с определенным статусом.
func (s Store) GetEncryptedDataByStatus(ctx context.Context, idUser string, status int) ([][]data.EncryptedData, error) {
	query := `
	SELECT  data_name,
			encrypted_data
	FROM user_data
	WHERE user_id = $1 AND status = $2
	`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, idUser, status)
	if err != nil {
		return nil, fmt.Errorf("query execution error, %w", err)
	}

	result := make([][]data.EncryptedData, 0)
	defer rows.Close()
	for rows.Next() {
		// получаю массив байт, который представляет собой несколько версий данных в бинарном виде
		binaryData := make([][]byte, 0)

		// переменная для хранения имени данных
		var dataName string

		err = rows.Scan(&dataName, pq.Array(&binaryData))
		if err != nil {
			return nil, fmt.Errorf("scan error, %w", err)
		}
		dataVersions := make([]data.EncryptedData, 0, len(binaryData))
		for _, d := range binaryData {
			// преобразую данные из бинарного вида в структуру
			jsonData := data.EncryptedData{
				EncryptedData: d,
				Name:          dataName,
			}
			dataVersions = append(dataVersions, jsonData)
		}
		result = append(result, dataVersions)
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteEncryptedData - метод для удаления данных в хранилище по id пользователя и имени данных.
// Если происходит попытка удалить несуществующие данные, возвращается false.
func (s Store) DeleteEncryptedData(ctx context.Context, idUser, dataName string) (bool, error) {
	query := `
	DELETE FROM user_data
	WHERE user_id = $1 AND data_name = $2
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, idUser, dataName)
	if err != nil {
		return false, fmt.Errorf("query execution error, %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Запись не найдена, попытка дополнить данные, которых не существует.
		return false, nil
	}
	return true, nil
}

// SetToken - метод для установки нового токена для конкретного пользователя.
// В случае, если не найден пользователь по данному логину возвращается false.
func (s Store) SetToken(ctx context.Context, login, token string) (bool, error) {
	query := `
	UPDATE auth
	SET token = $2
	WHERE login = $1
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, login, token)

	if err != nil {
		return false, fmt.Errorf("query execution error, %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// попытка обновить данные, которых не существует
		// пользователь с данным логином не зарегистрирован.
		return false, nil
	}
	return true, nil
}

// ChangeStatusOfEncryptedData - метод для изменения статуса существующих данных у пользователя по его ID.
// В случае, если пользователь или данные не найдены, возвращается false.
func (s Store) ChangeStatusOfEncryptedData(ctx context.Context, userID, dataName string, newStatus int) (ok bool, err error) {

	query := `
	UPDATE user_data
	SET status = $3
	WHERE user_id = $1 AND data_name = $2
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, userID, dataName, newStatus)

	if err != nil {
		return false, fmt.Errorf("query execution error, %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// попытка обновить данные, которых не существует
		// пользователь с данным логином не зарегистрирован или данные не найдены.
		return false, nil
	}
	return true, nil
}

// ReplaceDataWithMultiVersionData - метод для замены существующих в хранилище на данные с несколькими версиями
func (s Store) ReplaceDataWithMultiVersionData(ctx context.Context, idUser string, userData []data.EncryptedData,
	status int) (bool, error) {
	// проверяю, что существует как минимум одна версия данных
	if len(userData) == 0 {
		return false, fmt.Errorf("no one version of data is exists")
	}

	// Инициализирую слайс для хранения данных в виде слайса байт. Такой тип данных подходит для сохранения в БД
	dataToInsert := make([][]byte, len(userData))

	// Преобразую полученные данные пользователя в вид, готовый к сохранению в БД
	for i, d := range userData {
		dataToInsert[i] = d.EncryptedData
	}

	query := `
	UPDATE user_data
	SET encrypted_data = $3, status = $4
	WHERE user_id = $1 AND data_name = $2
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, idUser, userData[0].Name, dataToInsert, status)

	if err != nil {
		return false, fmt.Errorf("query execution error, %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// попытка обновить данные, которых не существует
		return false, nil
	}
	return true, nil
}

// GetStatus - метод для получения текущего статуса данных у пользователя с данным ID по имени данных.
// В случае, если данных не существует, возвращается false.
func (s Store) GetStatus(ctx context.Context, userID, dataName string) (status int, ok bool, err error) {
	query := `
	SELECT  status
	FROM user_data
	WHERE user_id = $1 AND data_name = $2
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		err = fmt.Errorf("prepare context error, %w", err)
		return
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, userID, dataName)

	err = row.Scan(&status)
	if err != nil {
		// данные не найдены
		err = nil
		ok = false
		return
	}
	ok = true
	return
}
