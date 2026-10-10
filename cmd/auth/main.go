package main

import (
	"auth-todo-service/internal/client"
	"auth-todo-service/internal/client/task"
	"auth-todo-service/internal/publisher"
	"auth-todo-service/internal/repository/postgres"
	"auth-todo-service/internal/transport"
	httpv1 "auth-todo-service/internal/transport/http"
	"auth-todo-service/internal/transport/middleware"
	"auth-todo-service/internal/transport/rpc"
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-todo-service/internal/config"
	"auth-todo-service/internal/repository"
	"auth-todo-service/internal/service"
	"auth-todo-service/internal/token"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к БД
	db, err := postgres.NewConnection(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	rabbitConn, err := amqp091.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	rabbitMq, err := client.NewRPCClient(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMq.Close()

	rabbitCh, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}
	defer rabbitCh.Close()

	eventPublisher := publisher.NewPublisher(rabbitCh)

	// Инициализация слоёв
	userRepo := repository.NewUserRepository(db)
	jwtManager := token.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	taskClient := task.NewTaskClient(rabbitMq) //поменять на конфиг круто
	authService := service.NewAuthService(userRepo, jwtManager, eventPublisher)
	authHandler := httpv1.NewAuthHandler(authService, taskClient)

	rpcServer, err := rpc.NewServer(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("failed to create rpc server: %w", err)
	}
	rpcHandler := rpc.NewHandler(authService)

	err = rpcServer.Start("auth-service.rpc", rpcHandler.GetUserInfoHandler)
	if err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}
	log.Println("RPC Server started on queue: auth-service.rpc")

	// Gin Setup
	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	// Настраиваем все роуты
	transport.SetupRoutes(authHandler, authMiddleware, router)

	// Запуск сервера
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Auth Service started on http://localhost:%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Auth Service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
}
