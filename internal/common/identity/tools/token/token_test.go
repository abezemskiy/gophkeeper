package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSecretKey(t *testing.T) {
	testKey := "test key"
	SetSecretKey(testKey)
	assert.Equal(t, testKey, secretKey)
}

func TestSerExpireHour(t *testing.T) {
	testExpire := 235235
	SerExpireHour(testExpire)
	assert.Equal(t, testExpire, expireHour)
}

func TestBuildJWT(t *testing.T) {
	SetSecretKey("test key")

	// генерирую токен
	id := "41614361346161346"
	token, err := BuildJWT(id)
	require.NoError(t, err)

	// получаю id из токена
	getID, err := GetIDFromToken(token)
	require.NoError(t, err)
	assert.Equal(t, id, getID)

	// генерирую новый токен
	id2 := "527274747542747"
	token2, err := BuildJWT(id2)
	require.NoError(t, err)

	// получаю id из токена
	getID2, err := GetIDFromToken(token2)
	require.NoError(t, err)
	assert.Equal(t, id2, getID2)
	assert.NotEqual(t, getID, getID2)

	// тест с ошибкой. При попытке извлечь id из токена устанавливаю неверный секретный ключ
	SetSecretKey("wrong key")
	_, err = GetIDFromToken(token2)
	require.Error(t, err)
}
