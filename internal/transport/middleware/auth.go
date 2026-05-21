package middleware

import (
	"auth-todo-service/internal/domain/model"
	"auth-todo-service/internal/token"
	"errors"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwt *token.JWTManager
}

func NewAuthMiddleware(jwtManager *token.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwt: jwtManager,
	}
}

// AuthRequired - основной middleware для защиты роутов
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.JSON(401, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		claims, err := m.jwt.ParseAccessToken(tokenStr)
		if err != nil {
			status := 401
			if errors.Is(err, token.ErrTokenExpired) {
				status = 401 // или 403
			}
			c.JSON(status, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		user := &model.User{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  claims.Role,
		}
		c.Set("user", user)

		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}
