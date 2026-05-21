package client

import (
	"context"

	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RPCClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRPCClient(amqpURL string) (*RPCClient, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RPCClient{conn: conn, channel: ch}, nil
}

func (c *RPCClient) Call(ctx context.Context, queue string, payload interface{}) ([]byte, error) {
	corrID := uuid.NewString()

	// Создаём временную очередь для ответа
	q, err := c.channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare reply queue: %w", err)
	}

	// Подписываемся на ответ
	msgs, err := c.channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = c.channel.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		CorrelationId: corrID,
		ReplyTo:       q.Name,
		Body:          body,
		ContentType:   "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to publish rpc request: %w", err)
	}

	// Ждём ответ с таймаутом
	select {
	case msg := <-msgs:
		if msg.CorrelationId == corrID {
			return msg.Body, nil
		}
		return nil, fmt.Errorf("correlation id mismatch")
	case <-ctx.Done():
		return nil, fmt.Errorf("rpc call cancelled: %w", ctx.Err())
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("rpc timeout")
	}

}

func (c *RPCClient) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
