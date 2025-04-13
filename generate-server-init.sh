#!/bin/bash
set -e

# Проверим, что .env существует
if [ ! -f .env ]; then
  echo ".env файл не найден!"
  exit 1
fi

# Загрузим переменные из .env
set -a
source .env
set +a

# Проверим, что нужные переменные заданы
: "${SERVER_DB_USER:?Не задан SERVER_DB_USER}"
: "${SERVER_DB_PASSWORD:?Не задан SERVER_DB_PASSWORD}"
: "${SERVER_DB_NAME:?Не задан SERVER_DB_NAME}"

# Путь к итоговому SQL
INIT_SQL=./db/server/init/0001_init.sql

# Создаём init.sql
mkdir -p ./db/server/init

cat > "$INIT_SQL" <<EOF
DO \$\$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_catalog.pg_roles WHERE rolname = '${SERVER_DB_USER}'
    ) THEN
        CREATE USER "${SERVER_DB_USER}" PASSWORD '${SERVER_DB_PASSWORD}';
    END IF;
END
\$\$;

DO \$\$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_database WHERE datname = '${SERVER_DB_NAME}'
    ) THEN
        CREATE DATABASE "${SERVER_DB_NAME}"
            OWNER "${SERVER_DB_USER}"
            ENCODING 'UTF8'
            LC_COLLATE = 'en_US.utf8'
            LC_CTYPE = 'en_US.utf8';
    END IF;
END
\$\$;
EOF

echo "[✓] SQL-файл создан: $INIT_SQL"
