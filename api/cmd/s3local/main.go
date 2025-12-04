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
	"github.com/tkasuz/s3local/internal/worker"
)

const (
	defaultPort     = "8080"
	defaultHost     = "0.0.0.0"
	defaultDBPath   = "s3local.db"
	shutdownTimeout = 30 * time.Second
)

// bucketPutHandler routes PUT /{bucket} requests based on query parameters
func bucketPutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("tagging") {
		bucket.PutBucketTagging(w, r)
		return
	}
	if r.URL.Query().Has("policy") {
		bucket.PutBucketPolicy(w, r)
		return
	}
	if r.URL.Query().Has("notification") {
		bucket.PutBucketNotificationConfiguration(w, r)
		return
	}
	bucket.CreateBucket(w, r)
}

// bucketGetHandler routes GET /{bucket} requests based on query parameters
func bucketGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("tagging") {
		bucket.GetBucketTagging(w, r)
		return
	}
	if r.URL.Query().Has("policy") {
		bucket.GetBucketPolicy(w, r)
		return
	}
	if r.URL.Query().Has("notification") {
		bucket.GetBucketNotificationConfiguration(w, r)
		return
	}
	object.ListObjectsV2(w, r)
}

// bucketDeleteHandler routes DELETE /{bucket} requests based on query parameters
func bucketDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("tagging") {
		bucket.DeleteBucketTagging(w, r)
		return
	}
	if r.URL.Query().Has("policy") {
		bucket.DeleteBucketPolicy(w, r)
		return
	}
	bucket.DeleteBucket(w, r)
}

// objectPutHandler routes PUT /{bucket}/{key} requests based on query parameters
func objectPutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("tagging") {
		object.PutObjectTagging(w, r)
		return
	}
	object.PutObject(w, r)
}

// objectGetHandler routes GET /{bucket}/{key} requests based on query parameters
func objectGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("tagging") {
		object.GetObjectTagging(w, r)
		return
	}
	object.GetObject(w, r)
}

// objectDeleteHandler routes DELETE /{bucket}/{key} requests based on query parameters
func objectDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("tagging") {
		object.DeleteObjectTagging(w, r)
		return
	}
	object.DeleteObject(w, r)
}

func registerRoutes(r chi.Router) {
	r.Get("/", bucket.ListBuckets)
	r.Route("/{bucket}", func(r chi.Router) {
		r.Use(ctx.WithBucketName())

		// Bucket operations with query parameter routing
		r.Put("/", bucketPutHandler)
		r.Get("/", bucketGetHandler)
		r.Delete("/", bucketDeleteHandler)
		r.Head("/", bucket.HeadBucket)

		// Use wildcard to match any object key path including nested paths and trailing slashes
		r.With(ctx.WithObjectKey()).Group(func(r chi.Router) {
			r.Put("/*", objectPutHandler)
			r.Get("/*", objectGetHandler)
			r.Head("/*", object.HeadObject)
			r.Delete("/*", objectDeleteHandler)
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

	database, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

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

	// Create and start notification worker
	notificationWorker := worker.NewNotificationWorker(store)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go notificationWorker.Start(workerCtx)

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

	// Stop the notification worker
	notificationWorker.Stop()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
