package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abezemskiy/gophkeeper/internal/common/identity/tools/token"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	testHandler := func(idWant string) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			// извлекаю id пользователя из контекста
			id, ok := req.Context().Value(UserIDKey).(string)
			require.Equal(t, true, ok)

			// проверяю, что id полученный из контекста совпадает с ожидаемым
			assert.Equal(t, idWant, id)

			res.WriteHeader(http.StatusOK)
		}
	}

	// Test. successful authentication ---------------------------------------
	successKey := "success secret key"
	token.SetSecretKey(successKey)
	token.SerExpireHour(1)
	idSuccess := "success id"
	tokenSuccess, err := token.BuildJWT(idSuccess)
	require.NoError(t, err)

	// Test error. token is expires ---------------------------------------
	token.SerExpireHour(-1)
	tokenExpired, err := token.BuildJWT("")
	require.NoError(t, err)

	type request struct {
		token       string
		key         string
		tokenExpire int
		setheader   bool
	}
	type want struct {
		id     string
		status int
	}
	tests := []struct {
		name string
		req  request
		want want
	}{
		{
			name: "successful authentication",
			req: request{
				token:       tokenSuccess,
				key:         successKey,
				tokenExpire: 1,
				setheader:   true,
			},
			want: want{
				id:     idSuccess,
				status: 200,
			},
		},
		{
			name: "header is not set",
			req: request{
				token:       "",
				key:         "",
				tokenExpire: 0,
				setheader:   false,
			},
			want: want{
				id:     "",
				status: 401,
			},
		},
		{
			name: "token is expried",
			req: request{
				token:       tokenExpired,
				key:         "",
				tokenExpire: -1,
				setheader:   true,
			},
			want: want{
				id:     "",
				status: 401,
			},
		},
		{
			name: "wrong token",
			req: request{
				token:       "wrong token",
				key:         successKey,
				tokenExpire: 1,
				setheader:   true,
			},
			want: want{
				id:     idSuccess,
				status: 401,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// устанавливаю секретный ключ для сервера
			token.SetSecretKey(tt.req.key)

			r := chi.NewRouter()
			r.Get("/test", Middleware(testHandler(tt.want.id)))

			request := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.req.setheader {
				// устанавливаю заголовк с токеном в запрос
				request.Header.Set("Authorization", "Bearer "+tt.req.token)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.status, result.StatusCode)
		})
	}
}
