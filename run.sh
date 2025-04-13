#!/bin/bash

set -e

ENV_FILE=".env"

generate_secret() {
  LC_ALL=C tr -dc 'a-zA-Z0-9' </dev/urandom | head -c "$1"
}

if [ -f "$ENV_FILE" ]; then
  echo ".env уже существует. Используем текущие переменные."
else
  echo "Создаём .env с параметрами для клиентской и серверной БД..."

  # Клиентская БД
  POSTGRES_USER_CLIENT="client_psql_user$(generate_secret 6)"
  POSTGRES_PASSWORD_CLIENT="client_psql_pass$(generate_secret 16)"
  POSTGRES_DB_CLIENT="gk_client"

  CLIENT_DB_USER="client_db_user$(generate_secret 6)"
  CLIENT_DB_PASSWORD="client_db_pass$(generate_secret 16)"
  CLIENT_DB_NAME="client_db_name$(generate_secret 6)"

  # Серверная БД
  POSTGRES_USER_SERVER="server_psql_user$(generate_secret 6)"
  POSTGRES_PASSWORD_SERVER="server_psql_pass$(generate_secret 16)"
  POSTGRES_DB_SERVER="gk_server"

  SERVER_DB_USER="server_db_user$(generate_secret 6)"
  SERVER_DB_PASSWORD="server_db_pass$(generate_secret 16)"
  SERVER_DB_NAME="server_db_name$(generate_secret 6)"

  # Конфигурация сервера
  SERVER_SECRET_KEY="server_key$(generate_secret 16)"

  cat > "$ENV_FILE" <<EOF
# Client DB
POSTGRES_USER_CLIENT=$POSTGRES_USER_CLIENT
POSTGRES_PASSWORD_CLIENT=$POSTGRES_PASSWORD_CLIENT
POSTGRES_DB_CLIENT=$POSTGRES_DB_CLIENT

CLIENT_DB_USER=$CLIENT_DB_USER
CLIENT_DB_PASSWORD=$CLIENT_DB_PASSWORD
CLIENT_DB_NAME=$CLIENT_DB_NAME

# Server DB
POSTGRES_USER_SERVER=$POSTGRES_USER_SERVER
POSTGRES_PASSWORD_SERVER=$POSTGRES_PASSWORD_SERVER
POSTGRES_DB_SERVER=$POSTGRES_DB_SERVER

SERVER_DB_USER=$SERVER_DB_USER
SERVER_DB_PASSWORD=$SERVER_DB_PASSWORD
SERVER_DB_NAME=$SERVER_DB_NAME

# DSN
CLIENT_DSN=postgresql://$CLIENT_DB_USER:$CLIENT_DB_PASSWORD@gophkeeper-client-postgres:5432/$CLIENT_DB_NAME?sslmode=disable
SERVER_DSN=postgresql://$SERVER_DB_USER:$SERVER_DB_PASSWORD@gophkeeper-server-postgres:5432/$SERVER_DB_NAME?sslmode=disable

#SERVER VARIABLES
SERVER_SECRET_KEY=$SERVER_SECRET_KEY
EOF

  echo ".env создан."
fi

echo "[INFO] Генерация client init.sql из .env..."
./generate-client-init.sh

echo "[INFO] Генерация server init.sql из .env..."
./generate-server-init.sh

echo "[INFO] Запуск контейнеров..."
docker compose up --build

