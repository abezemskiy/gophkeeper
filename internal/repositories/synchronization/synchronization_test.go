package synchronization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPeroidOfSynchr(t *testing.T) {
	getPeriod := GetPeroidOfSynchr()
	assert.Equal(t,PeroidOfSynchr, getPeriod)
}
