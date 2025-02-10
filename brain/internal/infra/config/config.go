// internal/infra/config/config.go
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBConnectionString string
	KafkaBroker        string
	EventInputTopic    string
	ResumeTopic        string
	DelayTopic         string

	// 批处理配置
	BatchSize    int
	BatchTimeout time.Duration

	// HTTP服务器配置
	HTTPAddr string
}

func LoadConfig() *Config {
	batchSize, _ := strconv.Atoi(getEnvOrDefault("BATCH_SIZE", "100"))
	batchTimeout, _ := time.ParseDuration(getEnvOrDefault("BATCH_TIMEOUT", "5s"))

	return &Config{
		DBConnectionString: getEnvOrDefault("DB_CONNECTION_STRING", "postgres://postgres:postgres@localhost:5432/twilight?sslmode=disable"),
		KafkaBroker:        getEnvOrDefault("KAFKA_BROKER", "localhost:9092"),
		EventInputTopic:    getEnvOrDefault("EVENT_INPUT_TOPIC", "events_in"),
		ResumeTopic:        getEnvOrDefault("RESUME_TOPIC", "resume_events"),
		DelayTopic:         getEnvOrDefault("DELAY_TOPIC", "delay_events"),

		BatchSize:    batchSize,
		BatchTimeout: batchTimeout,

		HTTPAddr: getEnvOrDefault("HTTP_ADDR", ":8080"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
