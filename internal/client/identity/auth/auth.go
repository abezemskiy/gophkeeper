package auth

import (
	"fmt"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/common/identity/tools/hasher"
	"gophkeeper/internal/common/identity/tools/header"
	repoIdent "gophkeeper/internal/repositories/identity"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// OnBeforeMiddleware - мидлварь для установки аутентификационных пользователя перед отправкой запроса на сервер.
func OnBeforeMiddleware(info identity.IUserInfoStorage, ident identity.ClientIdentifier) resty.RequestMiddleware {
	return func(c *resty.Client, req *resty.Request) error {
		// Извлекаю идентификационные данные текущего пользователя
		authData, _ := info.Get()

		// Извлекаю идентификационные данные пользователя, в том числе и токен из хранилища
		infoFromStor, ok, err := ident.Authorize(req.Context(), authData.Login)
		if err != nil {
			return fmt.Errorf("failed to get token from storage of user %s, %w", authData.Login, err)
		}
		if !ok {
			// Пользователь с данным логином не найден
			return fmt.Errorf("user %s not register", authData.Login)
		}

		// Устанавливаю токен в заголовок запроса
		req.Header.Set("Authorization", "Bearer "+infoFromStor.Token)

		// Заголовок для авторизации пользователя успешно установлен
		return nil
	}
}

// OnAfterMiddleware - мидлварь для обновления авторизационных данных пользователя на случай, если сервер вернет статус 401.
// Статус 401 может возникнуть по причине истечения срока действия JWT. Из-за хранения токена в локальном хранилище требуется ручное обновление.
func OnAfterMiddleware(info identity.IUserInfoStorage, ident identity.ClientIdentifier, authURL string) resty.ResponseMiddleware {
	return func(c *resty.Client, res *resty.Response) error {
		// Проверяю статус ответа сервера
		if res.StatusCode() == http.StatusUnauthorized {
			// Пользователь не авторизирован, пробую обновить авторизационные данные

			// Извлекаю авторизационные данные пользователя из хранилища
			authData, _ := info.Get()

			// Вычисляю хэш пары логин-пароль
			hash, err := hasher.CalkHash(authData.Login + authData.Password)
			if err != nil {
				return fmt.Errorf("failed to calculate hash, %w", err)
			}

			// Отправляю запрос на авторизацию пользователя на сервере
			resp, err := c.R().
				SetHeader("Content-Type", "application/json").
				SetBody(repoIdent.IdentityData{
					Login: authData.Login,
					Hash:  hash,
				}).
				Post(authURL)

			if err != nil {
				return fmt.Errorf("failed to post authorization request to server, %w", err)
			}

			// Проверяю ответ сервера, если статус код == 200, то обновляю токен в локальном хранилище.
			if resp.StatusCode() == http.StatusOK {
				// извлекаю новый токен из заголовка ответа сервера
				newToken, err := header.GetTokenFromRestyResponseHeader(resp)
				if err != nil {
					return fmt.Errorf("failed to get token from server responce, %w", err)
				}

				// Обновляю токен в локальном хранилище
				ok, err := ident.SetToken(res.Request.Context(), authData.Login, newToken)
				if err != nil {
					return fmt.Errorf("failed to set new token for user %s, %w", authData.Login, err)
				}
				if !ok {
					return fmt.Errorf("user %s not register", authData.Login)
				}
			} else {
				// Если статус ответа сервера другой, то заменяю ответ клиенту после оригинально запроса на ответ,
				// который был получен при попытке авторизации на сервер.
				*res = *resp
			}
		}
		return nil
	}
}
