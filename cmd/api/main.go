package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tushar-kalsi/product-views/internal/config"
	"github.com/tushar-kalsi/product-views/internal/handlers"
	"github.com/tushar-kalsi/product-views/internal/kafka"
	"github.com/tushar-kalsi/product-views/internal/repository"
	// Swagger support is optional for now, commenting out
	// _ "github.com/tushar-kalsi/product-views/docs"
)

// @title Product Views API
// @version 1.0
// @description API for tracking product views and getting top viewed products
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Run migrations
	log.Println("Running database migrations...")
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBroker, "product-views")
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Initialize repositories and handlers
	productRepo := repository.NewProductRepository(db.GetConn())
	productHandler := handlers.NewProductHandler(productRepo, kafkaProducer)

	// Start Kafka consumer in the background
	kafkaConsumer, err := kafka.NewConsumer(
		cfg.KafkaBroker,
		"product-views-consumer",
		"product-views",
		productRepo,
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Stop()

	if err := kafkaConsumer.Start(); err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}

	// Set up HTTP server
	router := setupRouter(productHandler)

	// Start server in a goroutine
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Server is starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Set a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func setupRouter(handler *handlers.ProductHandler) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger documentation - commented out for now
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", handler.CreateProduct)
			products.GET(":id", handler.GetProduct)
			products.GET("top", handler.GetTopProducts)
			products.POST("view", handler.ViewProduct)
		}
	}

	return router
}
