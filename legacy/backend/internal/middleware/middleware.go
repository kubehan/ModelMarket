package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"modelmarket/internal/config"
	"modelmarket/pkg/logger"
)

// Claims JWT 自定义载荷
type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"usr"`
	jwt.RegisteredClaims
}

// IssueToken 颁发 JWT
func IssueToken(uid uint, username string) (string, error) {
	cfg := config.Global
	claims := Claims{
		UserID:   uid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWTExpiresHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "modelmarket",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(cfg.JWTSecret))
}

// AuthRequired 校验 Authorization: Bearer xxx
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			logger.L().Warnf("AuthRequired: missing token from %s", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(config.Global.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			logger.L().Warnf("AuthRequired: invalid token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// RequestLogger 请求日志
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		if status >= 500 {
			logger.L().Errorf("[%d] %s %s %v ip=%s", status, c.Request.Method, c.Request.URL.Path, latency, c.ClientIP())
		} else if status >= 400 {
			logger.L().Warnf("[%d] %s %s %v ip=%s", status, c.Request.Method, c.Request.URL.Path, latency, c.ClientIP())
		} else {
			logger.L().Infof("[%d] %s %s %v ip=%s", status, c.Request.Method, c.Request.URL.Path, latency, c.ClientIP())
		}
	}
}
