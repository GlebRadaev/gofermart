package logger

import (
	"fmt"

	"github.com/GlebRadaev/gofermart/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const timeLayout = "15:04:05 02-01-2006"

var logLvlMap = map[string]zapcore.Level{
	"info":  zapcore.InfoLevel,
	"error": zapcore.ErrorLevel,
	"debug": zapcore.DebugLevel,
}

func InitLogger(conf *config.Config) error {
	lvl, ok := logLvlMap[conf.LogLvl]
	if !ok {
		return fmt.Errorf("unsupported log lvl: %s", conf.LogLvl)
	}

	encodeConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.TimeEncoderOfLayout(timeLayout),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
	}

	c := zap.Config{
		Level:            zap.NewAtomicLevelAt(lvl),
		Sampling:         nil,
		Encoding:         "console",
		EncoderConfig:    encodeConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := c.Build()
	if err != nil {
		return fmt.Errorf("unable to create zap logger, error: %w", err)
	}

	zap.ReplaceGlobals(logger)

	return nil
}
