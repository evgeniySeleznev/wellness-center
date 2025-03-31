package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wellness-center/handlers"
	"wellness-center/models"
	"wellness-center/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Инициализация логгера
	logger := log.New(os.Stdout, "WELLNESS: ", log.LstdFlags|log.Lshortfile)

	// 2. Инициализация Redis
	redisClient, err := utils.NewRedisClient()
	if err != nil {
		logger.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Printf("Error closing Redis connection: %v", err)
		}
	}()

	// 3. Инициализация Kafka
	time.Sleep(15 * time.Second) // Даем Kafka время на запуск

	kafkaProducer, err := utils.NewKafkaProducer()
	if err != nil {
		logger.Printf("WARNING: Kafka connection failed (retrying...): %v", err)
		time.Sleep(5 * time.Second)
		// Повторная попытка
		kafkaProducer, err = utils.NewKafkaProducer()
	}

	// 4. Инициализация Elasticsearch
	esClient, err := utils.NewElasticsearchClient()
	if err != nil {
		logger.Fatalf("Failed to initialize Elasticsearch: %v", err)
	}

	// 5. Инициализация базы данных
	dbRepo, err := models.NewPostgresRepository()
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := dbRepo.Close(); err != nil {
			logger.Printf("Error closing database connection: %v", err)
		}
	}()

	// 6. Инициализация обработчиков
	clientHandler := handlers.NewClientHandler(dbRepo, kafkaProducer, redisClient, esClient)
	healthHandler := handlers.NewHealthHandler(dbRepo, redisClient, kafkaProducer, esClient)

	// 7. Настройка HTTP-сервера
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// API Routes
	api := router.Group("/api/v1")
	{
		api.POST("/clients", clientHandler.CreateClient)
		api.GET("/clients/:id", clientHandler.GetClient)
		api.PUT("/clients/:id", clientHandler.UpdateClient)
		api.GET("/search", clientHandler.SearchClients)
		api.GET("/health", healthHandler.CheckHealth)
	}

	// 8. Настройка graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в отдельной goroutine
	go func() {
		logger.Printf("Server is running on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидание сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	// Таймаут для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exited properly")
}
