package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/lib/pq"

	"github.com/sjtutwilight/Twilight/brain/internal/domain/model"
	"github.com/sjtutwilight/Twilight/brain/internal/domain/orchestrator"
	"github.com/sjtutwilight/Twilight/brain/internal/infra/config"
	"github.com/sjtutwilight/Twilight/brain/internal/infra/db"
	"github.com/sjtutwilight/Twilight/brain/internal/infra/kafkaclient"
	"github.com/sjtutwilight/Twilight/brain/internal/utils"

	// 自动注册Condition和Action
	_ "github.com/sjtutwilight/Twilight/brain/internal/domain/action"
	_ "github.com/sjtutwilight/Twilight/brain/internal/domain/condition"
)

// kafkaProducerWrapper 将confluent-kafka-go的Producer包装为orchestrator.Producer接口
type kafkaProducerWrapper struct {
	p *kafka.Producer
}

func (kpw *kafkaProducerWrapper) Produce(topic string, key, value []byte) error {
	deliveryChan := make(chan kafka.Event, 1)

	err := kpw.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
	}, deliveryChan)

	if err != nil {
		return err
	}

	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return m.TopicPartition.Error
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("produce message timeout")
	}
	return nil
}

// decodeEvent用于从字节流中解析Event
func decodeEvent(data []byte) (model.Event, error) {
	var evt model.Event
	err := json.Unmarshal(data, &evt)
	return evt, err
}

func main() {
	cfg := config.LoadConfig()
	utils.Logger.Printf("Starting orchestrator service...")

	// 初始化DB连接
	dbConn, err := sql.Open("postgres", cfg.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	defer dbConn.Close()

	stratRepo := db.NewPgStrategyRepository(dbConn)

	// 初始化原生Kafka Producer
	rawProducer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.KafkaBroker})
	if err != nil {
		log.Fatalf("Fail to create producer: %v", err)
	}
	defer rawProducer.Close()
	producerWrapper := &kafkaProducerWrapper{p: rawProducer}

	// 创建消费者（使用kafkaclient包装）
	eventConsumer, err := kafkaclient.NewConsumer(cfg.KafkaBroker, "orchestrator-group", cfg.EventInputTopic)
	if err != nil {
		log.Fatalf("Failed to create event consumer: %v", err)
	}
	defer eventConsumer.Close()

	resumeConsumer, err := kafkaclient.NewConsumer(cfg.KafkaBroker, "resume-group", cfg.ResumeTopic)
	if err != nil {
		log.Fatalf("Failed to create resume consumer: %v", err)
	}
	defer resumeConsumer.Close()

	delayConsumer, err := kafkaclient.NewConsumer(cfg.KafkaBroker, "delay-group", cfg.DelayTopic)
	if err != nil {
		log.Fatalf("Failed to create delay consumer: %v", err)
	}
	defer delayConsumer.Close()

	// 加载所有策略
	strategies, err := stratRepo.GetAllStrategies()
	if err != nil {
		log.Fatalf("Failed to load strategies: %v", err)
	}
	if len(strategies) == 0 {
		log.Fatalf("No strategies found in DB")
	}

	// 为每个策略创建一个Orchestrator
	orchestrators := make([]*orchestrator.Orchestrator, len(strategies))
	for i, strategy := range strategies {
		orchestrators[i] = &orchestrator.Orchestrator{
			Strategy:    strategy,
			Producer:    producerWrapper,
			DelayTopic:  cfg.DelayTopic,
			ResumeTopic: cfg.ResumeTopic,
		}
		// 启动延迟调度
		orchestrators[i].StartDelayScheduler(delayConsumer)
	}

	// 创建批处理器
	batchProcessor := model.NewBatchProcessor(model.BatchConfig{
		BatchSize: cfg.BatchSize,
		Timeout:   cfg.BatchTimeout,
	}, func(events []model.Event) error {
		for _, evt := range events {
			for _, orch := range orchestrators {
				if err := orch.HandleEvent(evt); err != nil {
					utils.Logger.Printf("Error handling event: %v", err)
				}
			}
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动批处理器
	go batchProcessor.Start(ctx)

	go func() {
		for {
			msg, err := eventConsumer.Consume()
			if err != nil {
				utils.Logger.Printf("Error consuming events_in: %v", err)
				continue
			}
			if msg == nil {
				continue
			}
			evt, decErr := decodeEvent(msg.Value)
			if decErr != nil {
				utils.Logger.Printf("Invalid event: %v", decErr)
				continue
			}
			if err := batchProcessor.Add(evt); err != nil {
				utils.Logger.Printf("Error adding event to batch: %v", err)
			}
		}
	}()

	go func() {
		for {
			msg, err := resumeConsumer.Consume()
			if err != nil {
				utils.Logger.Printf("Error consuming resume_events: %v", err)
				continue
			}
			if msg == nil {
				continue
			}
			for _, orch := range orchestrators {
				if err := orch.ProcessResumeMessage(msg.Value); err != nil {
					utils.Logger.Printf("Error handling resume message: %v", err)
				}
			}
		}
	}()

	utils.Logger.Println("Orchestrator service started, waiting for events...")

	// 优雅关闭
	doneChan := make(chan os.Signal, 1)
	signal.Notify(doneChan, os.Interrupt, syscall.SIGTERM)
	<-doneChan

	utils.Logger.Println("Shutting down orchestrator...")
}
