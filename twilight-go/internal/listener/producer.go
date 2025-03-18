package listener

import (
	"fmt"
	"log"

	"github.com/sjtutwilight/twilight-go/pkg/kafka"
)

// Producer interface defines methods for producing messages
type Producer interface {
	Produce(key string, data interface{}) error
	Close() error
}

// KafkaProducer implements Producer using Kafka
type KafkaProducer struct {
	producer *kafka.Producer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	producer, err := kafka.NewProducer(brokers, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Printf("Created Kafka producer for topic %s", topic)
	return &KafkaProducer{
		producer: producer,
	}, nil
}

// Produce sends a message to Kafka
func (p *KafkaProducer) Produce(key string, data interface{}) error {
	return p.producer.SendMessage(key, data)
}

// Close closes the producer
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}
