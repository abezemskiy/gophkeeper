package key

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveKey(t *testing.T) {
	{
		pass := "some strong password"
		lenthKey := 32
		key1 := DeriveKey(pass, lenthKey)
		assert.Equal(t, 32, len(key1))

		key2 := DeriveKey(pass, lenthKey)
		assert.Equal(t, 32, len(key2))

		assert.Equal(t, key1, key2)
	}
}
