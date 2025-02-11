package main

import (
	"context"
	"database/sql"
	"gophkeeper/internal/client/identity"
	"gophkeeper/internal/client/identity/auth"
	"gophkeeper/internal/client/logger"
	"gophkeeper/internal/client/storage"
	"gophkeeper/internal/client/storage/info"
	"gophkeeper/internal/client/storage/inmemory"
	"gophkeeper/internal/client/storage/pg"
	"gophkeeper/internal/client/synchronization"
	"gophkeeper/internal/client/tui"
	"gophkeeper/internal/client/tui/app"
	"gophkeeper/internal/client/tui/home"
	"gophkeeper/internal/client/tui/ident/authorize"
	"gophkeeper/internal/client/tui/ident/register"
	repoSynch "gophkeeper/internal/repositories/synchronization"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	registerPattern      = "/api/client/register"  // паттерн api для регистрации пользователя
	authorizationPattern = "/api/client/authorize" // паттерн api для авторизации пользователя
	addDataPattern       = "/api/client/add"       // паттерн api для добавления новых данных на сервер
	replaceDataPattern   = "/api/client/replace"   // паттерн для замены старых данных на сервере новыми
	conflictDataPattern  = "/api/client/conflict"  // паттерн для обработки данных с потенциальным конфликтом
)

func main() {
	err := parseVariables()
	if err != nil {
		log.Fatalf("failed to set global variables, %v", err)
	}

	// Подключение к базе данных
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		log.Fatalf("Error connection to database: %v by address %s", err, databaseDsn)
	}
	defer db.Close()

	// Проверка соединения с БД
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Error checking connection with database: %v\n", err)
	}
	// создаем экземпляр хранилища pg
	stor := pg.NewStore(db)
	err = stor.Bootstrap(ctx)
	if err != nil {
		log.Fatalf("Error prepare database to work: %v\n", err)
	}

	// Инициализирую хранилище данных пользователя в оперативной памяти
	info := info.NewUserInfoStorage()

	// Инициализирую хранилище расшифрованных данных пользователя в оперативной памяти
	decrData := inmemory.NewDecryptedData()

	// Инициализирую resty клиента
	client := resty.New()

	// ------------------------------------------------------------------------------
	run(ctx, stor, info, client, decrData)
}

// run - будет полезна при инициализации зависимостей клиента перед запуском
func run(ctx context.Context, stor *pg.Store, info identity.IUserInfoStorage, client *resty.Client, decrData storage.IStorage) {
	// инициализация логера
	if err := logger.Initialize(logLevel); err != nil {
		log.Fatalf("Error starting client: %v", err)
	}
	// Добавляю многопоточность
	var wg sync.WaitGroup

	// Create a context with cancel function for graceful shutdown
	ctx, cancelCtx := context.WithCancel(ctx)

	// Создаю TUI интерфейс
	app := createTUI(ctx, stor, info, client)

	// Запускаю интерфейс в отдельной горутине
	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("tui stopped with error, %v", err)
		}
	}()

	// Горутина для остановки TUI при завершении контекста
	wg.Add(1)
	go func() {
		defer wg.Done()

		// ожидаю завершения контекста
		<-ctx.Done()

		// Завершаю работу интерфейса
		app.Stop()
	}()

	// Запускаю фоновую синхронизацию данных с сервером
	wg.Add(1)
	go func(ctx context.Context, stor storage.IEncryptedClientStorage, info identity.IUserInfoStorage, ident identity.ClientIdentifier,
		client *resty.Client, wg *sync.WaitGroup) {

		// Устанавливаю мидлвари для resty клиента
		client.OnBeforeRequest(auth.OnBeforeMiddleware(info, ident))
		client.OnAfterResponse(auth.OnAfterMiddleware(info, ident, netAddr+authorizationPattern))

		ticker := time.NewTicker(repoSynch.GetPeroidOfSynchr())
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done(): // Проверяю, был ли передан сигнал остановки
				wg.Done()
				logger.ClientLog.Info("Stopping synchronization")
				return
			case <-ticker.C:
				logger.ClientLog.Info("Start synchronization")

				err := synchronization.SynchronizeData(ctx, stor, info, client, netAddr+addDataPattern, netAddr+conflictDataPattern)
				if err != nil {
					logger.ClientLog.Error("failed to synchronize data", zap.String("server address", netAddr))
				}
			}
		}

	}(ctx, stor, info, stor, client, &wg)

	// Запускаю фоновое обновление расшифрованных данных пользователя во временном хранилище
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(inmemory.GetUpdatingPeriod())
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				wg.Done()
				logger.ClientLog.Info("Stopping decrypted data updating")
				return
			case <-ticker.C:
				logger.ClientLog.Info("Start decrypted data updating")
				err := decrData.Update(ctx, stor, info)
				if err != nil {
					logger.ClientLog.Error("failed to update decrypted data", zap.String("error", err.Error()))
				}
			}
		}
	}()

	// Канал для получения сигнала прерывания
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Блокирование до тех пор, пока не поступит сигнал о прерывании
	<-quit
	logger.ClientLog.Info("Shutting down client...")

	// Закрываю контекст, для остановки функции записи данных в канал для отправки на сервер
	cancelCtx()

	// Ожидаю завершения работы всех горутин
	wg.Wait()

	logger.ClientLog.Info("Shutdown the client gracefully")
}

func createTUI(ctx context.Context, ident identity.ClientIdentifier, info identity.IUserInfoStorage, client *resty.Client) *app.App {
	// создаю страницы TUI
	prims := []app.Primitives{}
	// Добавляю приветственную страницу
	prims = append(prims, app.Primitives{
		Name: tui.Home,
		Prim: home.HomePage,
	})
	// Добавляю страницу регистрации
	prims = append(prims, app.Primitives{
		Name: tui.Register,
		Prim: register.RegisterPage(ctx, ident, netAddr+registerPattern, client),
	})
	// Добавляю страницу авторизации
	prims = append(prims, app.Primitives{
		Name: tui.Login,
		Prim: authorize.LoginPage(ctx, ident, info),
	})

	return app.NewApp(prims)
}
