package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/impay/domain"
)

type BaseConsumer interface {
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) (err error)
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
}

type Consumer struct {
	consumer BaseConsumer
	logger   domain.Logger
}

func NewConsumer(consumer BaseConsumer, logger domain.Logger) *Consumer {
	return &Consumer{
		logger:   logger,
		consumer: consumer,
	}
}

func (s *Consumer) Subscribe(topic string, ch chan []byte) error {
	err := s.consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("cannot subscribe to topic %s: %v", topic, err)
	}

	go func() {
		for {
			msg, err := s.consumer.ReadMessage(domain.ConsumerTimeout)
			if err != nil {
				if !err.(kafka.Error).IsTimeout() {
					s.logger.Error("cannot consume event from topic "+topic, zap.Error(err))
				}
				continue
			}

			select {
			case ch <- msg.Value:
			default:
			}
		}
	}()
	return nil
}
