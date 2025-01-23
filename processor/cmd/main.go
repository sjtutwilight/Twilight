package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	db "github.com/twilight/processor/pkg"

	"github.com/segmentio/kafka-go"
	"github.com/twilight/common/pkg/config"
	"github.com/twilight/common/pkg/types"
)

var (
	configFile   = flag.String("config", "../../common/pkg/config/config.yaml", "Path to configuration file")
	workspaceDir = flag.String("workspace", "../..", "Path to workspace directory")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create database connection string
	dbConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
		cfg.Postgres.SSLMode,
	)

	// Initialize database processor
	processor, err := db.NewProcessor(dbConnStr)
	if err != nil {
		log.Fatalf("Failed to create database processor: %v", err)
	}
	defer processor.Close()

	// Initialize from deployment.json
	workspacePath, err := filepath.Abs(*workspaceDir)
	if err != nil {
		log.Fatalf("Failed to get absolute workspace path: %v", err)
	}

	if err := processor.InitializeFromDeployment(workspacePath); err != nil {
		log.Fatalf("Failed to initialize from deployment: %v", err)
	}
	log.Println("Successfully initialized from deployment.json")

	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topics.ChainTransactions,
		GroupID: "datamanager",
	})
	defer reader.Close()

	// Process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message: %v", err)
					continue
				}

				// Parse transaction message
				var tx struct {
					BlockNumber      int64         `json:"blockNumber"`
					BlockHash        string        `json:"blockHash"`
					Timestamp        int64         `json:"timestamp"`
					TransactionHash  string        `json:"transactionHash"`
					TransactionIndex int           `json:"transactionIndex"`
					Status           string        `json:"status"`
					GasUsed          uint64        `json:"gasUsed"`
					GasPrice         string        `json:"gasPrice"`
					Nonce            int           `json:"nonce"`
					From             string        `json:"from"`
					To               string        `json:"to"`
					Value            string        `json:"value"`
					InputData        string        `json:"inputData"`
					Events           []types.Event `json:"events"`
				}

				if err := json.Unmarshal(msg.Value, &tx); err != nil {
					log.Printf("Error parsing message: %v", err)
					continue
				}

				// Convert to database transaction
				dbTx := &types.Transaction{
					ChainID:         "31337", // Hardhat default chain ID
					TransactionHash: tx.TransactionHash,
					BlockNumber:     tx.BlockNumber,
					BlockTimestamp:  time.Unix(tx.Timestamp, 0),
					FromAddress:     tx.From,
					ToAddress:       tx.To,
					Status:          1, // Assuming success for now
					GasUsed:         tx.GasUsed,
					InputData:       tx.InputData,
				}

				// Process transaction and events
				if err := processor.ProcessTransaction(ctx, dbTx, tx.Events); err != nil {
					log.Printf("Error processing transaction: %v", err)
					continue
				}

				log.Printf("Processed transaction: %s", tx.TransactionHash)
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")

	// Close Kafka reader
	if err := reader.Close(); err != nil {
		log.Printf("Error closing Kafka reader: %v", err)
	}
}
