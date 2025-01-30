package checker

import (
	"gophkeeper/internal/common/identity/tools/hasher"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckLogin(t *testing.T) {
	assert.Equal(t, true, CheckLogin("not empty"))
	assert.Equal(t, false, CheckLogin(""))
}

func TestCheckPassword(t *testing.T) {
	assert.Equal(t, true, CheckPassword("not empty"))
	assert.Equal(t, false, CheckPassword(""))
}

func TestCheckHash(t *testing.T) {
	assert.Equal(t, true, CheckHash("not empty"))
	assert.Equal(t, false, CheckHash(""))
}

func TestIsAuthorize(t *testing.T) {
	{
		// true test
		wantHash, err := hasher.CalkHash("some string like login")
		require.NoError(t, err)
		assert.Equal(t, true, IsAuthorize(wantHash, wantHash))
	}
	{
		// hashs is not equal
		wantHash, err := hasher.CalkHash("some string like login")
		require.NoError(t, err)
		assert.Equal(t, false, IsAuthorize(wantHash, "wrong hash"))
	}
	{
		// passwords and hashs is not equal
		wantHash, err := hasher.CalkHash("some string like login")
		require.NoError(t, err)
		assert.Equal(t, false, IsAuthorize(wantHash, "wrong hash"))
	}
}
