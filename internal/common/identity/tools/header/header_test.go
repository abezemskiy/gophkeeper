package header

import (
	"gophkeeper/internal/common/identity/tools/token"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTokenFromHeader(t *testing.T) {
	{
		// Тест с успешныыи извлечение заголовка
		r := httptest.NewRequest("POST", "/header", nil)
		id := "254735724613466"
		tokenBuild, err := token.BuildJWT(id)
		require.NoError(t, err)

		r.Header.Set("Authorization", "Bearer "+tokenBuild)

		res, err := GetTokenFromHeader(r)
		require.NoError(t, err)
		// извлекаю id из извлеченного токена
		getID, err := token.GetIDFromToken(res)
		require.NoError(t, err)
		assert.Equal(t, id, getID)
	}
	{
		// Тест с неправильным ключом заголовка
		r := httptest.NewRequest("POST", "/header", nil)
		id := "254735724613466"
		tokenBuild, err := token.BuildJWT(id)
		require.NoError(t, err)

		r.Header.Set("Wrong header", "Bearer "+tokenBuild)

		_, err = GetTokenFromHeader(r)
		require.Error(t, err)
	}
	{
		// Тест с неправильным форматом заголовка
		r := httptest.NewRequest("POST", "/header", nil)
		id := "254735724613466"
		tokenBuild, err := token.BuildJWT(id)
		require.NoError(t, err)

		r.Header.Set("Authorization", "Wrong format "+tokenBuild)

		_, err = GetTokenFromHeader(r)
		require.Error(t, err)
	}
}

func TestGetTokenFromResponseHeader(t *testing.T) {
	testHandler := func(id, key, format string) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// генерирую токен
			token, err := token.BuildJWT(id)
			require.NoError(t, err)

			// устанавливаю токен в заголовок
			res.Header().Set(key, format+" "+token)
		}
	}

	type request struct {
		id     string
		key    string
		format string
	}
	type want struct {
		err bool
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "success test",
			req: request{
				id:     "2125125235",
				key:    "Authorization",
				format: "Bearer",
			},
			want: want{
				err: false,
			},
		},
		{
			name: "bad header",
			req: request{
				id:     "2125125235",
				key:    "Wrong Header",
				format: "Bearer",
			},
			want: want{
				err: true,
			},
		},
		{
			name: "bad format",
			req: request{
				id:     "2125125235",
				key:    "Authorization",
				format: "Wrong format",
			},
			want: want{
				err: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Post("/header", testHandler(tt.req.id, tt.req.key, tt.req.format))

			request := httptest.NewRequest(http.MethodPost, "/header", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			// извлекаю полученный от сервера токен
			getToken, err := GetTokenFromResponseHeader(result)

			if tt.want.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// извлекаю id из полученного от сервера токена
				getID, err := token.GetIDFromToken(getToken)
				require.NoError(t, err)
				assert.Equal(t, tt.req.id, getID)
			}
		})
	}
}
