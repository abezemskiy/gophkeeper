package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalkHash(t *testing.T) {
	// Тест с повторным хэшированием одной и той-же строки и проверкой повторяемости результата
	data := "some string for hashing"
	res1, err := CalkHash(data)
	require.NoError(t, err)

	for range 10 {
		res2, err := CalkHash(data)
		require.NoError(t, err)
		assert.Equal(t, res1, res2)
	}

	// хэширую другую строку и проверяю, что хэши не совпадают
	data2 := "some different string for hashing"
	res2, err := CalkHash(data2)
	require.NoError(t, err)
	assert.NotEqual(t, res1, res2)
}
