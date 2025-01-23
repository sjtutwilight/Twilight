#!/bin/bash

JAVA_OPTS="--add-opens java.base/java.util=ALL-UNNAMED \
           --add-opens java.base/java.lang=ALL-UNNAMED \
           --add-opens java.base/java.lang.invoke=ALL-UNNAMED"

mvn clean package -DskipTests

java $JAVA_OPTS -jar target/aggregator-1.0-SNAPSHOT.jar 