package handlers

import (
	"context"
	"net/http"
	"time"
	"wellness-center/models"
	"wellness-center/utils"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db      models.Repository
	redis   utils.RedisClient
	kafka   utils.KafkaProducer
	elastic utils.ElasticsearchClient
}

func NewHealthHandler(
	db models.Repository,
	redis utils.RedisClient,
	kafka utils.KafkaProducer,
	elastic utils.ElasticsearchClient,
) *HealthHandler {
	return &HealthHandler{
		db:      db,
		redis:   redis,
		kafka:   kafka,
		elastic: elastic,
	}
}

func (h *HealthHandler) CheckHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	status := gin.H{
		"status":  "ok",
		"details": gin.H{},
	}
	httpStatus := http.StatusOK

	// Проверяем подключение к Redis
	if _, err := h.redis.GetFromCache(ctx, "healthcheck"); err != nil {
		status["details"].(gin.H)["redis"] = "unavailable"
		status["status"] = "degraded"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status["details"].(gin.H)["redis"] = "available"
	}

	// Проверяем подключение к Kafka
	if err := h.kafka.SendToKafka("healthcheck", "ping"); err != nil {
		status["details"].(gin.H)["kafka"] = "unavailable"
		status["status"] = "degraded"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status["details"].(gin.H)["kafka"] = "available"
	}

	// Проверяем подключение к Elasticsearch
	if _, err := h.elastic.Search(ctx, "healthcheck", ""); err != nil {
		status["details"].(gin.H)["elasticsearch"] = "unavailable"
		status["status"] = "degraded"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status["details"].(gin.H)["elasticsearch"] = "available"
	}

	// Проверяем подключение к БД
	if err := h.dbHealthCheck(ctx); err != nil {
		status["details"].(gin.H)["database"] = "unavailable"
		status["status"] = "degraded"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status["details"].(gin.H)["database"] = "available"
	}

	c.JSON(httpStatus, status)
}

func (h *HealthHandler) dbHealthCheck(ctx context.Context) error {
	// Простая проверка подключения к БД
	_, err := h.db.GetClientByID(ctx, "healthcheck")
	if err != nil && err != models.ErrNotFound {
		return err
	}
	return nil
}
