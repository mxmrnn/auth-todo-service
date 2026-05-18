package middleware

import (
	"auth-todo-service/internal/token"
	// "auth-todo-service/internal/token"

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
			if err == token.ErrTokenExpired {
				status = 401 // или 403
			}
			c.JSON(status, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

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
