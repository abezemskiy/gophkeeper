package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/abezemskiy/gophkeeper/internal/common/identity/tools/token"
	"github.com/abezemskiy/gophkeeper/internal/server/config"
)

var (
	netAddr     string // адрес запуска сервиса
	databaseDsn string // адрес базы данных
	logLevel    string // уровень логирования
	configFile  string // путь к файлу конфигурации
	secretKey   string // секретный ключ для создания JWT
	expireToken int    // время действия JWT
)

// parseVariables - функция для установки конфигурационных параметров приложения.
// Конфигурирование приложения с приоритетом в порядке убывания: значения флагов, значения из файла, значения переменных окружения.
func parseVariables() error {
	parseFlags()
	parseConfigFile()
	parseEnvironment()

	// Проверяю корректность установки глобальных переменных
	err := checkVariables()
	if err != nil {
		return fmt.Errorf("failed to set global variable, %w", err)
	}

	// Устанавливаю полученные значения глобальных переменных
	token.SetSecretKey(secretKey)
	token.SerExpireHour(expireToken)
	return nil
}

// parseFlags - функция для определения параметров конфигурации из флагов.
func parseFlags() {
	flag.StringVar(&netAddr, "a", "", "address and port to run server")

	// настройка флага для хранения метрик в базе данных
	flag.StringVar(&databaseDsn, "d", "", "database connection address") // по умолчанию адрес не задан

	flag.StringVar(&logLevel, "l", "", "log level")
	flag.StringVar(&configFile, "c", "", "name of configuration file")
	flag.StringVar(&secretKey, "secret-key", "", "secret key for generating JWT")
	flagExpireToken := flag.Int("expire-token", 0, "JWT expiration date in hours")

	// Вызов flag.Parse() для парсинга аргументов
	flag.Parse()
	expireToken = *flagExpireToken
}

// parseConfigFile - функция для переопределения параметров конфигурации из файла конфигурации.
func parseConfigFile() {
	// если не указан файл конфигурации, то оставляю параметры запуска без изменения
	if configFile == "" {
		return
	}
	configs, err := config.ParseConfigFile(configFile)
	if err != nil {
		log.Fatalf("parse config file error: %v\n", err)
	}

	// обновляю параметры запуска если они не определены флагами
	if netAddr == "" {
		netAddr = configs.Address
	}
	if logLevel == "" {
		logLevel = configs.LogLevel
	}
	if databaseDsn == "" {
		databaseDsn = configs.DatabaseDSN
	}
	if secretKey == "" {
		secretKey = configs.SecretKey
	}
	if expireToken == 0 {
		expireToken = configs.ExpireToken
	}
}

// parceEnvironment - функция для переопределения конфигурации из глобальных переменных.
// Переопределяет конфигурацию, если значения не установлены флагами или файлом конфигурации.
func parseEnvironment() {
	if netAddr == "" {
		netAddr = os.Getenv("GOPHKEEPER_SERVER_ADDRESS")
	}
	if databaseDsn == "" {
		databaseDsn = os.Getenv("GOPHKEEPER_SERVER_DATABASE_URL")
	}
	if logLevel == "" {
		logLevel = os.Getenv("GOPHKEEPER_SERVER_LOG_LEVEL")
	}
	if secretKey == "" {
		secretKey = os.Getenv("GOPHKEEPER_SERVER_SECRET_KEY")
	}
	if expireToken == 0 {
		envExpireToken := os.Getenv("GOPHKEEPER_SERVER_EXPIRE_TOKEN")
		if envExpireToken != "" {
			expire, err := strconv.Atoi(envExpireToken)
			if err == nil {
				expireToken = expire
			}
		}
	}
}

// checkVariables - функция для проверки корректности утсановки глобальных переменных.
func checkVariables() error {
	if netAddr == "" {
		return fmt.Errorf("address and port to run server must be set")
	}
	if logLevel == "" {
		return fmt.Errorf("log level must be set")
	}
	if databaseDsn == "" {
		return fmt.Errorf("database connection address must be set")
	}
	if secretKey == "" {
		return fmt.Errorf("secret key must be set")
	}
	if expireToken == 0 {
		return fmt.Errorf("expire token must be set")
	}
	return nil
}
