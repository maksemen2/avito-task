package logger

import (
	"os"
	"path/filepath"

	"github.com/maksemen2/avito-shop/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

	consoleEncoder := zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)

	cores := []zapcore.Core{}

	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, level)
	cores = append(cores, consoleCore)

	if loggerConfig.FilePath != "" {
		dir := filepath.Dir(loggerConfig.FilePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic("failed to create log directory: " + err.Error())
		}

		file, err := os.Create(loggerConfig.FilePath)
		if err != nil {
			panic("failed to create log file: " + err.Error())
		}

		fileEncoder := zapcore.NewJSONEncoder(cfg.EncoderConfig)
		fileWriter := zapcore.AddSync(file)
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, level)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}
