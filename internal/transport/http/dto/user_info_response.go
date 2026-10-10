package dto

import (
	"auth-todo-service/internal/domain/model"
	"github.com/google/uuid"
	"time"
)

type UserInfoResponse struct {
	ID        uuid.UUID      `json:"id"`
	FullName  *string        `json:"full_name"`
	Email     string         `json:"email"`
	Role      model.UserRole `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	Tasks     TasksInfo      `json:"tasks"`
}

type TasksInfo struct {
	AsAssignee int `json:"as_assignee"`
	AsWatcher  int `json:"as_watcher"`
}
