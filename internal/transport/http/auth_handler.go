package http

import (
	"auth-todo-service/internal/client/task"
	"auth-todo-service/internal/domain/model"
	"auth-todo-service/internal/transport/http/dto"
	"errors"
	"log"
	"net/http"

	"auth-todo-service/internal/domain"
	"auth-todo-service/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	taskClient  task.TaskClient
}

func NewAuthHandler(authService *service.AuthService, taskClient task.TaskClient) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		taskClient:  taskClient,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.FullName, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrNameInvalid), errors.Is(err, domain.ErrPasswordTooWeak):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user_id": user.ID,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccToken, newRefreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "non valid refresh token"})
		return
	}

	response := dto.RefreshResponse{
		AccessToken:  newAccToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	u, ok := user.(*model.User)
	if !ok || u == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "non valid user data in context"})
		return
	}

	taskStats, err := h.taskClient.GetUserTaskStats(c.Request.Context(), u.ID)
	if err != nil {
		log.Printf("Warning: failed to get task stats for user %s: %v", u.ID, err)
		c.JSON(http.StatusInternalServerError, nil)
	}

	resp := dto.UserInfoResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		Tasks: dto.TasksInfo{
			AsAssignee: taskStats.AsAssignee,
			AsWatcher:  taskStats.AsWatcher,
		},
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) AuthHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"": "all good"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// TODO
	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}
