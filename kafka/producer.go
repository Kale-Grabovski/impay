package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/Kale-Grabovski/impay/domain"
)

const producerFlushMs = 100

type Producer struct {
	cfg      *domain.Config
	producer *kafka.Producer
	logger   domain.Logger
}

func NewProducer(cfg *domain.Config, logger domain.Logger) (*Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      cfg.Kafka.Host,
		"queue.buffering.max.ms": 5,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot connect to producer: %v", err)
	}
	return &Producer{
		cfg:      cfg,
		producer: producer,
		logger:   logger,
	}, nil
}

func (s *Producer) Send(topic string, partition int32, msg any) error {
	if partition == 0 {
		partition = kafka.PartitionAny
	}

	m, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("cannot marshal msg: %v", err)
	}

	err = s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: partition},
		Value:          m,
	}, nil)
	if err != nil {
		return fmt.Errorf("cannot producer a message: %v", err)
	}
	return nil
}

func (s *Producer) Close() {
	s.producer.Flush(producerFlushMs)
	s.producer.Close()
	s.logger.Debug("producer closed")
}
