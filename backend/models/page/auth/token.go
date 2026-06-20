package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/library/servlet"
	"github.com/zxh3032/two-to/backend/library/verification"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/proto"
)

// createProfileSetupToken 生成资料完善短期 token，把已验证的联系方式和密码哈希放入 cache。
func (s *Service) createProfileSetupToken(ctx context.Context, payload verification.SetupPayload) (string, error) {
	token, err := security.RandomToken(32)
	if err != nil {
		return "", internalError(err)
	}
	key := s.cacheStore.Key("profile-setup", token)
	if err := verification.StoreJSON(ctx, s.cacheStore, key, payload, time.Duration(s.cfg.Verification.ProfileSetupTTLSeconds)*time.Second); err != nil {
		return "", internalError(err)
	}
	return token, nil
}

// issueLogin 创建设备会话、写入 session cache，并返回 access token 与 refresh token。
func (s *Service) issueLogin(ctx *servlet.Context, user *dao.User) (*proto.LoginResponse, error) {
	now := data.Now()
	refreshToken, refreshHash, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, internalError(err)
	}
	expireTime := now + int64(s.tokenManager.RefreshTTL().Seconds())
	// MySQL 作为会话事实源，用于账号安全页展示设备和审计会话状态变化。
	session := &dao.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		DeviceName:       deviceName(ctx.UserAgent),
		UserAgentHash:    security.HashPlain(ctx.UserAgent),
		IPHash:           security.ClientIPHash(ctx.IP),
		LastActiveTime:   now,
		ExpireTime:       expireTime,
		BaseModel:        dao.NewBaseModel(dao.SessionStatusNormal, now),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, internalError(err)
	}
	// Redis/cache 作为鉴权热路径，避免每个 access token 请求都回源 MySQL。
	if err := authlib.SetCachedSession(ctx, s.cacheStore, session.ID, authlib.CachedSession{UserID: user.ID, RefreshTokenHash: refreshHash, ExpireTime: expireTime}); err != nil {
		return nil, internalError(err)
	}
	accessToken, accessTTL, err := s.tokenManager.GenerateAccessToken(user.ID, session.ID)
	if err != nil {
		return nil, internalError(err)
	}
	return &proto.LoginResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresIn:  accessTTL,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresIn: int64(s.tokenManager.RefreshTTL().Seconds()),
		User:                  toUserInfo(user),
	}, nil
}

func fillEmailLoginResponse(resp *proto.EmailLoginResponse, auth *proto.LoginResponse) {
	resp.AccessToken = auth.GetAccessToken()
	resp.AccessTokenExpiresIn = auth.GetAccessTokenExpiresIn()
	resp.RefreshToken = auth.GetRefreshToken()
	resp.RefreshTokenExpiresIn = auth.GetRefreshTokenExpiresIn()
	resp.User = auth.GetUser()
}

func fillCompleteProfileResponse(resp *proto.CompleteProfileResponse, auth *proto.LoginResponse) {
	resp.AccessToken = auth.GetAccessToken()
	resp.AccessTokenExpiresIn = auth.GetAccessTokenExpiresIn()
	resp.RefreshToken = auth.GetRefreshToken()
	resp.RefreshTokenExpiresIn = auth.GetRefreshTokenExpiresIn()
	resp.User = auth.GetUser()
}

// deviceName 从 User-Agent 粗略提取浏览器和系统，用于账号安全页展示登录设备。
func deviceName(userAgent string) string {
	lower := strings.ToLower(userAgent)
	browser := "浏览器"
	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "chrome/"):
		browser = "Chrome"
	case strings.Contains(lower, "safari/"):
		browser = "Safari"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	}
	osName := "未知设备"
	switch {
	case strings.Contains(lower, "mac os"):
		osName = "macOS"
	case strings.Contains(lower, "windows"):
		osName = "Windows"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad"):
		osName = "iOS"
	case strings.Contains(lower, "android"):
		osName = "Android"
	}
	return fmt.Sprintf("%s · %s", browser, osName)
}
