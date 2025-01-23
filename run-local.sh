#!/bin/bash

# Build the project
mvn clean package

# Run the JAR file directly
java \
  -Dlog4j.configurationFile=src/main/resources/logback.xml \
  -jar target/aggregator-1.0-SNAPSHOT-shaded.jar 