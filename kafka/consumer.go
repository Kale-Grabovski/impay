package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/impay/domain"
)

const consumerTimeout = 50 * time.Millisecond

type Consumer struct {
	cfg    *domain.Config
	logger domain.Logger
}

func NewConsumer(cfg *domain.Config, logger domain.Logger) *Consumer {
	return &Consumer{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Consumer) Subscribe(topic string, ch chan<- []byte) error {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        s.cfg.Kafka.Host,
		"group.id":                 "test",
		"auto.offset.reset":        "latest",
		"fetch.min.bytes":          "1",
		"allow.auto.create.topics": "true",
	})
	if err != nil {
		return fmt.Errorf("cannot init consumer: %v", err)
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("cannot subscribe to topic %s: %v", topic, err)
	}

	go func() {
		for {
			msg, err := consumer.ReadMessage(consumerTimeout)
			if err != nil {
				if !err.(kafka.Error).IsTimeout() {
					s.logger.Error("cannot consume event from topic "+topic, zap.Error(err))
				}
				continue
			}

			select {
			case ch <- msg.Value:
				s.logger.Debug("consumed from "+topic, zap.ByteString("msg", msg.Value))
			default:
			}
		}
	}()
	return nil
}
