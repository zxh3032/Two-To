package sms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	aliyun "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	volcsms "github.com/volcengine/volc-sdk-golang/service/sms"
	"github.com/zxh3032/two-to/backend/library/config"
	"github.com/zxh3032/two-to/backend/library/security"
	"go.uber.org/zap"
)

// Sender 负责按配置主备发送短信验证码。
type Sender interface {
	Name() string
	SendCode(ctx context.Context, phone string, code string, scene string) (provider string, err error)
}

// provider 是具体短信供应商的最小接口，便于主备 provider 编排。
type provider interface {
	name() string
	send(ctx context.Context, phone string, code string, scene string) error
}

// NewSender 根据配置组装短信主备供应商，未配置可用供应商时回退 mock。
func NewSender(cfg config.Config, log *zap.Logger) Sender {
	providers := make([]provider, 0, 2)
	for _, name := range []string{cfg.SMS.PrimaryProvider, cfg.SMS.BackupProvider} {
		if name == "" {
			continue
		}
		if p := newProvider(name, cfg, log); p != nil {
			providers = append(providers, p)
		}
	}
	if len(providers) == 0 {
		providers = append(providers, &mockProvider{log: log})
	}
	return &multiSender{providers: providers, log: log}
}

// multiSender 按顺序尝试主备短信 provider，任一发送成功即返回。
type multiSender struct {
	providers []provider
	log       *zap.Logger
}

// Name 返回当前启用的短信 provider 列表，便于启动后确认配置。
func (s *multiSender) Name() string {
	names := make([]string, 0, len(s.providers))
	for _, p := range s.providers {
		names = append(names, p.name())
	}
	return strings.Join(names, ",")
}

// SendCode 发送短信验证码，主 provider 失败时自动尝试备用 provider。
func (s *multiSender) SendCode(ctx context.Context, phone string, code string, scene string) (string, error) {
	var lastErr error
	for _, p := range s.providers {
		if err := p.send(ctx, phone, code, scene); err != nil {
			lastErr = err
			s.log.Error("短信验证码发送失败，准备尝试下一个 provider",
				zap.String("provider", p.name()),
				zap.String("phone", security.MaskPhone(phone)),
				zap.Error(err),
			)
			continue
		}
		return p.name(), nil
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", errors.New("没有可用短信 provider")
}

// newProvider 按配置名称创建短信供应商，配置不完整时跳过该供应商。
func newProvider(name string, cfg config.Config, log *zap.Logger) provider {
	switch strings.ToLower(name) {
	case "mock":
		return &mockProvider{log: log}
	case "aliyun":
		aliyunCfg := cfg.SMS.Aliyun
		if aliyunCfg.AccessKeyID == "" || aliyunCfg.AccessKeySecret == "" || aliyunCfg.SignName == "" || aliyunCfg.TemplateCode == "" {
			log.Warn("阿里云短信配置不完整，已跳过 aliyun provider")
			return nil
		}
		return &aliyunProvider{cfg: aliyunCfg}
	case "volcengine":
		volcCfg := cfg.SMS.Volcengine
		if volcCfg.AccessKeyID == "" || volcCfg.AccessKeySecret == "" || volcCfg.SignName == "" || volcCfg.TemplateID == "" || volcCfg.SMSAccount == "" {
			log.Warn("火山云短信配置不完整，已跳过 volcengine provider")
			return nil
		}
		instance := volcsms.NewInstance()
		instance.SetRegion(volcCfg.RegionID)
		instance.Client.SetAccessKey(volcCfg.AccessKeyID)
		instance.Client.SetSecretKey(volcCfg.AccessKeySecret)
		return &volcengineProvider{cfg: volcCfg, instance: instance}
	default:
		log.Warn("未知短信 provider，已跳过", zap.String("provider", name))
		return nil
	}
}

// mockProvider 用于本地开发，验证码只写日志不发送真实短信。
type mockProvider struct {
	log *zap.Logger
}

// name 返回 mock provider 名称。
func (p *mockProvider) name() string {
	return "mock"
}

// send 记录 mock 短信验证码，手机号写日志前会脱敏。
func (p *mockProvider) send(_ context.Context, phone string, code string, scene string) error {
	p.log.Warn("使用 mock 短信验证码发送", zap.String("phone", security.MaskPhone(phone)), zap.String("scene", scene), zap.String("code", code))
	return nil
}

// aliyunProvider 使用阿里云短信服务发送验证码。
type aliyunProvider struct {
	cfg config.AliyunSMSConfig
}

// name 返回阿里云短信 provider 名称。
func (p *aliyunProvider) name() string {
	return "aliyun"
}

// send 调用阿里云 SendSms 接口，当前模板参数统一传入 code。
func (p *aliyunProvider) send(ctx context.Context, phone string, code string, _ string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	client, err := aliyun.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(p.cfg.AccessKeyID),
		AccessKeySecret: tea.String(p.cfg.AccessKeySecret),
		RegionId:        tea.String(p.cfg.RegionID),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	})
	if err != nil {
		return err
	}
	templateParam, _ := json.Marshal(map[string]string{"code": code})
	resp, err := client.SendSmsWithOptions(&aliyun.SendSmsRequest{
		PhoneNumbers:  tea.String(strings.TrimPrefix(phone, "+86")),
		SignName:      tea.String(p.cfg.SignName),
		TemplateCode:  tea.String(p.cfg.TemplateCode),
		TemplateParam: tea.String(string(templateParam)),
	}, &util.RuntimeOptions{})
	if err != nil {
		return err
	}
	if resp == nil || resp.Body == nil || tea.StringValue(resp.Body.Code) != "OK" {
		codeText := ""
		message := ""
		if resp != nil && resp.Body != nil {
			codeText = tea.StringValue(resp.Body.Code)
			message = tea.StringValue(resp.Body.Message)
		}
		return fmt.Errorf("aliyun sms failed: %s %s", codeText, message)
	}
	return nil
}

// volcengineProvider 使用火山云短信服务发送验证码。
type volcengineProvider struct {
	cfg      config.VolcengineSMSConfig
	instance *volcsms.SMS
}

// name 返回火山云短信 provider 名称。
func (p *volcengineProvider) name() string {
	return "volcengine"
}

// send 调用火山云短信接口，当前模板参数统一传入 code。
func (p *volcengineProvider) send(ctx context.Context, phone string, code string, _ string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	templateParam, _ := json.Marshal(map[string]string{"code": code})
	resp, statusCode, err := p.instance.Send(&volcsms.SmsRequest{
		SmsAccount:    p.cfg.SMSAccount,
		Sign:          p.cfg.SignName,
		TemplateID:    p.cfg.TemplateID,
		TemplateParam: string(templateParam),
		PhoneNumbers:  strings.TrimPrefix(phone, "+86"),
	})
	if err != nil {
		return err
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("volcengine sms failed with http status %d", statusCode)
	}
	if resp != nil && resp.ResponseMetadata.Error != nil {
		return fmt.Errorf("volcengine sms failed: %s %s", resp.ResponseMetadata.Error.Code, resp.ResponseMetadata.Error.Message)
	}
	return nil
}
