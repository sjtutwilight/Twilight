package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Kafka    KafkaConfig    `yaml:"kafka"`
	Postgres PostgresConfig `yaml:"postgres"`
	Chain    ChainConfig    `yaml:"chain"`
	Services ServicesConfig `yaml:"services"`
}

type KafkaConfig struct {
	Brokers []string     `yaml:"brokers"`
	Topics  TopicsConfig `yaml:"topics"`
}

type TopicsConfig struct {
	ChainTransactions string `yaml:"chain_transactions"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

type ChainConfig struct {
	RPCURL  string `yaml:"rpc_url"`
	ChainID int64  `yaml:"chain_id"`
}

type ServicesConfig struct {
	Listener    ServiceConfig `yaml:"listener"`
	Strategy    ServiceConfig `yaml:"strategy"`
	DataManager ServiceConfig `yaml:"datamanager"`
}

type ServiceConfig struct {
	Port int `yaml:"port"`
}

// DeploymentConfig represents the structure of deployment.json
type DeploymentConfig struct {
	Factory      string        `json:"factory"`
	Router       string        `json:"router"`
	InitCodeHash string        `json:"initCodeHash"`
	Tokens       []TokenConfig `json:"tokens"`
	Pairs        []PairConfig  `json:"pairs"`
}

type TokenConfig struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals string `json:"decimals"`
}

type PairConfig struct {
	Token0  string `json:"token0"`
	Token1  string `json:"token1"`
	Address string `json:"address"`
}

// LoadConfig loads configuration from the specified YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// LoadDeploymentConfig loads deployment configuration from the specified JSON file
func LoadDeploymentConfig(filename string) (*DeploymentConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading deployment config file: %w", err)
	}

	var config DeploymentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing deployment config file: %w", err)
	}

	return &config, nil
}
