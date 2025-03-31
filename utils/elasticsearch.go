package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchClient interface {
	Search(ctx context.Context, index string, query string) ([]byte, error)
	Close() error
}

type elasticsearchClient struct {
	client *elasticsearch.Client
}

func NewElasticsearchClient() (ElasticsearchClient, error) {
	host := os.Getenv("ELASTICSEARCH_HOST")
	if host == "" {
		host = "http://localhost:9200"
	}

	cfg := elasticsearch.Config{
		Addresses: []string{host},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("Elasticsearch ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch ping error: %s", res.Status())
	}

	return &elasticsearchClient{client: client}, nil
}

func (e *elasticsearchClient) Search(ctx context.Context, index string, query string) ([]byte, error) {
	if e.client == nil {
		return nil, errors.New("Elasticsearch client is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(index),
		e.client.Search.WithQuery(query),
	)
	if err != nil {
		return nil, fmt.Errorf("Elasticsearch search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch search error: %s", res.Status())
	}

	return io.ReadAll(res.Body)
}

func (e *elasticsearchClient) Close() error {
	// Elasticsearch client не требует явного закрытия соединения,
	// но метод оставлен для совместимости с интерфейсом
	return nil
}
