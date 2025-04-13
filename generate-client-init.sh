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
: "${CLIENT_DB_USER:?Не задан CLIENT_DB_USER}"
: "${CLIENT_DB_PASSWORD:?Не задан CLIENT_DB_PASSWORD}"
: "${CLIENT_DB_NAME:?Не задан CLIENT_DB_NAME}"

# Путь к итоговому SQL
INIT_SQL=./db/client/init/0001_init.sql

# Создаём init.sql
mkdir -p ./db/client/init

cat > "$INIT_SQL" <<EOF
DO \$\$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_catalog.pg_roles WHERE rolname = '${CLIENT_DB_USER}'
    ) THEN
        CREATE USER "${CLIENT_DB_USER}" PASSWORD '${CLIENT_DB_PASSWORD}';
    END IF;
END
\$\$;

DO \$\$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_database WHERE datname = '${CLIENT_DB_NAME}'
    ) THEN
        CREATE DATABASE "${CLIENT_DB_NAME}"
            OWNER "${CLIENT_DB_USER}"
            ENCODING 'UTF8'
            LC_COLLATE = 'en_US.utf8'
            LC_CTYPE = 'en_US.utf8';
    END IF;
END
\$\$;
EOF

echo "[✓] SQL-файл создан: $INIT_SQL"
