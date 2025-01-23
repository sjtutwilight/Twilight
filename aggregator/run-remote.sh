#!/bin/bash

# Build the project
mvn clean package

# Submit to remote Flink cluster
flink run \
  -m flink-jobmanager:8081 \
  --class com.twilight.aggregator.job.PairMetricsJob \
  -p 4 \
  --detached \
  target/aggregator-1.0-SNAPSHOT-shaded.jar 