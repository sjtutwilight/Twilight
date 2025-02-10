package com.twilight.aggregator.config;

public class FlinkConfig extends BaseConfig {

    private static class SingletonHelper {
        private static final FlinkConfig INSTANCE = new FlinkConfig();
    }

    private FlinkConfig() {
        super();
    }

    public static FlinkConfig getInstance() {
        return SingletonHelper.INSTANCE;
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

    // 元数据配置
    public long getPairMetadataRefreshInterval() {
        return getLongProperty("pair.metadata.refresh.interval", "60000");
    }

    public String getUsdcAddress() {
        return getProperty("token.usdc.address", "0x74A6379d012ce53E3b0718C05dD72a3De87F0c6a");
    }
}
