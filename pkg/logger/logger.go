package logger

import (
	"log"
	"os"
	"path/filepath"

	"github.com/maksemen2/avito-shop/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newZapCore(encoderType string, writer zapcore.WriteSyncer, cfg zapcore.EncoderConfig, level zapcore.Level) zapcore.Core {
	var encoder zapcore.Encoder
	if encoderType == "json" {
		encoder = zapcore.NewJSONEncoder(cfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(cfg)
	}

	return zapcore.NewCore(encoder, writer, level)
}

// MustLoad инициализирует логгер с указанным в конфиге уровнем логирования.
// Если указан LoggerConfig.FilePath - логи будут дублироваться в файл по указанному пути.
// Создает необходимые директории для файла логов.
// В случае ошибки завершает работу программы.
// Время логирования форматируется в ISO8601.
func MustLoad(loggerConfig config.LoggerConfig) *zap.Logger {
	var level zapcore.Level
	if err := level.Set(loggerConfig.Level); err != nil {
		level = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	cores := []zapcore.Core{}

	consoleCore := newZapCore("console", zapcore.AddSync(os.Stdout), cfg.EncoderConfig, level)
	cores = append(cores, consoleCore)

	if loggerConfig.FilePath != "" {
		dir := filepath.Dir(loggerConfig.FilePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("failed to create log directory: %s", err.Error())
		}

		file, err := os.Create(loggerConfig.FilePath)
		if err != nil {
			log.Fatalf("failed to create log file: %s", err.Error())
		}

		fileCore := newZapCore("json", zapcore.AddSync(file), cfg.EncoderConfig, level)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}
