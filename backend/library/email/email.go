package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/security"
	"go.uber.org/zap"
)

// Sender 负责发送邮箱验证码。
type Sender interface {
	Name() string
	SendCode(ctx context.Context, to string, code string, scene string) error
}

// NewSender 根据 SMTP 配置创建真实邮件发送器；未配置时使用 mock 方便本地开发。
func NewSender(cfg config.Config, log *zap.Logger) Sender {
	if cfg.SMTP.Host == "" {
		return &MockSender{log: log}
	}
	return &SMTPSender{cfg: cfg.SMTP}
}

// SMTPSender 使用标准 SMTP 协议发送邮箱验证码。
type SMTPSender struct {
	cfg config.SMTPConfig
}

// Name 返回邮件发送 provider 名称，写入验证码审计记录。
func (s *SMTPSender) Name() string {
	return "smtp"
}

// SendCode 发送邮箱验证码，邮件正文只包含验证码和有效期提示。
func (s *SMTPSender) SendCode(ctx context.Context, to string, code string, scene string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if s.cfg.Host == "" || s.cfg.From == "" {
		return errors.New("SMTP 配置不完整")
	}
	subject := "Two-To 邮箱验证码"
	body := fmt.Sprintf("你的 Two-To 验证码是 %s，10 分钟内有效。若不是你本人操作，请忽略本邮件。", code)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", s.cfg.FromName, s.cfg.From),
		fmt.Sprintf("To: %s", to),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(message))
}

// MockSender 在本地开发时把验证码写入日志，不真正发送邮件。
type MockSender struct {
	log *zap.Logger
}

// Name 返回 mock 邮件 provider 名称。
func (s *MockSender) Name() string {
	return "mock-email"
}

// SendCode 记录 mock 邮箱验证码，日志中的收件人会先脱敏。
func (s *MockSender) SendCode(_ context.Context, to string, code string, scene string) error {
	s.log.Warn("使用 mock 邮箱验证码发送", zap.String("to", security.MaskEmail(to)), zap.String("scene", scene), zap.String("code", code))
	return nil
}
