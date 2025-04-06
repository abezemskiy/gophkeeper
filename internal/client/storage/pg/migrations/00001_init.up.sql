BEGIN TRANSACTION;

-- Создание таблицы auth
CREATE TABLE IF NOT EXISTS auth (
    login VARCHAR(128) PRIMARY KEY,
    hash VARCHAR(256),
    id VARCHAR(256),
    token VARCHAR(256)
);

-- Уникальный индекс по login
CREATE UNIQUE INDEX IF NOT EXISTS login ON auth (login);

-- Создание таблицы user_data
CREATE TABLE IF NOT EXISTS user_data (
    id SERIAL PRIMARY KEY,                            -- Уникальный идентификатор записи
    user_id VARCHAR(256) NOT NULL,                    -- ID пользователя
    data_name VARCHAR(128) NOT NULL,                  -- Имя данных
    encrypted_data BYTEA[],                           -- Массив зашифрованных данных
    status INT NOT NULL,                              -- Поле статуса

    CONSTRAINT unique_user_data UNIQUE (user_id, data_name)  -- Уникальность имени данных для пользователя
);

-- Индекс по user_id
CREATE INDEX IF NOT EXISTS user_id ON user_data (user_id);

COMMIT;