package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfigFile(t *testing.T) {
	testFlagNetAddr := "localhost:8082"
	testFlagDatabaseDsn := "test dsn"
	testFlagLogLevel := "test info"
	testSecretKey := "test secret key"
	testExpireToken := 30

	createFile := func(name string) {
		data := fmt.Sprintf("{\"address\": \"%s\",\"database_dsn\": \"%s\",\"log_level\": \"%s\", \"secret_key\":\"%s\", \"expire_token\":%d}",
			testFlagNetAddr, testFlagDatabaseDsn, testFlagLogLevel, testSecretKey, testExpireToken)
		f, err := os.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(data))
		require.NoError(t, err)
	}
	nameFile := "./test_config.json"
	createFile(nameFile)

	configs, err := ParseConfigFile(nameFile)
	require.NoError(t, err)

	assert.Equal(t, testFlagNetAddr, configs.Address)
	assert.Equal(t, testFlagDatabaseDsn, configs.DatabaseDSN)
	assert.Equal(t, testFlagLogLevel, configs.LogLevel)
	assert.Equal(t, testSecretKey, configs.SecretKey)
	assert.Equal(t, testExpireToken, configs.ExpireToken)

	err = os.Remove(nameFile)
	require.NoError(t, err)
}
