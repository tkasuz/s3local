package main

import (
	"context"
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/tkasuz/s3local/internal/config"
	"github.com/tkasuz/s3local/internal/db"
	"github.com/tkasuz/s3local/internal/handlers/bucket"
	"github.com/tkasuz/s3local/internal/handlers/ctx"
	"github.com/tkasuz/s3local/internal/handlers/object"
)

const (
	defaultPort     = "8080"
	defaultHost     = "0.0.0.0"
	defaultDBPath   = "s3local.db"
	shutdownTimeout = 30 * time.Second
)

func registerRoutes(r chi.Router) {
	r.Get("/", bucket.ListBuckets)
	r.Route("/{bucket}", func(r chi.Router) {
		r.Put("/", bucket.CreateBucket)
		r.Delete("/", bucket.DeleteBucket)
		r.Head("/", bucket.HeadBucket)
		r.Get("/", object.ListObjectsV2)
		r.Route("/{key:.*}", func(r chi.Router) {
			r.Put("/", object.PutObject)
			r.Get("/", object.GetObject)
			r.Delete("/", object.DeleteObject)
		})
	})
}

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

	dbURL := fmt.Sprintf("sqlite://%s?_fk=1", dbPath)
	if err := db.RunMigrations(dbURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	database, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	queries := db.New(database)
	store := db.NewStore(database, queries)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(ctx.WithConfig(cfg))
	r.Use(ctx.WithStore(store))

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

	registerRoutes(r)

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
