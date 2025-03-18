package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Producer handles Kafka message publishing
type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	// Enable idempotent producer to prevent duplicate messages
	// This requires Kafka >= 0.11.0.0
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1 // Required for idempotent producer

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	log.Printf("Created Kafka producer with idempotence enabled for topic %s", topic)
	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

// SendMessage sends a message to Kafka
func (p *Producer) SendMessage(key string, value interface{}) error {
	// Convert value to JSON
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value to JSON: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(jsonValue),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Message sent to partition %d at offset %d with key %s", partition, offset, key)
	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.producer.Close()
}
