package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"

	//"time"

	"auth-todo-service/internal/domain"
	"auth-todo-service/internal/repository"
	"auth-todo-service/internal/token"

	//"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   repository.UserRepository
	jwtManager *token.JWTManager
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *token.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, fullName, email, password string) (*domain.User, error) {
	// Проверяем, существует ли пользователь
	exists, err := s.userRepo.IsEmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(fullName, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", "", domain.ErrInvalidCredentials
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", domain.ErrInvalidCredentials
	}

	accessToken, err = s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshTTL()),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	err = s.userRepo.CreateRefreshToken(ctx, refToken)
	if err != nil {
		return "", "", fmt.Errorf("error in insert ref token in db: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	claims, err := s.jwtManager.ParseAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	rt, err := s.userRepo.GetRefreshTokenByValue(ctx, refreshToken)
	if err != nil {
		return "", "", domain.ErrInvalidRefreshToken
	}

	if rt.Revoked {
		return "", "", domain.ErrRefreshTokenRevoked
	}
	if rt.ExpiresAt.Before(time.Now()) {
		return "", "", domain.ErrRefreshTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", err
	}

	if err := s.userRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		log.Printf("Warning: failed to revoke old refresh token: %v", err)
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	newRefToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshTTL()), // или 7*24*time.Hour
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.userRepo.CreateRefreshToken(ctx, newRefToken); err != nil {
		return "", "", fmt.Errorf("failed to save new refresh token: %w", err)
	}
	// Добавить удаление старого рефреш токена

	return accessToken, newRefreshToken, nil
}
