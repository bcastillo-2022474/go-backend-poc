package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"class-backend/class/auth/handlers"
	"class-backend/class/shared/authorization"
	authv1 "class-backend/proto/generated/go/auth/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
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

	// Setup servers
	var wg sync.WaitGroup

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startGRPCServer(config.GRPCPort, pool, authzService); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Start HTTP gateway server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startHTTPGateway(config.HTTPPort, config.GRPCPort); err != nil {
			log.Fatalf("HTTP gateway failed: %v", err)
		}
	}()

	// Graceful shutdown
	setupGracefulShutdown()

	log.Println("Server started successfully!")
	log.Printf("gRPC server: localhost:%s", config.GRPCPort)
	log.Printf("HTTP API: http://localhost:%s", config.HTTPPort)
	log.Printf("Signup endpoint: POST http://localhost:%s/api/v1/auth/signup", config.HTTPPort)
	log.Printf("OpenAPI spec: http://localhost:%s/openapi.json", config.HTTPPort)

	wg.Wait()
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

	// Initialize Casbin service
	authzService, err := authorization.NewCasbinService(
		sqlDB,
		"configs/rbac_model.conf",
		"policies.yaml",
		tenants,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization service: %w", err)
	}

	log.Printf("Authorization service initialized for tenants: %v", tenants)
	return authzService, nil
}

func maskPassword(databaseURL string) string {
	// Mask password in log output for security
	if strings.Contains(databaseURL, "@") {
		parts := strings.Split(databaseURL, "@")
		if len(parts) >= 2 {
			userInfo := parts[0]
			if strings.Contains(userInfo, ":") {
				userParts := strings.Split(userInfo, ":")
				if len(userParts) >= 2 {
					return strings.Join(userParts[:len(userParts)-1], ":") + ":***@" + strings.Join(parts[1:], "@")
				}
			}
		}
	}
	return databaseURL
}

func startGRPCServer(port string, pool *pgxpool.Pool, authzService *authorization.CasbinService) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	// Create gRPC server with authorization middleware
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authorization.AuthorizationInterceptor(authzService),
			loggingInterceptor,
		),
	)

	// Register services
	authHandler := handlers.NewAuthHandler(pool)
	authv1.RegisterAuthServiceServer(grpcServer, authHandler)

	// Enable reflection for development
	reflection.Register(grpcServer)

	log.Printf("Starting gRPC server on :%s", port)
	return grpcServer.Serve(lis)
}

func startHTTPGateway(httpPort, grpcPort string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create gRPC-Gateway mux
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "Authorization", "X-User-Id", "X-Trace-Id":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)

	// Register AuthService
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := authv1.RegisterAuthServiceHandlerFromEndpoint(
		ctx,
		mux,
		"localhost:"+grpcPort,
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Serve OpenAPI spec files
	mux.HandlePath("GET", "/openapi.json", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "proto/generated/openapi/auth/v1/auth.swagger.json")
	})

	// Add CORS and logging middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting HTTP gateway on :%s", httpPort)
	return server.ListenAndServe()
}

// Middleware functions
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	status := "OK"
	if err != nil {
		status = "ERROR"
	}

	log.Printf("gRPC %s %s %v [%s]", info.FullMethod, duration, status, err)
	return resp, err
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-Id, X-Trace-Id")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("HTTP %s %s %v", r.Method, r.URL.Path, duration)
	})
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
