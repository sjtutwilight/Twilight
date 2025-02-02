#!/bin/bash

# Build the project
mvn clean package -DskipTests

# Submit to remote Flink cluster
flink run \
  -m jobmanager:8081 \
  -c com.twilight.aggregator.AggregatorJob \
  -p 4 \
  --detached \
  target/aggregator-1.0-SNAPSHOT.jar

echo "Job submitted successfully to remote cluster!" 