package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	JWTSecret    string

	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDatabase string

	RedisHost     string
	RedisPort     string
	RedisPassword string

	XrayAPIHost     string
	XrayAPIPort     string
	XrayInboundTag  string
}

func Load() (*Config, error) {
	// 优先尝试加载项目根 .env，找不到也不报错（生产用真实环境变量）
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	return &Config{
		AppEnv:    getenv("APP_ENV", "dev"),
		AppPort:   getenv("APP_PORT", "8080"),
		JWTSecret: getenv("APP_JWT_SECRET", "dev-secret-change-me"),

		PGHost:     getenv("POSTGRES_HOST", "127.0.0.1"),
		PGPort:     getenv("POSTGRES_PORT", "5432"),
		PGUser:     getenv("POSTGRES_USER", "routex"),
		PGPassword: getenv("POSTGRES_PASSWORD", "routex_dev_pwd"),
		PGDatabase: getenv("POSTGRES_DB", "routex"),

		RedisHost:     getenv("REDIS_HOST", "127.0.0.1"),
		RedisPort:     getenv("REDIS_PORT", "6379"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),

		XrayAPIHost:    getenv("XRAY_API_HOST", "127.0.0.1"),
		XrayAPIPort:    getenv("XRAY_API_PORT", "10085"),
		XrayInboundTag: getenv("XRAY_INBOUND_TAG", "vless-in"),
	}, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
