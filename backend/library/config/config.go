package config

import (
	"os"
	"strings"
)

// Config 保存服务启动所需的基础配置。
type Config struct {
	AppName  string
	Env      string
	HTTPAddr string
	LogLevel string
	MySQLDSN string
}

// Load 从环境变量读取配置，并为本地开发提供可直接启动的默认值。
func Load() Config {
	return Config{
		AppName:  envOrDefault("TWO_TO_APP_NAME", "two-to-api"),
		Env:      strings.ToLower(envOrDefault("TWO_TO_ENV", "development")),
		HTTPAddr: envOrDefault("TWO_TO_HTTP_ADDR", ":8080"),
		LogLevel: envOrDefault("TWO_TO_LOG_LEVEL", "debug"),
		MySQLDSN: os.Getenv("TWO_TO_MYSQL_DSN"),
	}
}

// IsProduction 判断当前是否为生产环境，用于切换日志和 Gin 运行模式。
func (c Config) IsProduction() bool {
	return c.Env == "production" || c.Env == "prod"
}

// envOrDefault 集中处理环境变量默认值，避免启动配置分散在 main 中。
func envOrDefault(key string, defaultValue string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return defaultValue
}
