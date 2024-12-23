package config

import (
	"flag"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string `env:"RUN_ADDRESS"            envDefault:"localhost:8080"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"localhost:8081"`
	Database       string `env:"DATABASE_URI"           envDefault:"postgres://gofermart:gofermart@localhost:54321/gofermart?sslmode=disable"`
	LogLvl         string `env:"LOG_LVL"                envDefault:"info"`
}

func New() *Config {
	cfg := &Config{}

	env.Parse(cfg)

	flag.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run server")
	flag.StringVar(&cfg.AccrualAddress, "r", cfg.AccrualAddress, "accrual system address and port")
	flag.StringVar(&cfg.Database, "d", cfg.Database, "database DSN")
	flag.StringVar(&cfg.LogLvl, "l", cfg.LogLvl, "log level")
	flag.Parse()

	if !strings.HasPrefix(cfg.AccrualAddress, "http://") && !strings.HasPrefix(cfg.AccrualAddress, "https://") {
		cfg.AccrualAddress = "http://" + cfg.AccrualAddress
	}

	return cfg
}
