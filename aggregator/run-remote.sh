#!/bin/bash

# Build the project
mvn clean package

# 运行指定的job或所有job
if [ "$RUN_ALL" = true ]; then
  # 运行所有job
  java $JAVA_OPTS -cp target/aggregator-1.0-SNAPSHOT.jar \
    com.twilight.aggregator.job.PairMetricsJob &
  java $JAVA_OPTS -cp target/aggregator-1.0-SNAPSHOT.jar \
    com.twilight.aggregator.job.TokenMetricsJob &
else
  # 运行指定的job
  java $JAVA_OPTS -cp target/aggregator-1.0-SNAPSHOT.jar $JOB_CLASS 