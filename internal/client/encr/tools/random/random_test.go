package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCryptoRandom(t *testing.T) {
	testSize := 32
	r1, err := GenerateCryptoRandom(testSize)
	require.NoError(t, err)

	r2, err := GenerateCryptoRandom(testSize)
	require.NoError(t, err)

	assert.NotEqual(t, r1, r2)
}
