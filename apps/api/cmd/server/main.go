package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/soumabali/vexa/config"
	"github.com/soumabali/vexa/internal/api"
	"github.com/soumabali/vexa/internal/db"
	"github.com/soumabali/vexa/internal/security"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	router := api.SetupRouter(cfg, database.DB, redisClient)

	tlsConfig, err := cfg.GetTLSConfig()
	if err != nil {
		log.Fatalf("Failed to load TLS config: %v", err)
	}

	scheme := "http"
	if tlsConfig != nil {
		scheme = "https"
	}

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if cfg.IsProduction() {
		server.ReadTimeout = 30 * time.Second
		server.WriteTimeout = 30 * time.Second
	}

	go func() {
		log.Printf("%s://0.0.0.0:%s", scheme, cfg.ServerPort)

		if tlsConfig != nil {
			listener, err := security.SecureListener(":"+cfg.ServerPort, tlsConfig)
			if err != nil {
				log.Fatalf("Failed to create TLS listener: %v", err)
			}
			if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server failed: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server failed: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
