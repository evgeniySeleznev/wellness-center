package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"wellness-center/models"
	"wellness-center/utils"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	db         models.Repository
	kafkaProd  utils.KafkaProducer
	redisCache utils.RedisClient
	esClient   utils.ElasticsearchClient
}

func NewClientHandler(db models.Repository, kafka utils.KafkaProducer, redis utils.RedisClient, es utils.ElasticsearchClient) *ClientHandler {
	return &ClientHandler{
		db:         db,
		kafkaProd:  kafka,
		redisCache: redis,
		esClient:   es,
	}
}

type ClientRequest struct {
	FullName           string `json:"full_name" binding:"required"`
	Phone              string `json:"phone" binding:"required"`
	Email              string `json:"email" binding:"required,email"`
	AdvertisingChannel string `json:"advertising_channel"`
	SpecialistID       uint   `json:"specialist_id"`
	MeetingPlace       string `json:"meeting_place" binding:"required"`
	Occupation         string `json:"occupation" binding:"required"`
	Gender             string `json:"gender" binding:"required"`
	Age                int    `json:"age" binding:"required"`
	ReasonForVisit     string `json:"reason_for_visit" binding:"required"`
	SpecialistNotes    string `json:"specialist_notes"`
}

func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req ClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	client := models.Client{
		FullName:           req.FullName,
		Phone:              req.Phone,
		Email:              req.Email,
		AdvertisingChannel: req.AdvertisingChannel,
		SpecialistID:       req.SpecialistID,
		MeetingPlace:       req.MeetingPlace,
		Occupation:         req.Occupation,
		Gender:             req.Gender,
		Age:                req.Age,
		ReasonForVisit:     req.ReasonForVisit,
		SpecialistNotes:    req.SpecialistNotes,
	}

	if err := h.db.CreateClient(c.Request.Context(), &client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create client: %v", err)})
		return
	}

	// Асинхронная отправка в Kafka
	go func() {
		if err := h.kafkaProd.SendToKafka("client-created", client.Email); err != nil {
			fmt.Printf("Failed to send Kafka message: %v\n", err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Client created successfully",
		"client":  client,
	})
}

func (h *ClientHandler) GetClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	client, err := h.db.GetClientByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get client: %v", err)})
		return
	}

	c.JSON(http.StatusOK, client)
}

func (h *ClientHandler) UpdateClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	var req ClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	client, err := h.db.GetClientByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get client: %v", err)})
		return
	}

	// Обновляем только разрешённые поля
	client.FullName = req.FullName
	client.Phone = req.Phone
	client.Email = req.Email
	client.AdvertisingChannel = req.AdvertisingChannel
	client.SpecialistID = req.SpecialistID
	client.MeetingPlace = req.MeetingPlace
	client.Occupation = req.Occupation
	client.Gender = req.Gender
	client.Age = req.Age
	client.ReasonForVisit = req.ReasonForVisit
	client.SpecialistNotes = req.SpecialistNotes

	if err := h.db.UpdateClient(c.Request.Context(), client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update client: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Client updated successfully",
		"client":  client,
	})
}

func (h *ClientHandler) SearchClients(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	results, err := h.esClient.Search(ctx, "clients", query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Search failed: %v", err)})
		return
	}

	c.Data(http.StatusOK, "application/json", results)
}
