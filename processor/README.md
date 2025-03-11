# Blockchain Event Processor

The Blockchain Event Processor is a component of the Twilight system that consumes blockchain events from Kafka and processes them into the PostgreSQL database. It handles transaction data, events, and specifically focuses on transfer events for tokens and LP tokens.

## Features

- Consumes blockchain events from Kafka
- Processes transactions and events into PostgreSQL
- Handles ERC20 token transfers
- Handles LP token transfers
- Maintains account asset balances
- Uses Redis for caching token, pair, and account metadata

## Architecture

The processor consists of several components:

1. **Kafka Consumer**: Consumes blockchain events from Kafka topics
2. **Cache Manager**: Manages Redis cache for token, pair, and account metadata
3. **Transfer Processor**: Processes transfer events for tokens and LP tokens
4. **Database Operations**: Handles database operations for inserting events and updating account assets

## Configuration

The processor can be configured using command-line flags:

- `--db`: Database connection string (default: `postgres://postgres:postgres@localhost:5432/twilight?sslmode=disable`)
- `--redis`: Redis address (default: `localhost:6379`)
- `--kafka-brokers`: Comma-separated list of Kafka brokers (default: `localhost:9092`)
- `--kafka-group`: Kafka consumer group ID (default: `twilight-processor`)
- `--kafka-topics`: Comma-separated list of Kafka topics (default: `blockchain-events`)

## Building and Running

### Prerequisites

- Go 1.22 or newer
- PostgreSQL
- Redis
- Kafka

### Building

```bash
cd processor
go build -o processor ./cmd/main.go
```

### Running

```bash
./processor --db="postgres://user:password@localhost:5432/twilight?sslmode=disable" --redis="localhost:6379" --kafka-brokers="localhost:9092" --kafka-topics="blockchain-events"
```

### Docker

You can also build and run the processor using Docker:

```bash
# Build the Docker image
docker build -t twilight/processor .

# Run the Docker container
docker run -d --name processor \
  --network host \
  twilight/processor \
  --db="postgres://user:password@postgres:5432/twilight?sslmode=disable" \
  --redis="redis:6379" \
  --kafka-brokers="kafka:9092" \
  --kafka-topics="blockchain-events"
```

## Development

### Project Structure

```
processor/
├── cmd/
│   └── main.go           # Main entry point
├── pkg/
│   ├── cache/            # Cache management
│   │   └── manager.go    # Redis cache manager
│   ├── kafka/            # Kafka consumer
│   │   └── consumer.go   # Kafka consumer implementation
│   ├── model/            # Data models
│   │   └── errors.go     # Custom error types
│   ├── processor/        # Main processor
│   │   ├── main.go       # Processor implementation
│   │   ├── statements.go # SQL prepared statements
│   │   └── transaction.go # Transaction processing
│   ├── transfer/         # Transfer event processing
│   │   ├── processor.go  # Transfer processor
│   │   ├── token_transfer.go # Token transfer handling
│   │   ├── lp_transfer.go # LP token transfer handling
│   │   └── db_operations.go # Database operations
│   └── types/            # Type definitions
│       └── transaction_data.go # Transaction data type
├── Dockerfile            # Docker build file
└── README.md             # This file
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 