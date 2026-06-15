package config

import (
	"os"
	"strconv"
	"strings"
)

// Config 保存服务启动所需的基础配置。
type Config struct {
	AppName      string
	Env          string
	HTTPAddr     string
	LogLevel     string
	MySQLDSN     string
	Token        TokenConfig
	Verification VerificationConfig
	Cache        CacheConfig
	SMTP         SMTPConfig
	SMS          SMSConfig
}

// TokenConfig 保存 access token 和 refresh token 的签发配置。
type TokenConfig struct {
	AccessSecret      string
	AccessTTLSeconds  int64
	RefreshSecret     string
	RefreshTTLSeconds int64
}

// VerificationConfig 保存验证码和短期流程 token 的时效配置。
type VerificationConfig struct {
	CaptchaTTLSeconds         int64
	CodeTTLSeconds            int64
	CodeResendCooldownSeconds int64
	ProfileSetupTTLSeconds    int64
	PasswordResetTTLSeconds   int64
	CodeTargetHourlyLimit     int64
	CodeTargetDailyLimit      int64
	CodeIPHourlyLimit         int64
	LoginCaptchaFailThreshold int64
}

// CacheConfig 保存 Redis 与本地内存 cache 的运行配置。
type CacheConfig struct {
	Driver    string
	RedisAddr string
	Username  string
	Password  string
	DB        int
	KeyPrefix string
}

// SMTPConfig 保存邮件验证码发送配置。
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// SMSConfig 保存短信验证码主备供应商配置。
type SMSConfig struct {
	PrimaryProvider string
	BackupProvider  string
	Aliyun          AliyunSMSConfig
	Volcengine      VolcengineSMSConfig
}

// AliyunSMSConfig 保存阿里云短信配置。
type AliyunSMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	RegionID        string
}

// VolcengineSMSConfig 保存火山云短信配置。
type VolcengineSMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateID      string
	SMSAccount      string
	RegionID        string
}

// Load 从环境变量读取配置，并为本地开发提供可直接启动的默认值。
func Load() Config {
	env := strings.ToLower(envOrDefault("TWO_TO_ENV", "development"))
	return Config{
		AppName:  envOrDefault("TWO_TO_APP_NAME", "two-to-api"),
		Env:      env,
		HTTPAddr: envOrDefault("TWO_TO_HTTP_ADDR", ":0806"),
		LogLevel: envOrDefault("TWO_TO_LOG_LEVEL", "debug"),
		MySQLDSN: os.Getenv("TWO_TO_MYSQL_DSN"),
		Token: TokenConfig{
			AccessSecret:      envOrDefault("TWO_TO_ACCESS_TOKEN_SECRET", "two-to-dev-access-secret"),
			AccessTTLSeconds:  int64OrDefault("TWO_TO_ACCESS_TOKEN_TTL_SECONDS", 900),
			RefreshSecret:     envOrDefault("TWO_TO_REFRESH_TOKEN_SECRET", "two-to-dev-refresh-secret"),
			RefreshTTLSeconds: int64OrDefault("TWO_TO_REFRESH_TOKEN_TTL_SECONDS", 2592000),
		},
		Verification: VerificationConfig{
			CaptchaTTLSeconds:         int64OrDefault("TWO_TO_CAPTCHA_TTL_SECONDS", 300),
			CodeTTLSeconds:            int64OrDefault("TWO_TO_VERIFICATION_CODE_TTL_SECONDS", 600),
			CodeResendCooldownSeconds: int64OrDefault("TWO_TO_CODE_RESEND_COOLDOWN_SECONDS", 60),
			ProfileSetupTTLSeconds:    int64OrDefault("TWO_TO_PROFILE_SETUP_TTL_SECONDS", 600),
			PasswordResetTTLSeconds:   int64OrDefault("TWO_TO_PASSWORD_RESET_TTL_SECONDS", 600),
			CodeTargetHourlyLimit:     int64OrDefault("TWO_TO_CODE_TARGET_HOURLY_LIMIT", 5),
			CodeTargetDailyLimit:      int64OrDefault("TWO_TO_CODE_TARGET_DAILY_LIMIT", 10),
			CodeIPHourlyLimit:         int64OrDefault("TWO_TO_CODE_IP_HOURLY_LIMIT", 30),
			LoginCaptchaFailThreshold: int64OrDefault("TWO_TO_LOGIN_CAPTCHA_FAIL_THRESHOLD", 3),
		},
		Cache: CacheConfig{
			Driver:    strings.ToLower(envOrDefault("TWO_TO_CACHE_DRIVER", defaultCacheDriver(env))),
			RedisAddr: envOrDefault("TWO_TO_REDIS_ADDR", "127.0.0.1:6379"),
			Username:  os.Getenv("TWO_TO_REDIS_USERNAME"),
			Password:  os.Getenv("TWO_TO_REDIS_PASSWORD"),
			DB:        intOrDefault("TWO_TO_REDIS_DB", 0),
			KeyPrefix: strings.Trim(envOrDefault("TWO_TO_REDIS_KEY_PREFIX", "two-to"), ":"),
		},
		SMTP: SMTPConfig{
			Host:     os.Getenv("TWO_TO_SMTP_HOST"),
			Port:     intOrDefault("TWO_TO_SMTP_PORT", 587),
			Username: os.Getenv("TWO_TO_SMTP_USERNAME"),
			Password: os.Getenv("TWO_TO_SMTP_PASSWORD"),
			From:     os.Getenv("TWO_TO_SMTP_FROM"),
			FromName: envOrDefault("TWO_TO_SMTP_FROM_NAME", "Two-To"),
		},
		SMS: SMSConfig{
			PrimaryProvider: strings.ToLower(envOrDefault("TWO_TO_SMS_PRIMARY_PROVIDER", "mock")),
			BackupProvider:  strings.ToLower(os.Getenv("TWO_TO_SMS_BACKUP_PROVIDER")),
			Aliyun: AliyunSMSConfig{
				AccessKeyID:     os.Getenv("TWO_TO_ALIYUN_SMS_ACCESS_KEY_ID"),
				AccessKeySecret: os.Getenv("TWO_TO_ALIYUN_SMS_ACCESS_KEY_SECRET"),
				SignName:        os.Getenv("TWO_TO_ALIYUN_SMS_SIGN_NAME"),
				TemplateCode:    os.Getenv("TWO_TO_ALIYUN_SMS_TEMPLATE_CODE"),
				RegionID:        envOrDefault("TWO_TO_ALIYUN_SMS_REGION_ID", "cn-hangzhou"),
			},
			Volcengine: VolcengineSMSConfig{
				AccessKeyID:     os.Getenv("TWO_TO_VOLCENGINE_SMS_ACCESS_KEY_ID"),
				AccessKeySecret: os.Getenv("TWO_TO_VOLCENGINE_SMS_ACCESS_KEY_SECRET"),
				SignName:        os.Getenv("TWO_TO_VOLCENGINE_SMS_SIGN_NAME"),
				TemplateID:      os.Getenv("TWO_TO_VOLCENGINE_SMS_TEMPLATE_ID"),
				SMSAccount:      os.Getenv("TWO_TO_VOLCENGINE_SMS_SMS_ACCOUNT"),
				RegionID:        envOrDefault("TWO_TO_VOLCENGINE_SMS_REGION_ID", "cn-north-1"),
			},
		},
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

// defaultCacheDriver 根据环境选择默认 cache，生产默认 Redis，本地默认内存。
func defaultCacheDriver(env string) string {
	if env == "production" || env == "prod" {
		return "redis"
	}
	return "memory"
}

// intOrDefault 读取整型环境变量，解析失败时使用默认值保证服务可启动。
func intOrDefault(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// int64OrDefault 读取 int64 环境变量，适用于各类秒级 TTL 配置。
func int64OrDefault(key string, defaultValue int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}
