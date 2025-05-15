package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/patrickmn/go-cache"
)

// 内存缓存，默认清理间隔 1 小时，项过期时间 24 小时
var secretCache = cache.New(24*time.Hour, 1*time.Hour)

// 缓存中存储密钥的键名
const cacheKey = "jwt_secret_base64"

// MyClaims 自定义的 JWT Claims，继承了 StandardClaims
type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenerateOrGetSecret 从缓存中读取 Base64 编码的密钥；若不存在则生成并存入缓存
func GenerateOrGetSecret(length int) ([]byte, error) {
	if val, found := secretCache.Get(cacheKey); found {
		if b64, ok := val.(string); ok {
			if secret, err := base64.RawURLEncoding.DecodeString(b64); err == nil {
				return secret, nil
			}
		}
	}
	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("生成随机密钥失败: %w", err)
	}
	b64 := base64.RawURLEncoding.EncodeToString(key)
	secretCache.Set(cacheKey, b64, cache.DefaultExpiration)
	return key, nil
}

// GenerateToken 签发一个**固定 10 分钟**后过期的 JWT。
// 内部自动从缓存读取（或生成）签名密钥。
func GenerateToken(user string) (string, error) {
	secret, err := GenerateOrGetSecret(32)
	if err != nil {
		return "", err
	}
	claims := MyClaims{
		Username: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(), // 过期时间
			IssuedAt:  time.Now().Unix(),                       // 签发时间
			Issuer:    "miragecore",                            // 签发者
			Subject:   fmt.Sprint(user),                        // 主题
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析并验证 JWT，返回 *MyClaims
func ParseToken(tokenString string) (*MyClaims, error) {
	secret, err := GenerateOrGetSecret(32)
	if err != nil {
		return nil, err
	}
	// ParseWithClaims 的第二个参数传入 &MyClaims{}，这样解析后的 Claims 就是 MyClaims
	tok, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 可选：额外校验签名方法
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	// 断言成 *MyClaims
	if claims, ok := tok.Claims.(*MyClaims); ok && tok.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
