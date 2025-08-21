package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/tushar-kalsi/product-views/internal/repository"
)

// Consumer handles consuming and processing messages from Kafka
type Consumer struct {
	consumer *kafka.Consumer
	topic    string
	repo     repository.ProductRepository
	wg       sync.WaitGroup
	done     chan struct{}
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers, groupID, topic string, repo repository.ProductRepository) (*Consumer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":    brokers,
		"group.id":             groupID,
		"auto.offset.reset":    "earliest",
		"enable.auto.commit":   false,  // We'll commit manually after processing
		"max.poll.interval.ms": 300000, // 5 minutes
		"session.timeout.ms":   10000,  // 10 seconds
	}

	c, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer: c,
		topic:    topic,
		repo:     repo,
		done:     make(chan struct{}),
	}, nil
}

// Start begins consuming messages
func (c *Consumer) Start() error {
	if err := c.consumer.Subscribe(c.topic, nil); err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	c.wg.Add(1)
	go c.processMessages()

	return nil
}

// Stop gracefully shuts down the consumer
func (c *Consumer) Stop() {
	close(c.done)
	c.wg.Wait()
	_ = c.consumer.Close()
}

func (c *Consumer) processMessages() {
	defer c.wg.Done()

	for {
		select {
		case <-c.done:
			return
		default:
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("Consumer error: %v\n", err)
				continue
			}

			// Process the message
			if err := c.handleMessage(msg); err != nil {
				log.Printf("Error handling message: %v\n", err)
				continue
			}

			// Commit the offset after successful processing
			if _, err := c.consumer.CommitMessage(msg); err != nil {
				log.Printf("Error committing offset: %v\n", err)
			}
		}
	}
}

func (c *Consumer) handleMessage(msg *kafka.Message) error {
	var event ViewEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Update the view count in the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.repo.IncrementViewCount(ctx, event.ProductID); err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}
