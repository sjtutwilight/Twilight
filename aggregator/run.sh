#!/bin/bash

JAVA_OPTS="--add-opens java.base/java.util=ALL-UNNAMED \
           --add-opens java.base/java.lang=ALL-UNNAMED \
           --add-opens java.base/java.lang.invoke=ALL-UNNAMED"

mvn clean package -DskipTests

# Run AggregatorJob
java $JAVA_OPTS -cp target/aggregator-1.0-SNAPSHOT.jar com.twilight.aggregator.AggregatorJob

echo "Job submitted successfully!" 