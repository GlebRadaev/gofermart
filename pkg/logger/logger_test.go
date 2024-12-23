package logger

import (
	"testing"

	"github.com/GlebRadaev/gofermart/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		expectedError  bool
		expectedLogLvl zapcore.Level
	}{
		{
			name: "Valid log level info",
			config: &config.Config{
				LogLvl: "info",
			},
			expectedError:  false,
			expectedLogLvl: zapcore.InfoLevel,
		},
		{
			name: "Valid log level error",
			config: &config.Config{
				LogLvl: "error",
			},
			expectedError:  false,
			expectedLogLvl: zapcore.ErrorLevel,
		},
		{
			name: "Valid log level debug",
			config: &config.Config{
				LogLvl: "debug",
			},
			expectedError:  false,
			expectedLogLvl: zapcore.DebugLevel,
		},
		{
			name: "Invalid log level",
			config: &config.Config{
				LogLvl: "invalid",
			},
			expectedError:  true,
			expectedLogLvl: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.config)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
