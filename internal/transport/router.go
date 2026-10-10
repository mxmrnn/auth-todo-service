package transport

import (
	"auth-todo-service/internal/transport/http"
	"auth-todo-service/internal/transport/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(handler *http.AuthHandler, authMiddleware *middleware.AuthMiddleware, router *gin.Engine) {

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
			auth.POST("/logout", handler.Logout)
			auth.POST("/refresh", handler.RefreshToken)
		}

		p := v1.Group("/")
		p.Use(authMiddleware.AuthRequired())
		{
			p.GET("/user/info/", handler.GetUserInfo)
			p.GET("/health", handler.AuthHealth)
		}

	}
}
