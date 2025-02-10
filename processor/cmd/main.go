package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sjtutwilight/Twilight/common/pkg/config"
	"github.com/sjtutwilight/Twilight/common/pkg/types"
	db "github.com/sjtutwilight/Twilight/processor/pkg"

	"github.com/segmentio/kafka-go"
)

var (
	configFile   = flag.String("config", "../../common/pkg/config/config.yaml", "Path to configuration file")
	workspaceDir = flag.String("workspace", "..", "Path to workspace directory")
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

	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Kafka.Brokers,
		Topic:       cfg.Kafka.Topics.ChainTransactions,
		StartOffset: kafka.LastOffset,
		GroupID:     "datamanager",
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

				// Log raw message for debugging

				// Parse transaction message
				var chainEvent struct {
					Transaction struct {
						BlockNumber       int64  `json:"blockNumber"`
						BlockHash         string `json:"blockHash"`
						Timestamp         int64  `json:"timestamp"`
						TransactionHash   string `json:"transactionHash"`
						TransactionIndex  int    `json:"transactionIndex"`
						TransactionStatus string `json:"transactionStatus"`
						GasUsed           int64  `json:"gasUsed"`
						GasPrice          string `json:"gasPrice"`
						Nonce             uint64 `json:"nonce"`
						FromAddress       string `json:"fromAddress"`
						ToAddress         string `json:"toAddress"`
						TransactionValue  string `json:"transactionValue"`
						InputData         string `json:"inputData"`
						ChainID           string `json:"chainID"`
					} `json:"transaction"`
					Events []struct {
						EventName       string                 `json:"eventName"`
						ContractAddress string                 `json:"contractAddress"`
						LogIndex        int                    `json:"logIndex"`
						BlockNumber     int64                  `json:"blockNumber"`
						Topics          []string               `json:"topics"`
						EventData       string                 `json:"eventData"`
						DecodedArgs     map[string]interface{} `json:"decodedArgs"`
					} `json:"events"`
				}

				if err := json.Unmarshal(msg.Value, &chainEvent); err != nil {
					log.Printf("Error parsing message: %v", err)
					continue
				}

				// Convert to database transaction
				dbTx := &types.Transaction{
					ChainID:         chainEvent.Transaction.ChainID,
					TransactionHash: chainEvent.Transaction.TransactionHash,
					BlockNumber:     chainEvent.Transaction.BlockNumber,
					BlockTimestamp:  time.Unix(chainEvent.Transaction.Timestamp, 0),
					FromAddress:     chainEvent.Transaction.FromAddress,
					ToAddress:       chainEvent.Transaction.ToAddress,
					Status:          1, // Assuming success for now
					GasUsed:         chainEvent.Transaction.GasUsed,
					InputData:       chainEvent.Transaction.InputData,
				}

				// Convert events
				events := make([]types.Event, len(chainEvent.Events))
				for i, e := range chainEvent.Events {
					events[i] = types.Event{
						ChainID:         chainEvent.Transaction.ChainID,
						EventName:       e.EventName,
						ContractAddress: e.ContractAddress,
						LogIndex:        e.LogIndex,
						EventData:       e.EventData,
						BlockNumber:     e.BlockNumber,
						DecodedArgs:     e.DecodedArgs,
					}
				}

				// Process transaction and events
				if err := processor.ProcessTransaction(ctx, dbTx, events); err != nil {
					log.Printf("Error processing transaction: %v", err)
					continue
				}

				log.Printf("Processed transaction: %s", chainEvent.Transaction.TransactionHash)
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan

	// Close Kafka reader
	if err := reader.Close(); err != nil {
		log.Printf("Error closing Kafka reader: %v", err)
	}
}
