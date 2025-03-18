package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/viper"
)

// AppConfig 应用配置结构体
type AppConfig struct {
	Chain    ChainConfig    `mapstructure:"chain"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Services ServicesConfig `mapstructure:"services"`
	Flink    FlinkConfig    `mapstructure:"flink"`
}

// ChainConfig 区块链配置
type ChainConfig struct {
	NodeURL         string             `mapstructure:"node_url"`
	Polling         PollingConfig      `mapstructure:"polling"`
	Confirmation    ConfirmationConfig `mapstructure:"confirmation"`
	ProcessedTxsTTL int64              `mapstructure:"processed_txs_ttl_ms"`
	Reorg           ReorgConfig        `mapstructure:"reorg"`
	ChainID         int                `mapstructure:"chain_id"`
	BlockTimeMS     int                `mapstructure:"block_time_ms"`
}

// PollingConfig 轮询配置
type PollingConfig struct {
	IntervalMS       int `mapstructure:"interval_ms"`
	MaxBlocksPerPoll int `mapstructure:"max_blocks_per_poll"`
	RetryIntervalMS  int `mapstructure:"retry_interval_ms"`
}

// ConfirmationConfig 确认配置
type ConfirmationConfig struct {
	RequiredBlocks int `mapstructure:"required_blocks"`
	TimeoutMS      int `mapstructure:"timeout_ms"`
}

// ReorgConfig 重组配置
type ReorgConfig struct {
	MaxBlocksToKeep int `mapstructure:"max_blocks_to_keep"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers  []string            `mapstructure:"brokers"`
	Topic    string              `mapstructure:"topic"`
	Topics   KafkaTopicsConfig   `mapstructure:"topics"`
	Consumer KafkaConsumerConfig `mapstructure:"consumer"`
}

// KafkaTopicsConfig Kafka主题配置
type KafkaTopicsConfig struct {
	ChainTransactions string `mapstructure:"chain_transactions"`
}

// KafkaConsumerConfig Kafka消费者配置
type KafkaConsumerConfig struct {
	PollTimeoutMS       int `mapstructure:"poll_timeout_ms"`
	MaxPollRecords      int `mapstructure:"max_poll_records"`
	SessionTimeoutMS    int `mapstructure:"session_timeout_ms"`
	HeartbeatIntervalMS int `mapstructure:"heartbeat_interval_ms"`
}

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host     string             `mapstructure:"host"`
	Port     int                `mapstructure:"port"`
	User     string             `mapstructure:"user"`
	Password string             `mapstructure:"password"`
	Database string             `mapstructure:"database"`
	SSLMode  string             `mapstructure:"sslmode"`
	Pool     PostgresPoolConfig `mapstructure:"pool"`
}

// PostgresPoolConfig PostgreSQL连接池配置
type PostgresPoolConfig struct {
	MaxConnections     int `mapstructure:"max_connections"`
	MinConnections     int `mapstructure:"min_connections"`
	MaxIdleTimeSeconds int `mapstructure:"max_idle_time_seconds"`
}

// ServicesConfig 服务配置
type ServicesConfig struct {
	Listener    ServiceConfig `mapstructure:"listener"`
	Strategy    ServiceConfig `mapstructure:"strategy"`
	DataManager ServiceConfig `mapstructure:"datamanager"`
}

// ServiceConfig 服务通用配置
type ServiceConfig struct {
	Port          int `mapstructure:"port"`
	BatchSize     int `mapstructure:"batch_size"`
	WorkerThreads int `mapstructure:"worker_threads"`
	QueueCapacity int `mapstructure:"queue_capacity"`
}

// FlinkConfig Flink配置
type FlinkConfig struct {
	JobManagerURL string           `mapstructure:"jobmanager_url"`
	Checkpoint    CheckpointConfig `mapstructure:"checkpoint"`
	Window        WindowConfig     `mapstructure:"window"`
	Watermark     WatermarkConfig  `mapstructure:"watermark"`
	Parallelism   int              `mapstructure:"parallelism"`
}

// CheckpointConfig Flink检查点配置
type CheckpointConfig struct {
	IntervalMS int `mapstructure:"interval_ms"`
	TimeoutMS  int `mapstructure:"timeout_ms"`
	MinPauseMS int `mapstructure:"min_pause_ms"`
}

// WindowConfig Flink窗口配置
type WindowConfig struct {
	SizeMS  int `mapstructure:"size_ms"`
	SlideMS int `mapstructure:"slide_ms"`
}

// WatermarkConfig Flink水印配置
type WatermarkConfig struct {
	DelayMS int `mapstructure:"delay_ms"`
}

// DeploymentConfig 部署配置结构体
type DeploymentConfig struct {
	Factory string        `json:"factory"`
	Router  string        `json:"router"`
	Tokens  []TokenConfig `json:"tokens"`
	Pairs   []PairConfig  `json:"pairs"`
}

// TokenConfig 代币配置
type TokenConfig struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals string `json:"decimals"`
	ID       string `json:"id"`
}

// PairConfig 交易对配置
type PairConfig struct {
	Token0  string `json:"token0"`
	Token1  string `json:"token1"`
	Address string `json:"address"`
}

// LoadConfig 加载应用配置
func LoadConfig() (*AppConfig, error) {
	// 设置默认值
	setDefaultConfig()

	// 读取配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Failed to read config file: %v", err)
		log.Println("Using default values")
	}

	// 读取环境变量
	viper.AutomaticEnv()

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// LoadDeploymentConfig 加载部署配置
func LoadDeploymentConfig(filePath string) (*DeploymentConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployment config file: %w", err)
	}

	var config DeploymentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment config: %w", err)
	}

	return &config, nil
}

// setDefaultConfig 设置默认配置值
func setDefaultConfig() {
	viper.SetDefault("chain.node_url", "http://127.0.0.1:8545")
	viper.SetDefault("chain.polling.interval_ms", 1000)
	viper.SetDefault("chain.polling.max_blocks_per_poll", 100)
	viper.SetDefault("chain.polling.retry_interval_ms", 5000)
	viper.SetDefault("chain.confirmation.required_blocks", 0)
	viper.SetDefault("chain.confirmation.timeout_ms", 30000)
	viper.SetDefault("chain.processed_txs_ttl_ms", 3600000)
	viper.SetDefault("chain.reorg.max_blocks_to_keep", 100)
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topic", "dex_transaction")
}
