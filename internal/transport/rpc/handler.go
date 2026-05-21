package rpc

import (
	"auth-todo-service/internal/service"
	"auth-todo-service/internal/transport/http/dto"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.AuthService
}

func NewHandler(svc *service.AuthService) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) GetUserInfoHandler(payload []byte) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := dto.UserGetInfoRequest{}
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if req.UserID == " " {
		return nil, fmt.Errorf("user_id is required")
	}

	id, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id uuid format: %w", err)
	}

	user, err := h.service.GetUserByID(ctx, id)
	if err != nil {
		return json.Marshal(map[string]string{
			"code":    "internal error",
			"message": err.Error(),
		})
	}

	response := dto.UserGetInfoResponse{
		ID:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
	}

	return json.Marshal(response)
}
