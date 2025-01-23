package com.twilight.aggregator.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class AppConfig {
    private static final Logger LOG = LoggerFactory.getLogger(AppConfig.class);

    // Default configurations
    private static final class Defaults {
        static final String LOCAL_KAFKA = "localhost:9092";
        static final String DOCKER_KAFKA = "kafka:29092";
        static final String LOCAL_POSTGRES = "jdbc:postgresql://localhost:5432/twilight";
        static final String DOCKER_POSTGRES = "jdbc:postgresql://postgres:5432/twilight";
        static final String TOPIC = "chain_transactions";
        static final String GROUP_ID = "pair_metrics_processor";
        static final String DB_USER = "twilight";
        static final String DB_PASSWORD = "twilight123";
        static final String CHAIN_ID = "31337";
    }

    private final Environment environment;
    private final String kafkaBootstrapServers;
    private final String inputTopic;
    private final String kafkaGroupId;
    private final String jdbcUrl;
    private final String jdbcUser;
    private final String jdbcPassword;
    private final String chainId;

    public enum Environment {
        DEV,
        PROD;

        public static Environment fromString(String env) {
            if (env == null) {
                return detectEnvironment();
            }
            try {
                return Environment.valueOf(env.toUpperCase());
            } catch (IllegalArgumentException e) {
                return detectEnvironment();
            }
        }

        private static Environment detectEnvironment() {
            try {
                return java.nio.file.Files.exists(java.nio.file.Paths.get("/.dockerenv")) 
                    ? PROD : DEV;
            } catch (Exception e) {
                return DEV;
            }
        }
    }

    private AppConfig(Builder builder) {
        this.environment = builder.environment;
        this.kafkaBootstrapServers = builder.kafkaBootstrapServers;
        this.inputTopic = builder.inputTopic;
        this.kafkaGroupId = builder.kafkaGroupId;
        this.jdbcUrl = builder.jdbcUrl;
        this.jdbcUser = builder.jdbcUser;
        this.jdbcPassword = builder.jdbcPassword;
        this.chainId = builder.chainId;
        
        logConfiguration();
    }

    private void logConfiguration() {
        LOG.info("Configuration loaded for {} environment:", environment);
        LOG.info("KAFKA_BOOTSTRAP_SERVERS: {}", kafkaBootstrapServers);
        LOG.info("INPUT_TOPIC: {}", inputTopic);
        LOG.info("KAFKA_GROUP_ID: {}", kafkaGroupId);
        LOG.info("JDBC_URL: {}", jdbcUrl);
        LOG.info("JDBC_USER: {}", jdbcUser);
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private Environment environment = Environment.detectEnvironment();
        private String kafkaBootstrapServers;
        private String inputTopic;
        private String kafkaGroupId;
        private String jdbcUrl;
        private String jdbcUser;
        private String jdbcPassword;
        private String chainId = Defaults.CHAIN_ID;

        public Builder withEnvironment(String env) {
            this.environment = Environment.fromString(env);
            return this;
        }

        public Builder withKafkaBootstrapServers(String servers) {
            this.kafkaBootstrapServers = servers;
            return this;
        }

        public Builder withInputTopic(String topic) {
            this.inputTopic = topic;
            return this;
        }

        public Builder withKafkaGroupId(String groupId) {
            this.kafkaGroupId = groupId;
            return this;
        }

        public Builder withJdbcUrl(String url) {
            this.jdbcUrl = url;
            return this;
        }

        public Builder withJdbcUser(String user) {
            this.jdbcUser = user;
            return this;
        }

        public Builder withJdbcPassword(String password) {
            this.jdbcPassword = password;
            return this;
        }

        public Builder withChainId(String chainId) {
            this.chainId = chainId;
            return this;
        }

        public AppConfig build() {
            // Set defaults based on environment if not specified
            if (kafkaBootstrapServers == null) {
                kafkaBootstrapServers = getEnvOrDefault("KAFKA_BOOTSTRAP_SERVERS",
                    environment == Environment.DEV ? Defaults.LOCAL_KAFKA : Defaults.DOCKER_KAFKA);
            }
            if (inputTopic == null) {
                inputTopic = getEnvOrDefault("INPUT_TOPIC", Defaults.TOPIC);
            }
            if (kafkaGroupId == null) {
                kafkaGroupId = getEnvOrDefault("KAFKA_GROUP_ID", Defaults.GROUP_ID);
            }
            if (jdbcUrl == null) {
                jdbcUrl = getEnvOrDefault("JDBC_URL",
                    environment == Environment.DEV ? Defaults.LOCAL_POSTGRES : Defaults.DOCKER_POSTGRES);
            }
            if (jdbcUser == null) {
                jdbcUser = getEnvOrDefault("JDBC_USER", Defaults.DB_USER);
            }
            if (jdbcPassword == null) {
                jdbcPassword = getEnvOrDefault("JDBC_PASSWORD", Defaults.DB_PASSWORD);
            }

            return new AppConfig(this);
        }
    }

    private static String getEnvOrDefault(String key, String defaultValue) {
        String value = System.getenv(key);
        return value != null ? value : defaultValue;
    }

    // Getters
    public Environment getEnvironment() {
        return environment;
    }

    public String getKafkaBootstrapServers() {
        return kafkaBootstrapServers;
    }

    public String getInputTopic() {
        return inputTopic;
    }

    public String getKafkaGroupId() {
        return kafkaGroupId;
    }

    public String getJdbcUrl() {
        return jdbcUrl;
    }

    public String getJdbcUser() {
        return jdbcUser;
    }

    public String getJdbcPassword() {
        return jdbcPassword;
    }

    public String getChainId() {
        return chainId;
    }

    @Override
    public String toString() {
        return "AppConfig{" +
            "environment=" + environment +
            ", kafkaBootstrapServers='" + kafkaBootstrapServers + '\'' +
            ", inputTopic='" + inputTopic + '\'' +
            ", kafkaGroupId='" + kafkaGroupId + '\'' +
            ", jdbcUrl='" + jdbcUrl + '\'' +
            ", jdbcUser='" + jdbcUser + '\'' +
            '}';
    }
} 