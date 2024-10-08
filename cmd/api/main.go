package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"database/sql"

	"timer-microservice/internal/config"
	"timer-microservice/internal/handlers"
	"timer-microservice/internal/repository"
	"timer-microservice/internal/server"
	"timer-microservice/internal/service"
	"timer-microservice/internal/websocket"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		sugar.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := sql.Open("mysql", cfg.GetDatabaseDSN())
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	gormDb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{})

	// Migrate the schema
	if err := repository.Migrate(gormDb); err != nil {
		sugar.Fatalf("Failed to migrate database schema: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.RedisPass,
		DB:       0, // use default DB
	})

	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		sugar.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repository
	repo := repository.NewTimerRepository(gormDb)

	// Initialize WebSocket handler
	wsHandler := websocket.NewHandler(nil, sugar)

	// Initialize service
	timerService := service.NewTimerService(repo, sugar, redisClient, wsHandler)

	// Set the service in WebSocket handler
	wsHandler.SetService(timerService)

	// Restore timers on startup
	err = timerService.RestoreTimers()
	if err != nil {
		sugar.Errorw("Failed to restore timers", "error", err)
	}

	go timerService.StartTimerUpdates()

	// Initialize handlers
	timerHandler := handlers.NewTimerHandler(timerService, sugar)

	// Initialize and start server
	srv := server.NewServer(cfg, sugar)
	srv.SetupRoutes(timerHandler, wsHandler)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
