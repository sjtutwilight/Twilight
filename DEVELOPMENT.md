# Development Guide

## Prerequisites

- Go 1.22 or newer
- Docker and Docker Compose
- Node.js (for local Hardhat node)
- Java 11 or newer
- Maven 3.6 or newer

## Project Structure

The project follows a microservices architecture with the following structure:

## Services

### Common (/common)
Shared code and utilities used across services:
- `/types`: Common data types and interfaces
- `/kafka`: Kafka utilities and message definitions
- `/config`: Configuration utilities

### Chain Listener (/listener)
Service responsible for monitoring blockchain events:
- `/cmd`: Service entry points
- `/internal`: Private implementation details
- `/pkg`: Public APIs and implementations

### Chain Processor (/processor)
Service responsible for processing and storing blockchain data:
- `/cmd`: Service entry points
- `/internal`: Private implementation details
- `/pkg`: Public APIs and implementations

## Dependencies
Each service has its own go.mod file and can be built independently:
- github.com/twilight/common
- github.com/twilight/listener
- github.com/twilight/processor

## Building and Running

### Building Individual Services
```bash
# Build listener
cd listener && go build ./cmd/...

# Build processor
cd processor && go build ./cmd/...
```

### Running Services
Each service can be run independently or via docker-compose.

## Getting Started

1. Start the infrastructure services:
   ```bash
   docker-compose up -d
   ```

2. Initialize the database:
   ```bash
   # TODO: Add database migration commands
   ```

3. Build and deploy Flink jobs:
   ```bash
   cd flink-jobs
   mvn clean package
   # Submit job to Flink cluster
   flink run target/twilight-flink-job-1.0-SNAPSHOT.jar
   ```

4. Start the Go services (in development):
   ```bash
   # Start the listener service
   go run cmd/listener/main.go
   
   # Start the strategy service
   go run cmd/strategy/main.go
   
   # Start the data manager service
   go run cmd/datamanager/main.go
   ```

## Development Workflow

1. Make sure to run tests before committing:
   ```bash
   # Go tests
   go test ./...
   
   # Java tests
   cd flink-jobs && mvn test
   ```

2. Follow best practices and project conventions:
   - Use the standard `net/http` package for Go services
   - Follow Flink best practices for streaming jobs
   - Implement proper error handling
   - Write tests for new functionality
   - Document public APIs

## Infrastructure Services

- Kafka: localhost:9092
- PostgreSQL: localhost:5432
- Flink JobManager UI: http://localhost:8081
- Local Hardhat Node: http://localhost:8545

## Configuration

- Go services: Configuration is managed through `configs/config.yaml`
- Flink jobs: Configuration is passed through job parameters and environment variables

## Flink Job Documentation

### Data Models

#### TransactionEvent
The core data model that represents a blockchain transaction and its associated events:
- `transaction`: Transaction details including hash, block number, from/to addresses, etc.
- `events`: List of events emitted during the transaction
- `timestamp`: Event timestamp for windowing operations

#### Event
Represents a single blockchain event:
- `eventName`: Name of the event (e.g., "Swap", "Mint", "Burn")
- `contractAddress`: Address of the contract that emitted the event
- `decodedArgs`: JSON string containing decoded event arguments
- `blockNumber`: Block number where the event occurred
- `createTime`: Event creation timestamp

### Processing Pipeline

1. **Event Ingestion**
   - Source: Kafka topic "chain_transactions"
   - Deserializer: `EventDeserializationSchema` converts JSON to `TransactionEvent`
   - Watermark Strategy: Monotonous timestamps for event-time processing

2. **Event Processing**
   - Window Configuration: 5-minute sliding windows with 1-minute slides
   - Key Selector: Group events by transaction hash
   - Window Function: `MetricsWindowFunction` processes events in each window

3. **Metrics Aggregation**
   - Swap Metrics (`SwapMetricsAggregator`):
     - Calculates total volume from Swap events
     - Tracks swap count per pair
     - Maintains window start/end times
   
   - Liquidity Metrics (`LiquidityMetricsAggregator`):
     - Tracks liquidity added/removed from Mint/Burn events
     - Counts mint and burn operations
     - Maintains window start/end times

4. **Output**
   - Sink: Kafka topic "aggregated_metrics"
   - Serializer: `TransactionEventSerializer` converts metrics to JSON
   - Delivery: At-least-once semantics

### Configuration

The Flink job uses the following configuration parameters:
- `INPUT_TOPIC`: "chain_transactions" - Source of blockchain events
- `OUTPUT_TOPIC`: "aggregated_metrics" - Destination for aggregated metrics
- `BOOTSTRAP_SERVERS`: "localhost:9092" - Kafka broker address
- `GROUP_ID`: "uniswap-metrics-group" - Consumer group ID
- Checkpoint interval: 60 seconds
- Window size: 5 minutes
- Window slide: 1 minute

### Testing

The job includes unit tests that verify:
- Swap event processing
- Mint event processing
- Window aggregation logic
- Serialization/deserialization

Test data is generated to simulate:
- Swap events with token amounts
- Mint events for liquidity provision
- Transaction metadata

### Deployment

The job is packaged as a fat JAR using the Maven shade plugin with:
- All dependencies included
- Proper resource merging
- Package relocation to avoid conflicts
- Main class: `com.twilight.jobs.UniswapMetricsJob`

Deploy using Flink CLI:
```bash
flink run target/twilight-flink-job-1.0-SNAPSHOT.jar
```

Monitor job execution through the Flink UI at http://localhost:8081 