package info

import (
	"gophkeeper/internal/client/identity"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	info := NewUserInfoStorage()
	assert.NotEqual(t, nil, info)

	login := "some login"
	password := "some password"
	id := "some id"
	info.Set(identity.AuthData{
		Login:    login,
		Password: password,
	}, id)

	assert.Equal(t, login, info.authData.Login)
	assert.Equal(t, password, info.authData.Password)
	assert.Equal(t, id, info.id)
}

func TestGet(t *testing.T) {
	info := NewUserInfoStorage()
	assert.NotEqual(t, nil, info)

	login := "some login"
	password := "some password"
	id := "some id"
	data := identity.AuthData{
		Login:    login,
		Password: password,
	}

	info.authData = data
	info.id = id

	getAuth, getID := info.Get()
	assert.Equal(t, login, getAuth.Login)
	assert.Equal(t, password, getAuth.Password)
	assert.Equal(t, id, getID)
}
