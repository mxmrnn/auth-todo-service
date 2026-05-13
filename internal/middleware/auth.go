package middleware

import (
	"errors"
	"net/http"
	"strings"

	"auth-todo-service/internal/domain"
	"auth-todo-service/internal/service"
	// "auth-todo-service/internal/token"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// AuthRequired - основной middleware для защиты роутов
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Валидируем токен
		user, err := m.authService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			if errors.Is(err, domain.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			}
			c.Abort()
			return
		}

		// Сохраняем пользователя в контекст
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		c.Next() // продолжаем выполнение handler'а
	}
}
