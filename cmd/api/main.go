package main

import (
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

	// Initialize repository
	repo := repository.NewTimerRepository(gormDb)

	// Initialize service
	timerService := service.NewTimerService(repo, sugar, redisClient)

	// Restore timers on startup
	err = timerService.RestoreTimers()
	if err != nil {
		sugar.Errorw("Failed to restore timers", "error", err)
	}

	// Initialize handlers
	timerHandler := handlers.NewTimerHandler(timerService, sugar)
	wsHandler := handlers.NewWebSocketHandler(timerService, sugar)

	// Initialize and start server
	srv := server.NewServer(cfg, sugar)
	srv.SetupRoutes(timerHandler, wsHandler)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
