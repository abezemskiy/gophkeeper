package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophkeeper/internal/common/identity/tools/checker"
	"gophkeeper/internal/common/identity/tools/id"
	"gophkeeper/internal/common/identity/tools/token"
	"gophkeeper/internal/repositories/data"
	"gophkeeper/internal/repositories/identity"
	"gophkeeper/internal/server/identity/auth"
	"gophkeeper/internal/server/logger"
	"gophkeeper/internal/server/storage"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// Register - хэндлер для регистрации пользователя в системе. Если пользователь успешно зарегистрирован, то в заголовок ответа устанавливается
// токен пользователя.
func Register(res http.ResponseWriter, req *http.Request, ident identity.Identifier) {
	res.Header().Set("Content-Type", "text/plain")
	defer req.Body.Close()

	var regData identity.IdentityData
	if err := json.NewDecoder(req.Body).Decode(&regData); err != nil {
		logger.ServerLog.Error("failed to parse identity data to structer", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
		http.Error(res, fmt.Errorf("failed to parse identity data to structer, %w", err).Error(), http.StatusBadRequest)
		return
	}

	// Проверяю корректность логина
	if ok := checker.CheckLogin(regData.Login); !ok {
		logger.ServerLog.Error("login is not valid", zap.String("address", req.URL.String()))
		http.Error(res, "login is not valid", http.StatusBadRequest)
		return
	}
	// Проверяю корректность хэша от суммы логин+пароль
	if ok := checker.CheckHash(regData.Hash); !ok {
		logger.ServerLog.Error("hash is not valid", zap.String("address", req.URL.String()))
		http.Error(res, "hash is not valid", http.StatusBadRequest)
		return
	}

	// вычисляю идентификатор пользователя
	id, err := id.GenerateId()
	if err != nil {
		logger.ServerLog.Error("failed to generate id", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
		http.Error(res, fmt.Errorf("failed to generate id, %w", err).Error(), http.StatusInternalServerError)
		return
	}

	// Регистрирую пользователя в хранилище
	err = ident.Register(req.Context(), regData.Login, regData.Hash, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// пользователь с данным логином уже зарегистрирован в системе
			logger.ServerLog.Error(fmt.Sprintf("login %s already exists", regData.Login), zap.String("address", req.URL.String()))
			http.Error(res, fmt.Errorf("login %s already exists, %w", regData.Login, err).Error(), http.StatusConflict)
		} else {
			logger.ServerLog.Error("register user error", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
			http.Error(res, fmt.Errorf("register user error, %w", err).Error(), http.StatusInternalServerError)
		}
		return
	}

	// При успешной регистрации создаю токен и устанавливаю токен в заголовок
	// генерирую токен
	token, err := token.BuildJWT(id)
	if err != nil {
		logger.ServerLog.Error("build JWT error", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
		http.Error(res, fmt.Errorf("build JWT error, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	// устанавливаю токен в заголовок
	res.Header().Set("Authorization", "Bearer "+token)
	res.WriteHeader(200)
}

func RegisterHandler(ident identity.Identifier) http.HandlerFunc {
	fn := func(res http.ResponseWriter, req *http.Request) {
		Register(res, req, ident)
	}
	return fn
}

// Authorize - хэндлер для авторизации пользователя в системе. Если пользователь авторизирован, то в заголовок ответа устанавливается
// токен пользователя.
func Authorize(res http.ResponseWriter, req *http.Request, ident identity.Identifier) {
	res.Header().Set("Content-Type", "text/plain")
	res.Header()
	defer req.Body.Close()

	var regData identity.IdentityData
	if err := json.NewDecoder(req.Body).Decode(&regData); err != nil {
		logger.ServerLog.Error("failed to parse identity data to structer", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
		http.Error(res, fmt.Errorf("failed to parse identity data to structer, %w", err).Error(), http.StatusBadRequest)
		return
	}

	// Проверяю корректность логина
	if ok := checker.CheckLogin(regData.Login); !ok {
		logger.ServerLog.Error(fmt.Sprintf("login %s is not valid", regData.Login), zap.String("address", req.URL.String()))
		http.Error(res, fmt.Sprintf("login %s is not valid", regData.Login), http.StatusBadRequest)
		return
	}
	// Проверяю корректность хэша
	if ok := checker.CheckHash(regData.Hash); !ok {
		logger.ServerLog.Error(fmt.Sprintf("hash %s is not valid", regData.Hash), zap.String("address", req.URL.String()))
		http.Error(res, fmt.Sprintf("hash %s is not valid", regData.Hash), http.StatusBadRequest)
		return
	}

	// Получаю авторизационные данные пользователя из хранилища
	data, ok, err := ident.Authorize(req.Context(), regData.Login)
	if err != nil {
		// внутренняя ошибка сервера
		logger.ServerLog.Error("authorize user error", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
		http.Error(res, fmt.Errorf("authorize user error, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		// не найдено записей по представленному логину. Пользователь не зарегистрирован.
		logger.ServerLog.Error(fmt.Sprintf("user %s not register", regData.Login), zap.String("address", req.URL.String()))
		http.Error(res, fmt.Sprintf("user %s not register", regData.Login), http.StatusBadRequest)
		return
	}

	// проверяю что хэш пары логин+пароль отправленный пользователем для авторизации совпадает с тем, что хранится в хранилище.
	if !checker.IsAuthorize(data.Hash, regData.Hash) {
		logger.ServerLog.Error("password is wrong", zap.String("address", req.URL.String()))
		http.Error(res, fmt.Errorf("password is wrong, %w", err).Error(), http.StatusBadRequest)
		return
	}

	// При успешной авторизации создаю токен и устанавливаю токен в заголовок
	// генерирую токен
	token, err := token.BuildJWT(data.ID)
	if err != nil {
		logger.ServerLog.Error("build JWT error", zap.String("address", req.URL.String()), zap.String("error", error.Error(err)))
		http.Error(res, fmt.Errorf("build JWT error, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	// устанавливаю токен в заголовок
	res.Header().Set("Authorization", "Bearer "+token)
	res.WriteHeader(200)
}

// AuthorizeHandler - обертка на функцией Authorize.
func AuthorizeHandler(ident identity.Identifier) http.HandlerFunc {
	fn := func(res http.ResponseWriter, req *http.Request) {
		Authorize(res, req, ident)
	}
	return fn
}

// AddEncryptedData - хэндлер для загрузки новых зашифрованных данных в хранилище.
func AddEncryptedData(res http.ResponseWriter, req *http.Request, stor storage.IEncryptedServerStorage) {
	// получаю id пользователя из контекста
	id, ok := req.Context().Value(auth.UserIDKey).(string)
	if !ok {
		logger.ServerLog.Error("user ID not found in context", zap.String("address", req.URL.String()))
		http.Error(res, "user ID not found in context", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	// Сериализую данные из запроса клиента
	var data data.EncryptedData
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		logger.ServerLog.Error("can't parse data from request", zap.String("address", req.URL.String()), zap.String("error", err.Error()))
		http.Error(res, "can't parse data from request", http.StatusInternalServerError)
		return
	}

	// Добавляю новые данные в хранилище
	ok, err := stor.AddEncryptedData(req.Context(), id, data)
	if err != nil {
		logger.ServerLog.Error("adding data to storage error", zap.String("address", req.URL.String()), zap.String("error", err.Error()))
		http.Error(res, fmt.Errorf("adding data to storage error, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		logger.ServerLog.Error("data is already exist", zap.String("address", req.URL.String()))
		http.Error(res, "data is already exist", http.StatusConflict)
		return
	}

	res.WriteHeader(http.StatusOK)
	logger.ServerLog.Debug("successful write encode data to storage")
}

// AuthorizeHandler - обертка над AddEncryptedData.
func AddEncryptedDataHandler(stor storage.IEncryptedServerStorage) http.HandlerFunc {
	fn := func(res http.ResponseWriter, req *http.Request) {
		AddEncryptedData(res, req, stor)
	}
	return fn
}

// ReplaceEncryptedData - хэндлер для замены старых данных значениями новых.
// В случае попытки заменить данные, когда данные с текущим id полязователя и именем ещё не загружены в хранилище
// возвращается ошибка.
func ReplaceEncryptedData(res http.ResponseWriter, req *http.Request, stor storage.IEncryptedServerStorage) {
	// получаю id пользователя из контекста
	id, ok := req.Context().Value(auth.UserIDKey).(string)
	if !ok {
		logger.ServerLog.Error("user ID not found in context", zap.String("address", req.URL.String()))
		http.Error(res, "user ID not found in context", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	// Сериализую данные из запроса клиента
	var newData data.EncryptedData
	if err := json.NewDecoder(req.Body).Decode(&newData); err != nil {
		logger.ServerLog.Error("can't parse data from request", zap.String("address", req.URL.String()), zap.String("error", err.Error()))
		http.Error(res, "can't parse data from request", http.StatusInternalServerError)
		return
	}

	// заменяю старые данные новыми в хранилище
	ok, err := stor.ReplaceEncryptedData(req.Context(), id, newData)
	if err != nil {
		logger.ServerLog.Error("replace data in storage error", zap.String("address", req.URL.String()), zap.String("error", err.Error()))
		http.Error(res, fmt.Errorf("replace data in storage error, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		logger.ServerLog.Error("data does not exist", zap.String("address", req.URL.String()))
		http.Error(res, "data does not exist", http.StatusNotFound)
		return
	}

	res.WriteHeader(http.StatusOK)
	logger.ServerLog.Debug("successful replace encode data in storage")
}

// ReplaceEncryptedDataHandler - обертка над ReplaceEncryptedData.
func ReplaceEncryptedDataHandler(stor storage.IEncryptedServerStorage) http.HandlerFunc {
	fn := func(res http.ResponseWriter, req *http.Request) {
		ReplaceEncryptedData(res, req, stor)
	}
	return fn
}

// GetAllEncryptedData - хэндлер для отправки пользователю всех его данных батчем.
func GetAllEncryptedData(res http.ResponseWriter, req *http.Request, stor storage.IEncryptedServerStorage) {
	// получаю id пользователя из контекста
	id, ok := req.Context().Value(auth.UserIDKey).(string)
	if !ok {
		logger.ServerLog.Error("user ID not found in context", zap.String("address", req.URL.String()))
		http.Error(res, "user ID not found in context", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	allData, err := stor.GetAllEncryptedData(req.Context(), id)
	if err != nil{
		logger.ServerLog.Error("get all data from storage error", zap.String("address", req.URL.String()), zap.String("error", err.Error()))
		http.Error(res, fmt.Errorf("get all data from storage error, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(res)
	if err := enc.Encode(allData); err != nil {
		logger.ServerLog.Error("encoding response error", zap.String("error", error.Error(err)))
		http.Error(res, fmt.Errorf("encoding response error, %w", err).Error(), http.StatusInternalServerError)
		return
	}
	logger.ServerLog.Debug("successful return all encrypted data to client")
}

// GetAllEncryptedDataHandler - обертка над GetAllEncryptedData.
func GetAllEncryptedDataHandler(stor storage.IEncryptedServerStorage) http.HandlerFunc {
	fn := func(res http.ResponseWriter, req *http.Request) {
		GetAllEncryptedData(res, req, stor)
	}
	return fn
}
