#!/bin/bash

# Script to run the Twilight Go application

# Set environment variables
export APP_ENV=${APP_ENV:-development}

# Function to display help
show_help() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  -h, --help      Show this help message"
    echo "  -s, --server    Run the server"
    echo "  -w, --worker    Run the worker"
    echo "  -l, --listener  Run the listener service"
    echo "  -p, --processor Run the processor service"
    echo "  -u, --uniquery  Run the UniQuery service"
    echo "  -a, --all       Run all services"
    echo "  -b, --build     Build the binaries"
    echo "  -c, --clean     Clean build artifacts"
}

# Parse command line arguments
if [ $# -eq 0 ]; then
    show_help
    exit 1
fi

# Process arguments
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--server)
            echo "Starting server..."
            cd "$(dirname "$0")/.." && make run-server
            ;;
        -w|--worker)
            echo "Starting worker..."
            cd "$(dirname "$0")/.." && make run-worker
            ;;
        -l|--listener)
            echo "Starting listener service..."
            cd "$(dirname "$0")/.." && make run-listener
            ;;
        -p|--processor)
            echo "Starting processor service..."
            cd "$(dirname "$0")/.." && make run-processor
            ;;
        -u|--uniquery)
            echo "Starting UniQuery service..."
            cd "$(dirname "$0")/.." && make run-uniquery
            ;;
        -a|--all)
            echo "Starting all services..."
            cd "$(dirname "$0")/.."
            make run-listener &
            make run-processor &
            make run-uniquery &
            make run-server &
            make run-worker &
            wait
            ;;
        -b|--build)
            echo "Building binaries..."
            cd "$(dirname "$0")/.." && make build
            ;;
        -c|--clean)
            echo "Cleaning build artifacts..."
            cd "$(dirname "$0")/.." && make clean
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
    shift
done 