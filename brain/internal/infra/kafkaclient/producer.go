package kafkaclient

import (
	"fmt"
	"time"

	ck "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sjtutwilight/Twilight/brain/internal/domain/ports"
)

type KafkaProducerWrapper struct {
	p *ck.Producer
}

func NewProducer(broker string) (ports.Producer, error) {
	p, err := ck.NewProducer(&ck.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		return nil, err
	}
	return &KafkaProducerWrapper{p: p}, nil
}

func (kpw *KafkaProducerWrapper) Produce(topic string, key, value []byte) error {
	deliveryChan := make(chan ck.Event, 1)

	err := kpw.p.Produce(&ck.Message{
		TopicPartition: ck.TopicPartition{Topic: &topic, Partition: ck.PartitionAny},
		Key:            key,
		Value:          value,
	}, deliveryChan)

	if err != nil {
		return err
	}

	select {
	case e := <-deliveryChan:
		m := e.(*ck.Message)
		if m.TopicPartition.Error != nil {
			return m.TopicPartition.Error
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("produce message timeout")
	}
	return nil
}
