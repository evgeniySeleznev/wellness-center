package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer interface {
	SendToKafka(topic string, message string) error
	Close() error
}

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer() (KafkaProducer, error) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:   []string{broker},
		Balancer:  &kafka.LeastBytes{},
		BatchSize: 1,
	})

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	return &kafkaProducer{writer: writer}, nil
}

func (k *kafkaProducer) SendToKafka(topic string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: []byte(message),
	})
}

func (k *kafkaProducer) Close() error {
	return k.writer.Close()
}
