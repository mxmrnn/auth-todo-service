package task

import (
	"auth-todo-service/internal/client"
	"auth-todo-service/internal/transport/http/dto"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"

	"strings"
)

type TaskClient interface {
	GetUserTaskStats(ctx context.Context, userID uuid.UUID) (*dto.UserTaskStats, error)
}

type taskClient struct {
	baseURL string
	rpc     *client.RPCClient
}

func NewTaskClient(baseURL string, client *client.RPCClient) TaskClient {
	return &taskClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		rpc:     client,
	}
}

func (c *taskClient) GetUserTaskStats(ctx context.Context, userID uuid.UUID) (*dto.UserTaskStats, error) {
	payload := struct {
		UserID string `json:"user_id"`
	}{
		UserID: userID.String(),
	}

	resp, err := c.rpc.Call(ctx, "task-service.rpc", payload)
	if err != nil {
		return nil, fmt.Errorf("task-service rpc failed: %w", err)
	}

	var stats dto.UserTaskStats
	if err := json.Unmarshal(resp, &stats); err != nil {
		return nil, fmt.Errorf("failed to decode task stats: %w", err)
	}

	return &stats, nil
}
