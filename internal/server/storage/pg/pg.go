package pg

import (
	"context"
	"database/sql"
	"fmt"
	"gophkeeper/internal/repositories/data"
	"gophkeeper/internal/repositories/identity"
	"time"

	"github.com/lib/pq"
)

// Store - реализует интерфейс storage.IStorage и позволяет взаимодествовать с СУБД PostgreSQL.
type Store struct {
	// Поле conn содержит объект соединения с СУБД
	conn *sql.DB
}

// NewStore - возвращает новый экземпляр PostgreSQL-хранилища.
func NewStore(conn *sql.DB) *Store {
	return &Store{
		conn: conn,
	}
}

// Bootstrap - подготавливает БД к работе, создавая необходимы таблицы и индексы.
func (s Store) Bootstrap(ctx context.Context) error {
	// запускаю транзакцию
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction error, %w", err)
	}

	// откат транзакции в случае ошибки
	defer tx.Rollback()

	// создаю таблицу для хранения данных пользователя -------------------------
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS auth (
			login varchar(128) PRIMARY KEY,
			hash varchar(128),
			id varchar(128)
		)
	`)
	if err != nil {
		return fmt.Errorf("create table auth error, %w", err)
	}
	// создаю уникальный индекс для логина
	_, err = tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS login ON auth (login)`)
	if err != nil {
		return fmt.Errorf("create unique index in auth table error, %w", err)
	}

	// создаю таблицу для хранения uplinks -----------------------------------------
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_data (
    		id SERIAL PRIMARY KEY,                 							-- Уникальный идентификатор записи
    		user_id varchar(128) NOT NULL,                  				-- ID пользователя
    		data_name varchar(128) NOT NULL,               					-- Имя данных
    		encrypted_data BYTEA[],                							-- Массив зашифрованных данных
    		status INT NOT NULL,                   							-- Поле статуса

    		CONSTRAINT unique_user_data UNIQUE (user_id, data_name) 		-- Гарант уникальности имени данных для пользователя
		)
	`)

	if err != nil {
		return fmt.Errorf("create table user_data error, %w", err)
	}
	// создаю уникальный индекс для ID пользователя
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS user_id ON user_data (user_id)`)
	if err != nil {
		return fmt.Errorf("create unique index in user_data table error, %w", err)
	}

	// Создаю таблицу для хранения даты последнего визита пользователя-------------------
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_last_seen (
			user_id varchar(128) PRIMARY KEY,
			last_seen TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create table user_last_seen error, %w", err)
	}
	// создаю уникальный индекс для ID пользователя
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS user_id ON user_last_seen (user_id)`)
	if err != nil {
		return fmt.Errorf("create unique index in user_last_seen table error, %w", err)
	}

	// коммитим транзакцию
	return tx.Commit()
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

	// удаляю все записи в таблице user_last_seen----------------------
	_, err = tx.ExecContext(ctx, `
		TRUNCATE TABLE user_last_seen
	`)
	if err != nil {
		return fmt.Errorf("truncate table user_last_seen error, %w", err)
	}

	// коммитим транзакцию
	return tx.Commit()
}

// Register - сохраняет в базу данные нового пользователя.
func (s Store) Register(ctx context.Context, login, hash, id string) error {
	query := `
	INSERT INTO auth (login, hash, id)
	VALUES ($1, $2, $3)
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, login, hash, id)
	return err
}

// Authorize - получаю авторизационные данные пользователя (хэш) по логину.
// В случае, если пользователь с переданным логином не найден, возвращается ошибка.
func (s Store) Authorize(ctx context.Context, login string) (data identity.AuthorizationData, ok bool, err error) {
	query := `
		SELECT  hash,
				id
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

	err = row.Scan(&data.Hash, &data.ID)
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
func (s Store) ReplaceEncryptedData(ctx context.Context, idUser string, userData data.EncryptedData, status int) (bool, error) {
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
	result, err := stmt.ExecContext(ctx, idUser, userData.Name, pq.Array([][]byte{userData.EncryptedData}), status)

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

// AppendEncryptedData - метод для сохранения дополнительной версии существующих данных в случае конфликта.
func (s Store) AppendEncryptedData(ctx context.Context, idUser string, userData data.EncryptedData) (bool, error) {
	query := `
	UPDATE user_data
	SET 
    	encrypted_data = array_append(encrypted_data, $3), -- Добавление новой версии данных в массив
    	status = $4 									   -- Обновление статуса
	WHERE user_id = $1 AND data_name = $2
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, idUser, userData.Name, userData.EncryptedData, data.CONFLICT)
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

// AddDateOfLastVisit - метод для добавления даты последнего визита пользователя.
func (s Store) AddDateOfLastVisit(ctx context.Context, idUser string, date time.Time) error {
	query := `
	INSERT INTO user_last_seen (user_id, last_seen) 
	VALUES ($1, $2) 
	ON CONFLICT (user_id) DO UPDATE SET last_seen = EXCLUDED.last_seen
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, idUser, date)
	if err != nil {
		return fmt.Errorf("query execution error, %w", err)
	}
	return nil
}

// GetDateOfLastVisit - метод для выгрузки даты последнего визита пользователя.
func (s Store) GetDateOfLastVisit(ctx context.Context, idUser string) (time.Time, error) {
	query := `
	SELECT last_seen
	FROM user_last_seen
	WHERE user_id = $1
`
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return time.Time{}, fmt.Errorf("prepare context error, %w", err)
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, idUser)

	var lastVisit time.Time
	err = row.Scan(&lastVisit)
	if err != nil {
		return time.Time{}, fmt.Errorf("scan row error, %w", err)
	}
	return lastVisit, nil
}
