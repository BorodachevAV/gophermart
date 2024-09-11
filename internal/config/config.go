package config

import (
	"os"
)

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
