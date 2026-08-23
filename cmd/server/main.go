package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Sall-lah/store_user/internal/client/order"
	"github.com/Sall-lah/store_user/internal/config"
	"github.com/Sall-lah/store_user/internal/db"
	"github.com/Sall-lah/store_user/internal/handler"
	"github.com/Sall-lah/store_user/internal/kafka"
	"github.com/Sall-lah/store_user/internal/ratelimit"
	"github.com/Sall-lah/store_user/internal/repository"
	"github.com/Sall-lah/store_user/internal/router"
	"github.com/Sall-lah/store_user/internal/service"
)

func main() {
	log.Println("Starting store_user microservice...")

	// 1. Load Environment Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load environment configuration: %v", err)
	}

	// 2. Initialize Prisma Database Client
	prismaClient := db.NewClient()
	if err := prismaClient.Prisma.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL via Prisma: %v (service will still run)", err)
	}
	defer func() {
		if err := prismaClient.Prisma.Disconnect(); err != nil {
			log.Printf("Error disconnecting from database: %v", err)
		}
	}()

	// 3. Initialize Redis Rate Limiter
	var limiter ratelimit.Limiter
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Failed to parse Redis URL %s: %v. Rate limiter running degraded/in-memory fallback.", cfg.RedisURL, err)
		limiter = ratelimit.NewRedisLimiter(nil, false, 50*time.Millisecond)
	} else {
		if cfg.RedisPassword != "" {
			redisOpts.Password = cfg.RedisPassword
		}
		rdb := redis.NewClient(redisOpts)
		limiter = ratelimit.NewRedisLimiter(rdb, cfg.EnableRateLimiter, 50*time.Millisecond)
	}
	defer limiter.Close()

	// 4. Initialize Kafka Event Producer
	kafkaProducer := kafka.NewProducer(cfg.KafkaBrokers)
	defer kafkaProducer.Close()

	// 5. Initialize gRPC Client to store_order
	orderClient, err := order.NewOrderClient(cfg.OrderServiceGrpcAddr, cfg.OrderServiceGrpcTimeout)
	if err != nil {
		log.Printf("Warning: Failed to initialize order gRPC client: %v", err)
	} else {
		defer orderClient.Close()
	}

	// 6. Assemble Architecture Layers
	repo := repository.NewPrismaUserProfileRepository(prismaClient)
	svc := service.NewUserService(repo, orderClient, kafkaProducer, cfg.KafkaTopicUserEvents)
	profileHandler := handler.NewProfileHandler(svc)
	httpRouter := router.NewRouter(cfg, profileHandler, limiter)

	// 7. Create HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      httpRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Run HTTP Server in Background Goroutine
	go func() {
		log.Printf("store_user HTTP server listening on %s (Environment: %s)", serverAddr, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server encountered fatal error: %v", err)
		}
	}()

	// 9. Listen for OS Termination Signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stopChan
	log.Printf("Received shutdown signal (%s). Initiating graceful termination...", sig)

	// 10. Execute Graceful Shutdown with Timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced shutdown: %v", err)
	}

	log.Println("store_user microservice terminated cleanly.")
}
