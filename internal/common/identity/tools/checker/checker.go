package checker

// CheckLogin - функция для проверки корректности логина.
func CheckLogin(login string) bool {
	// проверяю, что логин не является пустой строкой
	return login != ""
}

// CheckPassword - функция для проверки корректности пароля.
func CheckPassword(password string) bool {
	// проверяю, что пароль не является пустой строкой
	return password != ""
}

// CheckHash - функция для проверки корректности хэша.
func CheckHash(hash string) bool {
	// проверяю, что хэш не является пустой строкой
	return hash != ""
}

// IsAuthorize - функция для проверки совпадения авторизационных данных пользователя.
func IsAuthorize(wanrHash, getHash string) bool {
	return wanrHash == getHash
}
