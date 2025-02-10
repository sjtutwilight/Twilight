#!/bin/bash

# Function to check if a command was successful
check_error() {
    if [ $? -ne 0 ]; then
        echo "Error: $1 failed"
        exit 1
    fi
}

# Function to wait for a port to be available
wait_for_port() {
    local port=$1
    local service=$2
    local retries=30
    local wait=2
    
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

# Function to create a new terminal with a command
create_terminal() {
    local name=$1
    local command=$2
    # Try iTerm2 first, fall back to Terminal.app
    if osascript -e 'tell application "iTerm2" to version' &>/dev/null; then
        # Using iTerm2
        osascript <<EOF
        tell application "iTerm2"
            create window with default profile
            tell current window
                tell current session
                    write text "echo '=== $name ==='"
                    write text "cd $(pwd)"
                    write text "$command"
                end tell
            end tell
        end tell
EOF
    else
        # Using Terminal.app
        osascript <<EOF
        tell application "Terminal"
            do script "echo '=== $name ==='; cd $(pwd); $command"
            activate
        end tell
EOF
    fi
}

# Start docker-compose
echo "Starting docker services..."
docker-compose up -d
check_error "Docker Compose"

# Wait for essential services
wait_for_port 6379 "Redis"
wait_for_port 5432 "PostgreSQL"
wait_for_port 9092 "Kafka"

# Start Hardhat node in a new terminal
echo "Starting Hardhat node..."
create_terminal "node" "cd localhost && npx hardhat node"
sleep 5  # Wait for node to initialize
wait_for_port 8545 "Hardhat node"

# Deploy contracts
echo "Deploying contracts..."
create_terminal "deploy" "cd localhost && npm run deploy"
sleep 10  # Wait for deployment to complete

# Start simulation
echo "Starting simulation..."
create_terminal "simulate" "cd localhost && npm run simulate"
sleep 5  # Wait for simulation to initialize

# Start listener
echo "Starting listener..."
create_terminal "listener" "cd listener && go run cmd/main.go"
sleep 3  # Wait for listener to initialize

# Start remaining services in parallel
echo "Starting remaining services..."
create_terminal "processor" "cd processor/cmd && go run main.go"
create_terminal "aggregator" "cd aggregator && ./run.sh"
create_terminal "graphql" "cd api && go run cmd/server/main.go"

echo "All services started successfully!" 