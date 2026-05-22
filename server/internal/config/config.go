package config

import (
	"os"
	"strconv"
	"time"

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

	XrayAPIHost    string
	XrayAPIPort    string
	XrayInboundTag string

	// PublicHost/PublicPort：客户端订阅链接里写的连接地址（兜底，节点未单独配置时使用）
	PublicHost string
	PublicPort int

	// Reality 客户端侧参数（与 deploy/xray/config.json 里的入站对应）
	RealityPublicKey string
	RealitySNI       string
	RealityShortID   string

	TrafficPollInterval    time.Duration
	ProviderHealthInterval time.Duration
}

func Load() (*Config, error) {
	// 优先尝试加载项目根 .env，找不到也不报错（生产用真实环境变量）
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	return &Config{
		AppEnv:    getenv("APP_ENV", "dev"),
		AppPort:   getenv("APP_PORT", "8891"),
		JWTSecret: getenv("APP_JWT_SECRET", "dev-secret-change-me"),

		PGHost:     getenv("POSTGRES_HOST", "127.0.0.1"),
		PGPort:     getenv("POSTGRES_PORT", "8892"),
		PGUser:     getenv("POSTGRES_USER", "routex"),
		PGPassword: getenv("POSTGRES_PASSWORD", "routex_dev_pwd"),
		PGDatabase: getenv("POSTGRES_DB", "routex"),

		RedisHost:     getenv("REDIS_HOST", "127.0.0.1"),
		RedisPort:     getenv("REDIS_PORT", "8893"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),

		XrayAPIHost:    getenv("XRAY_API_HOST", "127.0.0.1"),
		XrayAPIPort:    getenv("XRAY_API_PORT", "8894"),
		XrayInboundTag: getenv("XRAY_INBOUND_TAG", "vless-in"),

		PublicHost:       getenv("PUBLIC_HOST", "127.0.0.1"),
		PublicPort:       getenvInt("PUBLIC_PORT", 8895),
		RealityPublicKey: getenv("REALITY_PUBLIC_KEY", ""),
		RealitySNI:       getenv("REALITY_SNI", "www.cloudflare.com"),
		RealityShortID:   getenv("REALITY_SHORT_ID", ""),

		TrafficPollInterval:    time.Duration(getenvInt("TRAFFIC_POLL_SECONDS", 60)) * time.Second,
		ProviderHealthInterval: time.Duration(getenvInt("PROVIDER_HEALTH_SECONDS", 120)) * time.Second,
	}, nil
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
