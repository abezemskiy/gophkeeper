package main

import (
	"context"
	"gophkeeper/internal/server/handlers"
	"gophkeeper/internal/server/identity/auth"
	"gophkeeper/internal/server/logger"
	"gophkeeper/internal/server/storage/pg"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const shutdownWaitPeriod = 20 * time.Second // для установки в контекст для реализаации graceful shutdown

func main() {
	err := parseVariables()
	if err != nil {
		log.Fatalf("failed to set global variables, %v", err)
	}

	ctx := context.Background()
	// создаем экземпляр хранилища pg
	stor, err := pg.NewStore(ctx, netAddr)
	if err != nil {
		log.Fatalf("Failed to create storage: %v\n", err)
	}
	// ------------------------------------------------------------------------------

	run(ctx, stor)
}

// функция run будет необходима для инициализации зависимостей сервера перед запуском
func run(ctx context.Context, stor *pg.Store) {
	// Инициализация логера
	if err := logger.Initialize(logLevel); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	logger.ServerLog.Info("Running gophkeeper", zap.String("address", netAddr))

	// запускаю сам сервис с проверкой отмены контекста для реализации graceful shutdown--------------
	srv := &http.Server{
		Addr:    netAddr,
		Handler: MetricRouter(stor),
	}
	// Канал для получения сигнала прерывания
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Горутина для запуска сервера
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Блокирование до тех пор, пока не поступит сигнал о прерывании
	<-quit
	logger.ServerLog.Info("Shutting down server...", zap.String("address", netAddr))

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(ctx, shutdownWaitPeriod)
	defer cancel()

	// останавливаю сервер, чтобы он перестал принимать новые запросы
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Stopping server error: %v", err)
	}

	logger.ServerLog.Info("Shutdown the server gracefully", zap.String("address", netAddr))
}

// MetricRouter - дирежирует обработку http запросов к серверу.
func MetricRouter(stor *pg.Store) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/client", func(r chi.Router) {
		r.Post("/register", logger.RequestLogger(handlers.RegisterHandler(stor)))
		r.Post("/authorize", logger.RequestLogger(handlers.AuthorizeHandler(stor)))

		r.Route("/data", func(r chi.Router) {
			r.Post("/add", logger.RequestLogger(auth.Middleware(handlers.AddEncryptedDataHandler(stor))))
			r.Post("/replace", logger.RequestLogger(auth.Middleware(handlers.ReplaceEncryptedDataHandler(stor))))
			r.Get("/get", logger.RequestLogger(auth.Middleware(handlers.GetAllEncryptedDataHandler(stor))))
			r.Delete("/delete", logger.RequestLogger(auth.Middleware(handlers.DeleteEncryptedDataHandler(stor))))
			r.Post("/conflict", logger.RequestLogger(auth.Middleware(handlers.HandleConflictDataHandler(stor))))
		})
	})

	// Определяем маршрут по умолчанию для некорректных запросов
	r.NotFound(logger.RequestLogger(handlers.HandleOtherRequest()))

	return r
}
