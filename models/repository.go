package models

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var ErrNotFound = errors.New("record not found")

type Repository interface {
	CreateClient(ctx context.Context, client *Client) error
	GetClientByID(ctx context.Context, id string) (*Client, error)
	UpdateClient(ctx context.Context, client *Client) error
	Close() error
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository() (*PostgresRepository, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(&Client{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreateClient(ctx context.Context, client *Client) error {
	if err := r.db.WithContext(ctx).Create(client).Error; err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetClientByID(ctx context.Context, id string) (*Client, error) {
	var client Client
	if err := r.db.WithContext(ctx).First(&client, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return &client, nil
}

func (r *PostgresRepository) UpdateClient(ctx context.Context, client *Client) error {
	if err := r.db.WithContext(ctx).Save(client).Error; err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB.Close()
}
