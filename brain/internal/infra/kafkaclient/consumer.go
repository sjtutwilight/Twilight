package kafkaclient

import (
	ck "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sjtutwilight/Twilight/brain/internal/domain/ports"
	"github.com/sjtutwilight/Twilight/brain/internal/utils"
)

type KafkaConsumerWrapper struct {
	c *ck.Consumer
}

func NewConsumer(broker, group, topic string) (*KafkaConsumerWrapper, error) {
	c, err := ck.NewConsumer(&ck.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          group,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumerWrapper{c: c}, nil
}

func (kcw *KafkaConsumerWrapper) Consume() (*ports.Message, error) {
	ev := kcw.c.Poll(100)
	if ev == nil {
		return nil, nil
	}
	switch e := ev.(type) {
	case *ck.Message:
		return &ports.Message{
			Key:   e.Key,
			Value: e.Value,
		}, nil
	case ck.Error:
		return nil, e
	default:
		return nil, nil
	}
}

func (kcw *KafkaConsumerWrapper) Close() {
	err := kcw.c.Close()
	if err != nil {
		utils.Logger.Printf("Error closing consumer: %v", err)
	}
}
