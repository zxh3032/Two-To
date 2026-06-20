package auth

import (
	"context"
	"strings"
	"time"

	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/proto"
)

// Captcha 生成图片验证码，把答案写入短期 cache，返回给前端验证码 ID 和图片。
func (s *Service) Captcha(ctx context.Context) (*proto.CaptchaResponse, error) {
	id, code, image, err := verification.NewCaptcha()
	if err != nil {
		return nil, internalError(err)
	}
	// 图片验证码只保存小写答案，校验时忽略大小写；TTL 到期后自动失效。
	key := s.cacheStore.Key("captcha", id)
	if err := s.cacheStore.Set(ctx, key, strings.ToLower(code), time.Duration(s.cfg.Verification.CaptchaTTLSeconds)*time.Second); err != nil {
		return nil, internalError(err)
	}
	return &proto.CaptchaResponse{
		CaptchaId:   id,
		ImageBase64: image,
		ExpiresIn:   s.cfg.Verification.CaptchaTTLSeconds,
	}, nil
}
