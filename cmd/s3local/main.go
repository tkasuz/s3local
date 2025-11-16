package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/tkasuz/s3local/internal/adapters/http/bucket"
	"github.com/tkasuz/s3local/internal/adapters/http/object"
	bucketdomain "github.com/tkasuz/s3local/internal/domain/bucket"
	objectdomain "github.com/tkasuz/s3local/internal/domain/object"
	appmiddleware "github.com/tkasuz/s3local/internal/middleware"
	"github.com/tkasuz/s3local/internal/storage"
)

const (
	defaultPort     = "8080"
	defaultHost     = "0.0.0.0"
	defaultDBPath   = "s3local.db"
	shutdownTimeout = 30 * time.Second
)

func main() {
	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	// Initialize database storage adapter
	log.Printf("Initializing storage adapter (database: %s)...", dbPath)
	dbAdapter, err := storage.NewAdapter(storage.Config{
		Type:   storage.StorageTypeSQLite,
		DBPath: dbPath,
	})
	if err != nil {
		log.Fatalf("Failed to create storage adapter: %v", err)
	}

	db, err := dbAdapter.Open()
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbAdapter.Close()

	log.Printf("%s storage initialized successfully", dbAdapter.Name())

	// Initialize domain services
	bucketService := bucketdomain.NewService()
	objectService := objectdomain.NewService()

	// Initialize REST handlers with dependency injection
	bucketHandler := bucket.NewBucketHandler(bucketService)
	objectHandler := object.NewObjectHandler(objectService)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Add custom middleware to inject queries into context
	r.Use(appmiddleware.QueriesMiddleware(db))

	// CORS middleware for S3 compatibility
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"ETag", "x-amz-*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Mount S3-compatible REST API (root level)
	bucketHandler.RegisterRoutes(r)
	objectHandler.RegisterRoutes(r)

	// Create server with HTTP/2 support
	addr := fmt.Sprintf("%s:%s", host, port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(r, &http2.Server{}),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting s3local server on %s", addr)
		log.Printf("S3-compatible REST API: http://%s", addr)
		log.Printf("Health check: http://%s/health", addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
