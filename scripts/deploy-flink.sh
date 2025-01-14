#!/bin/bash

# Create necessary directories
mkdir -p flink-jobs flink-checkpoints

# Build Flink job jar
echo "Building Flink job jar..."
mvn clean package -DskipTests

# Copy the jar to flink-jobs directory
echo "Copying jar to flink-jobs directory..."
cp target/twilight-flink-job.jar flink-jobs/

# Start or restart Flink services
echo "Starting/restarting Flink services..."
docker-compose up -d jobmanager taskmanager job-deployer

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 30

echo "Deployment complete. Check the job status at http://localhost:8081" 