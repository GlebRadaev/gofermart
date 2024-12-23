package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetFlagsAndArgs() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

}

func resetEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")
	t.Setenv("DATABASE_URI", "")
	t.Setenv("LOG_LVL", "")
}

func setEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "localhost:9000")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "localhost:9001")
	t.Setenv("DATABASE_URI", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")
	t.Setenv("LOG_LVL", "debug")
}

func TestNew(t *testing.T) {
	setEnv(t)
	os.Args = []string{
		"cmd",
		"-a", "localhost:8080",
		"-r", "http://localhost:8082",
		"-d", "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
		"-l", "error",
	}
	cfg := New()

	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, "http://localhost:8082", cfg.AccrualAddress)
	assert.Equal(t, "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable", cfg.Database)
	assert.Equal(t, "error", cfg.LogLvl)
}

func TestAccrualAddressDefaultProtocol(t *testing.T) {
	resetFlagsAndArgs()
	resetEnv(t)
	setEnv(t)

	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "localhost:8083")

	cfg := New()

	assert.Equal(t, "http://localhost:8083", cfg.AccrualAddress)
	assert.Equal(t, "localhost:9000", cfg.Address)
}
