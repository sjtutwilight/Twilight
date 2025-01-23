#!/bin/bash

# Build the project
mvn clean package

# Set Flink home directory
FLINK_HOME=/usr/local/Cellar/apache-flink/1.14.4/libexec

# Submit to remote Flink cluster
$FLINK_HOME/bin/flink run \
  -m flink-jobmanager:8081 \
  --class com.twilight.aggregator.job.PairMetricsJob \
  -p 4 \
  --detached \
  target/aggregator-1.0-SNAPSHOT-shaded.jar 