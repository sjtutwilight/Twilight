package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sjtutwilight/Twilight/processor/pkg/kafka"
	"github.com/sjtutwilight/Twilight/processor/pkg/processor"
)

func main() {
	// Parse command line flags
	dbConnStr := flag.String("db", "postgres://twilight:twilight123@localhost:5432/twilight?sslmode=disable", "Database connection string")
	redisAddr := flag.String("redis", "localhost:6379", "Redis address")
	kafkaBrokers := flag.String("kafka-brokers", "localhost:9092", "Comma-separated list of Kafka brokers")
	kafkaGroupID := flag.String("kafka-group", "twilight-processor", "Kafka consumer group ID")
	kafkaTopics := flag.String("kafka-topics", "chain_transactions_new", "Comma-separated list of Kafka topics")
	flag.Parse()

	// Create logger
	logger := log.New(os.Stdout, "[processor] ", log.LstdFlags)
	logger.Println("Starting blockchain event processor")

	// Create processor
	proc, err := processor.NewProcessor(*dbConnStr, *redisAddr)
	if err != nil {
		logger.Fatalf("Failed to create processor: %v", err)
	}
	defer proc.Close()

	// Create Kafka consumer
	brokers := strings.Split(*kafkaBrokers, ",")
	topics := strings.Split(*kafkaTopics, ",")
	consumerConfig := &kafka.ConsumerConfig{
		Brokers:       brokers,
		GroupID:       *kafkaGroupID,
		Topics:        topics,
		ProcessorAddr: *dbConnStr,
		RedisAddr:     *redisAddr,
	}

	consumer, err := kafka.NewConsumer(consumerConfig, proc)
	if err != nil {
		logger.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Printf("Received signal %s, shutting down", sig)
		cancel()
	}()

	// Start consumer
	logger.Println("Starting Kafka consumer")
	if err := consumer.Start(ctx); err != nil {
		logger.Fatalf("Failed to start Kafka consumer: %v", err)
	}

	logger.Println("Processor shutdown complete")
}
