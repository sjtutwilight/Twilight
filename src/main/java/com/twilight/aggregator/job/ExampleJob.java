package com.twilight.aggregator.job;

import com.twilight.aggregator.config.AppConfig;
import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.api.common.serialization.SimpleStringSchema;

public class ExampleJob extends BaseJob {
    private final AppConfig config;

    public ExampleJob() {
        super();
        this.config = AppConfig.getInstance();
        logger.info("Initializing Example Job with config: {}", config);
    }

    @Override
    public void execute() throws Exception {
        // Create Kafka source
        KafkaSource<String> source = KafkaSource.<String>builder()
                .setBootstrapServers(config.getKafkaBootstrapServers())
                .setTopics(config.getInputTopic())
                .setGroupId(config.getKafkaGroupId())
                .setStartingOffsets(OffsetsInitializer.earliest())
                .setValueOnlyDeserializer(new SimpleStringSchema())
                .build();

        // Add source to environment
        env.fromSource(source, WatermarkStrategy.noWatermarks(), "Kafka Source")
           .print(); // Just print the messages for now

        // Execute job
        env.execute("Example Flink Job");
    }

    public static void main(String[] args) throws Exception {
        ExampleJob job = new ExampleJob();
        job.execute();
    }
} 