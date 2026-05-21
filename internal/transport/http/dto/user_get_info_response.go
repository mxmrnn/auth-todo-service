package dto

import (
	"auth-todo-service/internal/domain/model"
	"github.com/google/uuid"
)

type UserGetInfoResponse struct {
	ID       uuid.UUID      `json:"id"`
	FullName string         `json:"full_name"`
	Email    string         `json:"email,omitempty"`
	Role     model.UserRole `json:"role,omitempty"`
}
