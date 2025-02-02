package com.twilight.aggregator.config;

import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.twilight.aggregator.model.Transaction;
import com.twilight.aggregator.serialization.TransactionDeserializer;

public class FlinkConfig extends BaseConfig {
    private static final Logger LOG = LoggerFactory.getLogger(FlinkConfig.class);
    private static volatile FlinkConfig instance;
    private String[] windowSizes;
    private String[] windowNames;

    private FlinkConfig() {
        super();
        initializeConfig();
    }

    public static synchronized FlinkConfig getInstance() {
        if (instance == null) {
            instance = new FlinkConfig();
        }
        return instance;
    }

    private void initializeConfig() {
        String windowSizesStr = getProperty("flink.window.sizes", "20s,1m,5m,30m");
        String windowNamesStr = getProperty("flink.window.names", "20s,1m,5m,30m");
        windowSizes = windowSizesStr.split(",");
        windowNames = windowNamesStr.split(",");
        validateWindowConfigs();
    }

    // Flink基本配置
    public int getParallelism() {
        return getIntProperty("flink.parallelism", "1");
    }

    public long getCheckpointInterval() {
        return getLongProperty("flink.checkpoint.interval", "60000");
    }

    // Kafka Source配置
    public String getKafkaBootstrapServers() {
        return getProperty("kafka.bootstrap.servers", "localhost:9092");
    }

    public String getKafkaTopic() {
        return getProperty("kafka.topic", "chain_transactions_new");
    }

    public String getKafkaGroupId() {
        return getProperty("kafka.group.id", "flink-aggregator");
    }

    // 窗口配置
    public String[] getWindowSizes() {
        return windowSizes;
    }

    public String[] getWindowNames() {
        return windowNames;
    }

    // 元数据配置
    public long getPairMetadataRefreshInterval() {
        return getLongProperty("pair.metadata.refresh.interval", "60000");
    }

    public String getUsdcAddress() {
        return getProperty("token.usdc.address", "0xeCC540e356b9E7c6e17fEA13c6Fe192deBefB51D");
    }

    // Kafka Source Builder
    public KafkaSource<Transaction> buildKafkaSource() {
        return KafkaSource.<Transaction>builder()
            .setBootstrapServers(getKafkaBootstrapServers())
            .setTopics(getKafkaTopic())
            .setGroupId(getKafkaGroupId())
            .setStartingOffsets(OffsetsInitializer.latest())
            .setValueOnlyDeserializer(new TransactionDeserializer())
            .build();
    }

    // 窗口配置验证
    private void validateWindowConfigs() {
        if (windowSizes.length != windowNames.length) {
            throw new RuntimeException("Window sizes and names must have the same length");
        }
        
        for (String size : windowSizes) {
            if (!size.matches("\\d+[smh]")) {
                throw new RuntimeException("Invalid window size format: " + size + 
                    ". Must be a number followed by s(seconds), m(minutes), or h(hours)");
            }
        }
    }

    // 窗口大小解析
    public Time parseWindowSize(String windowSize) {
        if (!windowSize.matches("\\d+[smh]")) {
            throw new IllegalArgumentException("Invalid window size format: " + windowSize + 
                ". Must be a number followed by s(seconds), m(minutes), or h(hours)");
        }

        String value = windowSize.substring(0, windowSize.length() - 1);
        String unit = windowSize.substring(windowSize.length() - 1);
        long size = Long.parseLong(value);

        switch (unit.toLowerCase()) {
            case "s":
                return Time.seconds(size);
            case "m":
                return Time.minutes(size);
            case "h":
                return Time.hours(size);
            default:
                throw new IllegalArgumentException("Unsupported time unit: " + unit);
        }
    }
}
