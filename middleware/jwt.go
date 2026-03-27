package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JwtValidator JWT 验证接口
type JwtValidator interface {
	ValidateToken(c *gin.Context, claims jwt.Claims) bool
	GetUser(c *gin.Context, claims jwt.Claims) bool
}

// JwtConfig JWT 中间件配置
type JwtConfig struct {
	Secret        []byte
	SigningMethod jwt.SigningMethod
	TokenHeader   string
	TokenPrefix   string
	Validator     JwtValidator
}

// JwtClaimsFunc Claims 工厂函数
type JwtClaimsFunc func() jwt.Claims

// JwtAuth 创建 JWT 认证中间件
func JwtAuth(cfg JwtConfig, claimsFunc JwtClaimsFunc) gin.HandlerFunc {
	if cfg.SigningMethod == nil {
		cfg.SigningMethod = jwt.SigningMethodHS256
	}
	if cfg.TokenHeader == "" {
		cfg.TokenHeader = "Authorization"
	}
	// 如果 TokenPrefix 未设置，使用默认值 "Bearer "
	// 如果要禁用前缀，将 TokenPrefix 设置为 "!"（一个不会出现在 token 开头的字符）
	if cfg.TokenPrefix == "" {
		cfg.TokenPrefix = "Bearer "
	}

	return func(c *gin.Context) {
		tokenStr := c.GetHeader(cfg.TokenHeader)
		if tokenStr == "" {
			RespondErrorWithCode(c, ErrMissingToken)
			c.Abort()
			return
		}

		// 删除 Token 前缀
		if cfg.TokenPrefix != "!" && len(tokenStr) >= len(cfg.TokenPrefix) {
			tokenStr = tokenStr[len(cfg.TokenPrefix):]
		}

		var claims jwt.Claims
		if claimsFunc != nil {
			claims = claimsFunc()
		}
		if claims == nil {
			claims = &jwt.RegisteredClaims{}
		}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return cfg.Secret, nil
		})

		if err != nil || !token.Valid {
			RespondErrorWithCode(c, ErrInvalidToken)
			c.Abort()
			return
		}

		if cfg.Validator != nil {
			if !cfg.Validator.ValidateToken(c, claims) {
				RespondErrorWithCode(c, ErrTokenBlocked)
				c.Abort()
				return
			}
			if !cfg.Validator.GetUser(c, claims) {
				RespondErrorWithCode(c, ErrUserNotFound)
				c.Abort()
				return
			}
		}

		c.Set("jwt_claims", claims)
		c.Next()
	}
}

// GetJwtClaims 从 context 获取 Claims
func GetJwtClaims[T jwt.Claims](c *gin.Context) T {
	claims, _ := c.Get("jwt_claims")
	var zero T
	if claims == nil {
		return zero
	}
	return claims.(T)
}
