package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nahualventure/class-backend/infra/shared/authorization"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	config := loadConfig()

	// Setup database connection pool
	pool, err := setupDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Setup authorization service
	authzService, err := setupAuthorization(pool, config.Tenants)
	if err != nil {
		log.Fatalf("Failed to setup authorization: %v", err)
	}
	defer authzService.Close()

	// Setup Gin router
	router := gin.Default()

	// Setup Huma API with Gin adapter
	humaConfig := huma.DefaultConfig("Class Backend API", "1.0.0")
	humaConfig.Info.Description = "A Go-based backend system with clean architecture and RBAC authorization"
	api := humagin.New(router, humaConfig)

	type HealthResponse struct {
		Body struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		}
	}

	huma.Register(api, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/health",
		Summary: "Health endpoint",
		Tags:    []string{"Health"},
	}, func(ctx context.Context, i *struct{}) (*HealthResponse, error) {
		return &HealthResponse{
			Body: struct {
				Message string `json:"message"`
				Status  string `json:"status"`
			}{
				Message: "Service is healthy",
				Status:  "OK",
			},
		}, nil
	})
	// TODO: Register routes here
	// registerAuthRoutes(api, pool, authzService)

	// Setup graceful shutdown
	setupGracefulShutdown()

	log.Println("Server started successfully!")
	log.Printf("HTTP API: http://localhost:%s", config.HTTPPort)
	log.Printf("API Documentation: http://localhost:%s/docs", config.HTTPPort)

	// Start HTTP server
	if err := router.Run(":" + config.HTTPPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type Config struct {
	DatabaseURL string
	GRPCPort    string
	HTTPPort    string
	Tenants     []string
}

func loadConfig() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5437/edoo_class?sslmode=disable"),
		GRPCPort:    getEnv("GRPC_PORT", "8080"),
		HTTPPort:    getEnv("HTTP_PORT", "8081"),
		Tenants:     []string{"tenant1", "tenant2"}, // TODO: Load from environment or database
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupDatabase(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Connecting to database: %s", maskPassword(databaseURL))

	// Create connection pool
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database with connection pool")
	return pool, nil
}

func setupAuthorization(pool *pgxpool.Pool, tenants []string) (*authorization.CasbinService, error) {
	// Convert pgxpool to database/sql for Casbin adapter
	sqlDB := stdlib.OpenDBFromPool(pool)

	authzService, err := authorization.NewCasbinService(
		sqlDB,
		"infra/configs/rbac_model.conf",
		"policies.yaml",
		tenants, // should be loaded from database
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization service: %w", err)
	}

	log.Printf("Authorization service initialized for tenants: %v", tenants)
	return authzService, nil
}

func maskPassword(databaseURL string) string {
	// Mask password in log output for security
	parts := strings.Split(databaseURL, "@")
	if len(parts) < 2 {
		return databaseURL
	}

	userInfo := parts[0]
	userParts := strings.Split(userInfo, ":")
	if len(userParts) < 2 {
		return databaseURL
	}

	maskedURL := strings.Join(userParts[:len(userParts)-1], ":") + ":***@" + strings.Join(parts[1:], "@")
	return maskedURL
}

func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		os.Exit(0)
	}()
}
