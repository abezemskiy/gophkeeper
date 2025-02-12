package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateId(t *testing.T) {
	id1, err := GenerateID()
	require.NoError(t, err)

	id2, err := GenerateID()
	require.NoError(t, err)
	assert.NotEqual(t, id1, id2)
}
