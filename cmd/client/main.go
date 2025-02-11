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
	"gophkeeper/internal/client/tui/add"
	"gophkeeper/internal/client/tui/add/bankcard"
	"gophkeeper/internal/client/tui/add/binary"
	"gophkeeper/internal/client/tui/add/password"
	"gophkeeper/internal/client/tui/add/text"
	"gophkeeper/internal/client/tui/app"
	"gophkeeper/internal/client/tui/data"
	"gophkeeper/internal/client/tui/home"
	"gophkeeper/internal/client/tui/ident/authorize"
	"gophkeeper/internal/client/tui/ident/register"
	"gophkeeper/internal/client/tui/view"
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
	registerPattern      = "/api/client/register"      // паттерн api для регистрации пользователя
	authorizationPattern = "/api/client/authorize"     // паттерн api для авторизации пользователя
	addDataPattern       = "/api/client/data/add"      // паттерн api для добавления новых данных на сервер
	replaceDataPattern   = "/api/client/data/replace"  // паттерн для замены старых данных на сервере новыми
	conflictDataPattern  = "/api/client/data/conflict" // паттерн для обработки данных с потенциальным конфликтом
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
	if err := logger.Initialize(logLevel, logFile); err != nil {
		log.Fatalf("Error starting client: %v", err)
	}
	// Добавляю многопоточность
	var wg sync.WaitGroup

	// Create a context with cancel function for graceful shutdown
	ctx, cancelCtx := context.WithCancel(ctx)

	// Создаю TUI интерфейс
	app := createTUI(ctx, stor, stor, info, client, decrData)

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
		client resty.Client, wg *sync.WaitGroup) {

		defer wg.Done()

		// Функция принимает копию resty клиента, чтобы установить на него мидлвари, необходимые только для синхронизации данных
		// Устанавливаю мидлвари для resty клиента
		client.OnBeforeRequest(auth.OnBeforeMiddleware(info, ident))
		client.OnAfterResponse(auth.OnAfterMiddleware(info, ident, netAddr+authorizationPattern))

		ticker := time.NewTicker(repoSynch.GetPeroidOfSynchr())
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done(): // Проверяю, был ли передан сигнал остановки
				logger.ClientLog.Info("Stopping data synchronization with server")
				return
			case <-ticker.C:
				logger.ClientLog.Info("Start data synchronization with server")

				err := synchronization.SynchronizeData(ctx, stor, info, &client, netAddr+addDataPattern, netAddr+conflictDataPattern)
				if err != nil {
					logger.ClientLog.Error("failed to synchronize data", zap.String("server address", netAddr))
				}
			}
		}

	}(ctx, stor, info, stor, *client, &wg)

	// Запускаю фоновое обновление расшифрованных данных пользователя во временном хранилище
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(inmemory.GetUpdatingPeriod())
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
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

// createTUI - функция для создания интерфейса tui.
func createTUI(ctx context.Context, stor storage.IEncryptedClientStorage, ident identity.ClientIdentifier, info identity.IUserInfoStorage,
	client *resty.Client, decrData storage.IStorage) *app.App {

	// Копирую resty клиента и устанавливаю авторизационную middleware для запросов, которые требуют,
	// чтобы пользователь был авторизирован
	authClient := *client
	authClient.OnBeforeRequest(auth.OnBeforeMiddleware(info, ident))

	// создаю страницы TUI
	prims := []app.Primitives{}

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
	// Добавляю страницу для взаимодействия с данными
	prims = append(prims, app.Primitives{
		Name: tui.Data,
		Prim: data.DataPage,
	})
	// Добавляю страницу для визуализации данных
	prims = append(prims, app.Primitives{
		Name: tui.View,
		Prim: view.ViewPage(ctx, decrData),
	})
	// Добавляю страницу для добавления новых данных
	prims = append(prims, app.Primitives{
		Name: tui.Add,
		Prim: add.AddData,
	})
	// Добавляю страницу для добавления новой банковской карты
	prims = append(prims, app.Primitives{
		Name: tui.AddBankCard,
		Prim: bankcard.AddBankcardPage(ctx, netAddr+addDataPattern, &authClient, stor, info),
	})
	// Добавляю страницу для добавления новых бинарных данных
	prims = append(prims, app.Primitives{
		Name: tui.AddBinary,
		Prim: binary.AddBinaryPage(ctx, netAddr+addDataPattern, &authClient, stor, info),
	})
	// Добавляю страницу для добавления нового пароля
	prims = append(prims, app.Primitives{
		Name: tui.AddPassword,
		Prim: password.AddPasswordPage(ctx, netAddr+addDataPattern, &authClient, stor, info),
	})
	// Добавляю страницу для добавления новых текстовых данных
	prims = append(prims, app.Primitives{
		Name: tui.AdddText,
		Prim: text.AddTextPage(ctx, netAddr+addDataPattern, &authClient, stor, info),
	})
	// Добавляю приветственную страницу
	prims = append(prims, app.Primitives{
		Name: tui.Home,
		Prim: home.HomePage,
	})

	return app.NewApp(prims)
}
