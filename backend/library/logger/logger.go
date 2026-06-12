package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 创建 zap logger。开发环境使用便于阅读的输出，生产环境使用结构化 JSON 输出。
func New(level string, env string) (*zap.Logger, error) {
	var parsedLevel zapcore.Level
	if err := parsedLevel.Set(strings.ToLower(level)); err != nil {
		parsedLevel = zapcore.InfoLevel
	}

	cfg := zap.NewDevelopmentConfig()
	if env == "production" || env == "prod" {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = zap.NewAtomicLevelAt(parsedLevel)
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg.Build(zap.AddCaller())
}
