package jwt

import (
	"autoSSL/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    int      `json:"userId"`
	TokenType string   `json:"tokenType"`       // "access"或"refresh"
	Roles     []string `json:"roles,omitempty"` // 用户角色列表
	jwt.RegisteredClaims
}

type JWTManager struct {
	Secret        []byte
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

// New 根据配置创建JWTManager实例
func New(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		Secret:        []byte(cfg.Secret),
		AccessExpire:  time.Duration(cfg.AccessExpire),
		RefreshExpire: time.Duration(cfg.RefreshExpire),
	}
}

func (m *JWTManager) GenerateAccessToken(userID int, roles []string) (string, error) {
	return m.generateToken(userID, roles, m.AccessExpire, "access")
}

func (m *JWTManager) GenerateRefreshToken(userID int, roles []string) (string, error) {
	return m.generateToken(userID, roles, m.RefreshExpire, "refresh")
}

func (m *JWTManager) generateToken(userID int, roles []string, duration time.Duration, tokenType string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "autoSSL",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.Secret)
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.Secret, nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
