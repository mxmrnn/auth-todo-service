package rpc

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewServer создаёт RPC сервер
func NewServer(amqpURL string) (*Server, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Server{
		conn:    conn,
		channel: ch,
	}, nil
}

// Start запускает слушателя входящих RPC запросов
func (s *Server) Start(queueName string, handler func([]byte) ([]byte, error)) error {
	// Объявляем очередь
	_, err := s.channel.QueueDeclare(
		queueName, // "auth-service.rpc"
		true,      // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	msgs, err := s.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // autoAck = false (вручную)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for msg := range msgs {
			response, err := handler(msg.Body)
			if err != nil {
				log.Printf("[RPC] Handler error: %v", err)
				response, _ = json.Marshal(map[string]string{
					"error":   "internal_error",
					"message": err.Error(),
				})
			}

			// Отправляем ответ
			err = s.channel.Publish(
				"",          // exchange
				msg.ReplyTo, // куда отправить ответ
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: msg.CorrelationId,
					Body:          response,
				},
			)
			if err != nil {
				log.Printf("[RPC] Failed to publish response: %v", err)
			}

			msg.Ack(false) // подтверждаем обработку
		}
	}()

	log.Printf("✅ RPC Server started on queue: %s", queueName)
	return nil
}

// Close закрывает соединение
func (s *Server) Close() error {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
