package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"twilight-go/internal/config"
	"twilight-go/internal/contracts"
	"twilight-go/internal/listener"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

const (
	deploymentPath = "deployment.json"
)

func main() {
	// 加载应用配置
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load application config: %v", err)
	}

	// 加载部署配置
	deployment, err := config.LoadDeploymentConfig(deploymentPath)
	if err != nil {
		log.Printf("Warning: Failed to load deployment config: %v", err)
		log.Println("Will use config from config.yaml")
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建共享注册表
	registry := listener.NewRegistry()

	// 创建监听器
	nodeURL := appConfig.Chain.NodeURL
	if nodeURL == "" {
		nodeURL = "http://127.0.0.1:8545" // 默认本地节点URL
	}

	kafkaBrokers := appConfig.Kafka.Brokers
	if len(kafkaBrokers) == 0 {
		kafkaBrokers = []string{"localhost:9092"} // 默认Kafka代理
	}

	topic := appConfig.Kafka.Topic
	if topic == "" {
		topic = "dex_transaction" // 默认主题
	}

	// 创建监听器
	listenerInstance, err := listener.CreateLocalNodeListener(nodeURL, kafkaBrokers, topic, registry)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	// 创建合约管理器
	contractManager := contracts.NewContractManager()

	// 加载合约
	if deployment != nil {
		// 从deployment.json加载合约
		contractManager.LoadContracts(deployment)
	} else {
		// 从config.yaml加载默认合约
		pairAddress := common.HexToAddress(viper.GetString("contracts.pair"))
		contractManager.LoadDefaultContract(pairAddress)
	}

	// 将合约注册到监听器
	contractManager.RegisterContractsWithListener(listenerInstance)

	// 启动监听器
	if err := listenerInstance.Start(ctx); err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	log.Println("Listener started successfully")

	// 等待终止信号
	<-sigChan
	log.Println("Received termination signal, shutting down...")

	// 停止监听器
	if err := listenerInstance.Stop(); err != nil {
		log.Printf("Error stopping listener: %v", err)
	}

	log.Println("Listener stopped successfully")
}
