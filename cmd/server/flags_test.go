package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetVariables() {
	netAddr = ""
	databaseDsn = ""
	logLevel = ""
	configFile = ""
}

func TestParseFlags(t *testing.T) {
	// Сбрасываю значения переменных перед и после тестирования
	resetVariables()
	defer resetVariables()

	os.Args = []string{"cmd", "-a", ":9000", "-l", "debug", "-d", "db_dsn", "-c", "/config/file"}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	parseFlags()

	assert.Equal(t, ":9000", netAddr)
	assert.Equal(t, "debug", logLevel)
	assert.Equal(t, "db_dsn", databaseDsn)
	assert.Equal(t, "/config/file", configFile)
}

func TestParseFlagsPriority(t *testing.T) {
	// Сбрасываю значения переменных перед и после тестирования
	resetVariables()
	defer resetVariables()

	// Устанавливаю переменные окружения
	os.Setenv("GOPHKEEPER_SERVER_ADDRESS", "env_url")
	os.Setenv("GOPHKEEPER_SERVER_DATABASE_URL", "env_dsn")
	os.Setenv("GOPHKEEPER_SERVER_LOG_LEVEL", "env_info")

	defer func() {
		os.Unsetenv("GOPHKEEPER_SERVER_ADDRESS")
		os.Unsetenv("GOPHKEEPER_SERVER_DATABASE_URL")
		os.Unsetenv("GOPHKEEPER_SERVER_LOG_LEVEL")
	}()

	// Создаю временный конфигурационный файл
	testConfigFile := "./test_config.json"
	configContent := `{
        "address": "file_url",
		"log_level": "file_debug",
		"database_dsn": "file_dsn",
    }`
	err := os.WriteFile(testConfigFile, []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove(testConfigFile)

	// Устанавливаю значения флагов
	os.Args = []string{"cmd", "-a", "flag_url", "-l", "flag_info", "-d", "flag_dsn", "-c", testConfigFile}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	parseFlags()

	// Флаги имеют приоритет
	assert.Equal(t, "flag_url", netAddr)
	assert.Equal(t, "flag_info", logLevel)
	assert.Equal(t, "flag_dsn", databaseDsn)
	assert.Equal(t, configFile, testConfigFile)
}

func TestParseEnvironment(t *testing.T) {
	// Сбрасываю значения переменных перед и после тестирования
	resetVariables()
	defer resetVariables()

	// Устанавливаем переменные окружения
	os.Setenv("GOPHKEEPER_SERVER_ADDRESS", ":8000")
	os.Setenv("GOPHKEEPER_SERVER_DATABASE_URL", "env_dsn")
	os.Setenv("GOPHKEEPER_SERVER_LOG_LEVEL", "test_info")

	defer func() {
		os.Unsetenv("GOPHKEEPER_SERVER_ADDRESS")
		os.Unsetenv("GOPHKEEPER_SERVER_DATABASE_URL")
		os.Unsetenv("GOPHKEEPER_SERVER_LOG_LEVEL")
	}()

	parseEnvironment()

	assert.Equal(t, ":8000", netAddr)
	assert.Equal(t, "test_info", logLevel)
	assert.Equal(t, "env_dsn", databaseDsn)
}

func TestParseConfigFile(t *testing.T) {
	// Сбрасываю значения переменных перед и после тестирования
	resetVariables()
	defer resetVariables()

	testFlagNetAddr := "localhost:8082"
	testFlagLogLevel := "info"
	testFlagDatabaseDsn := "test dsn"

	createFile := func(name string) {
		data := fmt.Sprintf("{\"address\": \"%s\",\"log_level\": \"%s\",\"database_dsn\": \"%s\"}",
			testFlagNetAddr, testFlagLogLevel, testFlagDatabaseDsn)
		f, err := os.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(data))
		require.NoError(t, err)
	}
	nameFile := "./test_config.json"
	createFile(nameFile)

	// Утсанавливаю путь к файлу конфигурации
	configFile = nameFile
	parseConfigFile()

	assert.Equal(t, testFlagNetAddr, netAddr)
	assert.Equal(t, testFlagLogLevel, logLevel)
	assert.Equal(t, testFlagDatabaseDsn, databaseDsn)

	err := os.Remove(nameFile)
	require.NoError(t, err)
}

func TestCheckVariables(t *testing.T) {
	// Сбрасываю значения переменных перед и после тестирования
	resetVariables()
	defer resetVariables()

	err := checkVariables()
	require.Error(t, err)

	netAddr = "some addr"
	err = checkVariables()
	require.Error(t, err)

	logLevel = "some level"
	err = checkVariables()
	require.Error(t, err)

	databaseDsn = "some dsn"
	err = checkVariables()
	require.NoError(t, err)
}
