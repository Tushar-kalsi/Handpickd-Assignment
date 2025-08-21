package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

// ViewEvent represents a product view event
type ViewEvent struct {
	ProductID uuid.UUID `json:"product_id"`
	Timestamp int64     `json:"timestamp"`
}

// Producer handles producing messages to Kafka
type Producer struct {
	producer *kafka.Producer
	topic    string
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers, topic string) (*Producer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"message.max.bytes": 1000000, // 1MB
		"retries":           3,
		"acks":              "all",
	}

	p, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Start a goroutine to handle delivery reports
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				}
			}
		}
	}()

	return &Producer{
		producer: p,
		topic:    topic,
	}, nil
}

// SendViewEvent sends a product view event to Kafka
func (p *Producer) SendViewEvent(ctx context.Context, productID uuid.UUID) error {
	event := ViewEvent{
		ProductID: productID,
		Timestamp: time.Now().Unix(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Value: payload,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() {
	if p.producer != nil {
		p.producer.Flush(15 * 1000) // Wait up to 15 seconds for any queued messages
		p.producer.Close()
	}
}
