package main

import (
	"flag"
	"fmt"
	"gophkeeper/internal/client/storage/inmemory"
	"gophkeeper/internal/server/config"
	"log"
	"os"
)

var (
	netAddr     string // адрес запуска сервиса
	databaseDsn string // адрес базы данных
	logLevel    string // уровень логирования
	configFile  string // путь к файлу конфигурации
)

// parseVariables - функция для установки конфигурационных параметров приложения.
// Конфигурирование приложения с приоритетом в порядке убывания: значения флагов, значения из файла, значения переменных окружения.
func parseVariables() error {
	parseFlags()
	parseConfigFile()
	parseEnvironment()

	// Проверка корректности установки глобальных переменных
	err := checkVariables()
	if err != nil {
		return err
	}

	// устанавливаю время обновления данных каждые 2 секунды
	inmemory.SetUpdatingPeriod(5)
	return nil
}

// parseFlags - функция для определения параметров конфигурации из флагов.
func parseFlags() {
	flag.StringVar(&netAddr, "a", "", "address and port to run client")

	// настройка флага для хранения метрик в базе данных
	flag.StringVar(&databaseDsn, "d", "", "database connection address") // по умолчанию адрес не задан

	flag.StringVar(&logLevel, "l", "", "log level")
	flag.StringVar(&configFile, "c", "", "name of configuration file")

	// Вызов flag.Parse() для парсинга аргументов
	flag.Parse()
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
}

// parceEnvironment - функция для переопределения конфигурации из глобальных переменных.
// Переопределяет конфигурацию, если значения не установлены флагами или файлом конфигурации.
func parseEnvironment() {
	if netAddr == "" {
		netAddr = os.Getenv("GOPHKEEPER_CLIENT_ADDRESS")
	}
	if databaseDsn == "" {
		databaseDsn = os.Getenv("GOPHKEEPER_CLIENT_DATABASE_URL")
	}
	if logLevel == "" {
		logLevel = os.Getenv("GOPHKEEPER_CLIENT_LOG_LEVEL")
	}
}

// checkVariables - функция для проверки корректности утсановки глобальных переменных.
func checkVariables() error {
	if netAddr == "" {
		return fmt.Errorf("address and port to run client must be set")
	}
	if logLevel == "" {
		return fmt.Errorf("log level must be set")
	}
	if databaseDsn == "" {
		return fmt.Errorf("database connection address must be set")
	}
	return nil
}
