package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// Configs представляет структуру конфигурации.
type Configs struct {
	Address     string `json:"address"`      // аналог переменной окружения GOPHKEEPER_CLIENT_ADDRESS или флага -a
	LogLevel    string `json:"log_level"`    // аналог переменной окружения GOPHKEEPER_CLIENT_LOG_LEVEL или флага -l
	DatabaseDSN string `json:"database_dsn"` // аналог переменной окружения GOPHKEEPER_CLIENT_DATABASE_URL или флага -d
}

// ParseConfigFile - функция для переопределения параметров конфигурации из файла конфигурации.
func ParseConfigFile(configFileNAme string) (Configs, error) {
	var configs Configs
	f, err := os.Open(configFileNAme)
	if err != nil {
		return Configs{}, fmt.Errorf("open cofiguration file error: %w", err)
	}
	reader := bufio.NewReader(f)
	dec := json.NewDecoder(reader)
	err = dec.Decode(&configs)
	if err != nil {
		return Configs{}, fmt.Errorf("parse cofiguration file error: %w", err)
	}

	return configs, nil
}
