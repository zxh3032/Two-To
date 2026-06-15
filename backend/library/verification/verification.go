package verification

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/zxh3032/two-to/backend/library/cache"
	"github.com/zxh3032/two-to/backend/library/security"
)

const (
	// CodeTypeImage 表示图片验证码。
	CodeTypeImage int32 = 1
	// CodeTypeSMS 表示短信验证码。
	CodeTypeSMS int32 = 2
	// CodeTypeEmail 表示邮箱验证码。
	CodeTypeEmail int32 = 3

	// SceneLogin 表示登录场景验证码。
	SceneLogin = "login"
	// SceneRegister 表示注册场景验证码。
	SceneRegister = "register"
	// SceneForgotPassword 表示找回密码场景验证码。
	SceneForgotPassword = "forgot_password"
	// SceneBindPhone 表示绑定或换绑手机号场景验证码。
	SceneBindPhone = "bind_phone"
	// SceneBindEmail 表示绑定或换绑邮箱场景验证码。
	SceneBindEmail = "bind_email"
)

var captchaChars = []rune("23456789ABCDEFGHJKLMNPQRSTUVWXYZ")

// SetupPayload 是注册资料完善 token 中缓存的已验证联系方式和密码摘要。
type SetupPayload struct {
	Kind         string `json:"kind"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	PasswordHash string `json:"passwordHash,omitempty"`
}

// ResetPayload 是找回密码 reset token 中缓存的目标用户。
type ResetPayload struct {
	UserID uint64 `json:"userId"`
}

// NewDigitCode 生成固定长度数字验证码，首期短信和邮箱都使用 6 位数字。
func NewDigitCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var builder strings.Builder
	for i := 0; i < length; i++ {
		builder.WriteByte(byte('0' + r.Intn(10)))
	}
	return builder.String()
}

// NewCaptcha 生成图片验证码 ID、答案和可直接给前端展示的 base64 SVG。
func NewCaptcha() (id string, code string, imageBase64 string, err error) {
	token, err := security.RandomToken(18)
	if err != nil {
		return "", "", "", err
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var builder strings.Builder
	for i := 0; i < 4; i++ {
		builder.WriteRune(captchaChars[r.Intn(len(captchaChars))])
	}
	code = builder.String()
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="160" height="56" viewBox="0 0 160 56">
<rect width="160" height="56" rx="8" fill="#fff7ed"/>
<path d="M8 43 C38 12,68 48,104 16 S140 22,152 8" fill="none" stroke="#f9735b" stroke-width="3" opacity=".38"/>
<path d="M10 18 C45 54,85 2,150 42" fill="none" stroke="#2f9e78" stroke-width="2" opacity=".35"/>
<text x="80" y="37" text-anchor="middle" font-family="Verdana, sans-serif" font-size="27" font-weight="700" letter-spacing="6" fill="#3a2418">%s</text>
</svg>`, code)
	return token, code, "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg)), nil
}

// StoreJSON 将短期流程数据序列化后写入 cache，例如资料完善 token 和重置密码 token。
func StoreJSON(ctx context.Context, store cache.Store, key string, value interface{}, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.Set(ctx, key, string(raw), ttl)
}

// LoadJSON 从 cache 读取并反序列化短期流程数据。
func LoadJSON[T any](ctx context.Context, store cache.Store, key string) (*T, error) {
	raw, err := store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var value T
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, err
	}
	return &value, nil
}
