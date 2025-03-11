package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/sjtutwilight/Twilight/processor/pkg/processor"
	"github.com/sjtutwilight/Twilight/processor/pkg/types"
)

// Consumer represents a Kafka consumer for blockchain events
type Consumer struct {
	processor     *processor.Processor
	saramaConfig  *sarama.Config
	consumerGroup sarama.ConsumerGroup
	topics        []string
	groupID       string
	brokers       []string
	logger        *log.Logger
}

// ConsumerConfig holds configuration for the Kafka consumer
type ConsumerConfig struct {
	Brokers       []string
	GroupID       string
	Topics        []string
	ProcessorAddr string
	RedisAddr     string
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config *ConsumerConfig, proc *processor.Processor) (*Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	logger := log.New(os.Stdout, "[kafka-consumer] ", log.LstdFlags)

	return &Consumer{
		processor:     proc,
		saramaConfig:  saramaConfig,
		consumerGroup: consumerGroup,
		topics:        config.Topics,
		groupID:       config.GroupID,
		brokers:       config.Brokers,
		logger:        logger,
	}, nil
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) error {
	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create consumer handler
	handler := &consumerHandler{
		processor: c.processor,
		logger:    c.logger,
		ready:     make(chan bool),
	}

	// Start consuming in a goroutine
	consumeErrors := make(chan error, 1)
	go func() {
		for {
			c.logger.Println("Starting consumer group session")
			if err := c.consumerGroup.Consume(ctx, c.topics, handler); err != nil {
				consumeErrors <- err
				return
			}
			if ctx.Err() != nil {
				return
			}
			handler.ready = make(chan bool)
		}
	}()

	// Wait for consumer to be ready
	<-handler.ready
	c.logger.Println("Consumer is ready")

	select {
	case <-sigChan:
		c.logger.Println("Received shutdown signal, closing consumer")
		cancel()
	case err := <-consumeErrors:
		c.logger.Printf("Error from consumer: %v", err)
		return err
	case <-ctx.Done():
		c.logger.Println("Context cancelled, closing consumer")
	}

	// Wait for consumer to close
	if err := c.consumerGroup.Close(); err != nil {
		return fmt.Errorf("error closing consumer group: %w", err)
	}

	return nil
}

// consumerHandler implements sarama.ConsumerGroupHandler
type consumerHandler struct {
	processor *processor.Processor
	logger    *log.Logger
	ready     chan bool
}

// Setup is run at the beginning of a new session
func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		h.logger.Printf("Received message: topic=%s, partition=%d, offset=%d",
			message.Topic, message.Partition, message.Offset)

		// Process the message
		if err := h.processMessage(session.Context(), message); err != nil {
			h.logger.Printf("Error processing message: %v", err)
			// Continue processing other messages despite errors
		}

		// Mark message as processed
		session.MarkMessage(message, "")
	}
	return nil
}

// processMessage handles a single Kafka message
func (h *consumerHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	// Create a timeout context for processing
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse the message
	var txData types.TransactionData
	if err := json.Unmarshal(message.Value, &txData); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Process the transaction and its events
	if err := h.processor.ProcessTransaction(ctx, &txData.Transaction, txData.Events); err != nil {
		return fmt.Errorf("failed to process transaction: %w", err)
	}

	h.logger.Printf("Successfully processed transaction: %s", txData.Transaction.TransactionHash)
	return nil
}
