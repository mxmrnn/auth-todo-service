package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auth-todo-service/internal/transport/http/dto"
	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	channel *amqp091.Channel
}

func NewPublisher(ch *amqp091.Channel) *Publisher {
	return &Publisher{channel: ch}
}

// PublishUserRegistered публикует событие регистрации
func (p *Publisher) PublishUserRegistered(ctx context.Context, user dto.UserResponse) error {
	event := map[string]interface{}{
		"event_type": "user_registered",
		"user_id":    user.ID,
		"email":      user.Email,
		"full_name":  user.FullName,
		"timestamp":  time.Now().UTC(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.PublishWithContext(ctx,
		"email.notifications",
		"",
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish registration event: %w", err)
	}

	return nil
}
