#!/bin/bash

# Set error handling
set -e

# Function to check if a command was successful
check_error() {
    if [ $? -ne 0 ]; then
        echo "Error: $1 failed"
        exit 1
    fi
}

# Function to check and kill process using a port
kill_port_process() {
    local port=$1
    local pid=$(lsof -ti:$port)
    if [ ! -z "$pid" ]; then
        echo "Port $port is in use by process $pid. Killing it..."
        kill $pid 2>/dev/null || kill -9 $pid 2>/dev/null
        sleep 2
    fi
}

# Function to wait for a port to be available
wait_for_port() {
    local port=$1
    local service=$2
    local retries=60  # Increased retries
    local wait=3      # Increased wait time
    
    echo "Waiting for $service to be ready on port $port..."
    while ! nc -z localhost $port && [ $retries -gt 0 ]; do
        retries=$((retries-1))
        echo "Retrying $service connection... ($retries attempts left)"
        sleep $wait
    done
    
    if [ $retries -eq 0 ]; then
        echo "Error: $service did not start properly"
        exit 1
    fi
    echo "$service is ready!"
}

# Store the original directory
ORIGINAL_DIR=$(pwd)

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    kill $(jobs -p) 2>/dev/null || true
    # Kill processes on specific ports
    kill_port_process 8545  # Hardhat
    kill_port_process 6379  # Redis
    kill_port_process 5432  # PostgreSQL
    kill_port_process 9092  # Kafka
    cd "$ORIGINAL_DIR"
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

# Clean up any existing processes
echo "Cleaning up existing processes..."
cleanup

# Start services one by one
echo "Starting docker services..."
docker-compose up -d
check_error "Docker Compose"

# Wait for essential services with increased timeout
echo "Waiting for essential services..."
sleep 10
wait_for_port 6379 "Redis"
wait_for_port 5432 "PostgreSQL"
wait_for_port 9092 "Kafka"

# Start Hardhat node with increased wait time
echo "Starting Hardhat node..."
kill_port_process 8545  # Ensure port 8545 is free
cd "$ORIGINAL_DIR/localhost" && npx hardhat node > hardhat.log 2>&1 &
sleep 15  # Increased wait time for Hardhat
wait_for_port 8545 "Hardhat node"

# Deploy contracts
echo "Deploying contracts..."
cd "$ORIGINAL_DIR/localhost" && npm run deploy > deploy.log 2>&1 &
sleep 10

# Start simulation
echo "Starting simulation..."
cd "$ORIGINAL_DIR/localhost" && npm run simulate > simulate.log 2>&1 &
sleep 8

# Start listener
echo "Starting listener..."
cd "$ORIGINAL_DIR/listener" && go run cmd/main.go > listener.log 2>&1 &
sleep 8

# Start processor
echo "Starting processor..."
cd "$ORIGINAL_DIR/processor/cmd" && go run main.go > processor.log 2>&1 &
sleep 8

# Start aggregator
echo "Starting aggregator..."
cd "$ORIGINAL_DIR/aggregator" && ./run.sh > aggregator.log 2>&1 &
sleep 8

# Start GraphQL
echo "Starting GraphQL server..."
cd "$ORIGINAL_DIR/api" && go run cmd/server/main.go > graphql.log 2>&1 &

# Print status message
echo "All services have been started!"
echo "Log files are being written to *.log in respective directories"
echo "Use 'tail -f */**.log' to monitor all logs"
echo "Press Ctrl+C to stop all services"

# Split the window into panes
# Main layout:
# +----------------+----------------+
# |    Services    |    Hardhat    |
# +----------------+----------------+
# |    Deploy     |    Simulate    |
# +----------------+----------------+
# |   Listener    |   Processor    |
# +----------------+----------------+
# |  Aggregator   |    GraphQL     |
# +----------------+----------------+

# Create the layout
tmux split-window -h
tmux split-window -v
tmux select-pane -t 0
tmux split-window -v
tmux select-pane -t 0
tmux split-window -v
tmux select-pane -t 2
tmux split-window -v
tmux select-pane -t 4
tmux split-window -v

# Start services in each pane
# Pane 0: Docker services
tmux send-keys -t 0 'echo "=== Services ===" && docker-compose up' C-m

# Wait for essential services
sleep 5
wait_for_port 6379 "Redis"
wait_for_port 5432 "PostgreSQL"
wait_for_port 9092 "Kafka"

# Pane 1: Hardhat node
tmux send-keys -t 1 'echo "=== Hardhat Node ===" && cd localnode && npx hardhat node' C-m
sleep 5
wait_for_port 8545 "Hardhat node"

# Pane 2: Deploy contracts
tmux send-keys -t 2 'echo "=== Deploy Contracts ===" && cd localnode && npm run deploy' C-m
sleep 2

# Pane 3: Simulation
tmux send-keys -t 3 'echo "=== Simulation ===" && cd localnode && npm run simulate' C-m
sleep 2

# Pane 4: Listener
tmux send-keys -t 4 'echo "=== Listener ===" && cd listener/cmd && go run main.go' C-m
sleep 2

# Pane 5: Processor
tmux send-keys -t 5 'echo "=== Processor ===" && cd processor/cmd && go run main.go' C-m
sleep 2

# Pane 6: Aggregator
tmux send-keys -t 6 'echo "=== Aggregator ===" && cd aggregator && ./run.sh' C-m
sleep 2

# Pane 7: GraphQL
tmux send-keys -t 7 'echo "=== GraphQL ===" && cd api && go run cmd/server/main.go' C-m

# Attach to the tmux session
tmux attach-session -t twilight 