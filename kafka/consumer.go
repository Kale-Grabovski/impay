package kafka

import (
	"context"
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

func (s *Consumer) Subscribe(ctx context.Context, chans map[string]chan []byte) error {
	var topics []string
	for topic := range chans {
		topics = append(topics, topic)
	}

	err := s.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("cannot subscribe to topics: %w", err)
	}

	go func() {
		for {
			msg, err := s.consumer.ReadMessage(domain.ConsumerTimeout)
			if err != nil && !err.(kafka.Error).IsTimeout() {
				s.logger.Error("cannot consume event", zap.Error(err))
			}

			select {
			case <-ctx.Done():
				return
			default:
				if msg != nil {
					topic := *msg.TopicPartition.Topic
					s.logger.Debug("msg consumed from topic: "+topic, zap.Error(err))
					chans[topic] <- msg.Value
				}
			}
		}
	}()
	return nil
}
