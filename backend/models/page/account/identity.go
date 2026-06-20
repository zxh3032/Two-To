package account

import (
	"context"
	"errors"

	"github.com/zxh3032/two-to/backend/library/apperror"
	authlib "github.com/zxh3032/two-to/backend/library/auth"
	"github.com/zxh3032/two-to/backend/library/response"
	"github.com/zxh3032/two-to/backend/library/security"
	"github.com/zxh3032/two-to/backend/models/dao"
	"github.com/zxh3032/two-to/backend/models/data"
	"github.com/zxh3032/two-to/backend/models/page/pagectx"
	"gorm.io/gorm"
)

// updateIdentity 绑定或换绑联系方式；同类型旧联系方式会先解绑，再写入新联系方式。
func (s *Service) updateIdentity(ctx context.Context, userID uint64, identityType int32, identityValue string, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	now := data.Now()
	err := s.store.WithTx(ctx, func(tx *data.Store) error {
		// 一个邮箱或手机号只能绑定到一个账号，避免登录入口产生歧义。
		existing, err := tx.FindIdentity(ctx, identityType, identityValue)
		if err != nil {
			return err
		}
		if existing != nil && existing.UserID != userID {
			return gorm.ErrDuplicatedKey
		}
		// 先解绑同类型旧身份，再 upsert 新身份，保证账号侧同类型联系方式只有一个有效值。
		if err := tx.UnbindIdentity(ctx, userID, identityType, now); err != nil {
			return err
		}
		return tx.UpsertIdentity(ctx, &dao.UserAuthIdentity{
			UserID:        userID,
			IdentityType:  identityType,
			IdentityValue: data.StringPtr(identityValue),
			VerifyTime:    now,
			BaseModel:     dao.NewBaseModel(dao.IdentityStatusNormal, now),
		})
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperror.Conflict(response.CodeConflict, "该联系方式已被其他账号绑定")
	}
	if err != nil {
		return internalError(err)
	}
	eventType := "bind_phone"
	masked := security.MaskPhone(identityValue)
	if identityType == dao.IdentityTypeEmail {
		eventType = "bind_email"
		masked = security.MaskEmail(identityValue)
	}
	s.logSecurity(ctx, userID, eventType, map[string]string{"value": masked}, meta)
	return nil
}

// unbindIdentity 解绑联系方式，必须通过当前密码校验并保留至少一种登录联系方式。
func (s *Service) unbindIdentity(ctx context.Context, userID uint64, identityType int32, currentPassword string, meta pagectx.RequestMeta) error {
	if err := s.requireDB(); err != nil {
		return err
	}
	user, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return internalError(err)
	}
	if user == nil {
		return apperror.Unauthorized(response.CodeUnauthorized, "登录态已失效")
	}
	if user.PasswordHash == "" {
		return apperror.BadRequest(response.CodeBadRequest, "请先设置密码后再解绑联系方式")
	}
	if !authlib.CheckPassword(user.PasswordHash, currentPassword) {
		return apperror.BadRequest(response.CodeBadRequest, "当前密码不正确")
	}
	// 邮箱和手机号至少保留一个，否则账号会失去可找回、可登录的联系方式。
	count, err := s.store.CountActiveIdentities(ctx, userID)
	if err != nil {
		return internalError(err)
	}
	if count <= 1 {
		return apperror.BadRequest(response.CodeBadRequest, "账号至少需要保留一种联系方式")
	}
	now := data.Now()
	if err := s.store.UnbindIdentity(ctx, userID, identityType, now); err != nil {
		return internalError(err)
	}
	eventType := "unbind_phone"
	if identityType == dao.IdentityTypeEmail {
		eventType = "unbind_email"
	}
	s.logSecurity(ctx, userID, eventType, map[string]string{}, meta)
	return nil
}
