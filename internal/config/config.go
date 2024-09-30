package config

import (
	"flag"
	"os"
)

func InitParams() Config {
	a := flag.String("a", "localhost:8080", "server host")
	r := flag.String("b", "https://zod9d.wiremockapi.cloud", "accrual system address")
	d := flag.String("d", "postgresql://postgres:password@localhost", "db connect string")

	flag.Parse()
	conf := NewConfig()
	if conf.Cfg.ServerAddress == "" {
		conf.Cfg.ServerAddress = *a
	}
	if conf.Cfg.AccrualAddress == "" {
		conf.Cfg.AccrualAddress = *r
	}

	if conf.Cfg.DataBaseDNS == "" {
		conf.Cfg.DataBaseDNS = *d
	}
	return *conf
}

type ServerConfig struct {
	ServerAddress  string
	AccrualAddress string
	DataBaseDNS    string
}

type Config struct {
	Cfg ServerConfig
}

func NewConfig() *Config {
	return &Config{
		Cfg: ServerConfig{
			ServerAddress:  getEnv("RUN_ADDRESS", ""),
			AccrualAddress: getEnv("ACCRUAL_SYSTEM_ADDRESS", ""),
			DataBaseDNS:    getEnv("DATABASE_URI", ""),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
