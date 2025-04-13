package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MockResponseWriter struct {
	HeaderMap http.Header
	Status    int
}

func (m *MockResponseWriter) Header() http.Header {
	if m.HeaderMap == nil {
		m.HeaderMap = make(http.Header)
	}
	return m.HeaderMap
}

func (m *MockResponseWriter) Write(body []byte) (int, error) {
	return len(body), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.Status = statusCode
}

func TestInitialize(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectError bool
	}{
		{"ValidDebugLevel", "debug", false},
		{"ValidInfoLevel", "info", false},
		{"ValidWarnLevel", "warn", false},
		{"InvalidLevel", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Инициализирую логгер
			err := Initialize(tt.level)

			// Проверяю наличие ошибки
			if tt.expectError {
				require.Error(t, err)
			}
			if !tt.expectError {
				require.NoError(t, err)
			}

			// Если ошибок нет, проверяю уровень логирования
			if !tt.expectError {
				require.NotEqual(t, nil, ServerLog)

				// получаю текущий уровень логгера
				level := ServerLog.Core().Enabled(zap.DebugLevel)
				expectedLevel := tt.level == "debug" // уровень "debug" должен быть доступен только при debug
				require.Equal(t, expectedLevel, level)
			}
		})
	}
}

func TestInitializeInvalidConfig(t *testing.T) {
	// Создаю буфер для вывода логов
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zap.DebugLevel,
	)
	ServerLog = zap.New(core)

	// Инициализация с некорректным уровнем
	err := Initialize("invalid")
	require.Error(t, err)

	// Проверяю, что глобальный логгер не был перезаписан
	require.Equal(t, false, buf.Len() > 0)
}

func TestWrite(t *testing.T) {
	responseData := &responseData{}

	testHeaderMap := http.Header{"test_header1": []string{"test_value1"}}

	mockWriter := &MockResponseWriter{
		HeaderMap: testHeaderMap,
		Status:    200,
	}

	loggingResponseWriter := loggingResponseWriter{
		ResponseWriter: mockWriter,
		responseData:   responseData,
	}

	firstMessage := []byte("first message")
	lenFirstMessage, err := loggingResponseWriter.Write(firstMessage)
	require.NoError(t, err)
	assert.Equal(t, len(firstMessage), lenFirstMessage)
	assert.Equal(t, len(firstMessage), responseData.size)

	secondMessage := []byte("write second message")
	lenSecondMessage, err := loggingResponseWriter.Write(secondMessage)
	require.NoError(t, err)
	assert.Equal(t, len(secondMessage), lenSecondMessage)
	assert.Equal(t, len(firstMessage)+len(secondMessage), responseData.size)
}

func TestWriteHeader(t *testing.T) {
	responseData := &responseData{}

	testHeaderMap := http.Header{"test_header1": []string{"test_value1"}}

	mockWriter := &MockResponseWriter{
		HeaderMap: testHeaderMap,
		Status:    200,
	}

	loggingResponseWriter := loggingResponseWriter{
		ResponseWriter: mockWriter,
		responseData:   responseData,
	}

	firstStatusCode := 300
	loggingResponseWriter.WriteHeader(firstStatusCode)
	assert.Equal(t, firstStatusCode, mockWriter.Status)
	assert.Equal(t, firstStatusCode, responseData.status)

	secondStatusCode := 500
	loggingResponseWriter.WriteHeader(secondStatusCode)
	assert.Equal(t, secondStatusCode, mockWriter.Status)
	assert.Equal(t, secondStatusCode, responseData.status)
}

func TestRequestLogger(t *testing.T) {

	testHandler := func() http.HandlerFunc {
		return func(res http.ResponseWriter, _ *http.Request) {
			res.WriteHeader(200)
		}
	}

	r := chi.NewRouter()
	r.Post("/", RequestLogger(testHandler()))

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close() // Закрываем тело ответа
	// проверяем код ответа
	assert.Equal(t, 200, res.StatusCode)
}
