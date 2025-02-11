package logger

import (
	"os"

	"go.uber.org/zap"
)

// Log будет доступен всему коду как синглтон.
// Никакой код, кроме функции InitLogger, не должен модифицировать эту переменную.
// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
var ClientLog *zap.Logger = zap.NewNop()

// Initialize - инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level, logFile string) error {

	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl

	// Определяю поток вывода логов
	// если установлен файл, то направляю вывод логов в файл
	if logFile != "" {
		// очищаю файл логов при старте
		err := os.Truncate(logFile, 0)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		// перенаправляю логи в файл
		cfg.OutputPaths = []string{logFile}
		cfg.ErrorOutputPaths = []string{logFile}
	}

	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	ClientLog = zl.With(zap.String("role", "agent"))
	return nil
}
