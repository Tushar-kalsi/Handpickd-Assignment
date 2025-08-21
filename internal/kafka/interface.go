package kafka

import (
	"context"
	"github.com/google/uuid"
)

// ProducerInterface defines the interface for Kafka producer
type ProducerInterface interface {
	SendViewEvent(ctx context.Context, productID uuid.UUID) error
	Close()
}
