package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kafkaBase "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/impay/domain"
)

type BaseConsumer interface {
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) (err error)
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
}

type Consumer struct {
	cfg    *domain.Config
	logger domain.Logger
}

func NewConsumer(cfg *domain.Config, logger domain.Logger) *Consumer {
	return &Consumer{
		logger: logger,
		cfg:    cfg,
	}
}

func (s *Consumer) Subscribe(ctx context.Context, topic string, ch chan []byte) error {
	consumer, err := kafkaBase.NewConsumer(&kafkaBase.ConfigMap{
		"bootstrap.servers":        s.cfg.Kafka.Host,
		"group.id":                 "test",
		"auto.offset.reset":        "latest",
		"fetch.min.bytes":          "1",
		"allow.auto.create.topics": "true",
	})
	if err != nil {
		return err
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("cannot subscribe to topics: %w", err)
	}

	go func() {
		for {
			msg, err := consumer.ReadMessage(domain.ConsumerTimeout)
			if err != nil && !err.(kafka.Error).IsTimeout() {
				s.logger.Error("cannot consume event", zap.Error(err))
			}

			select {
			case <-ctx.Done():
				s.logger.Debug("closing consumer " + topic)

				err = consumer.Close()
				if err != nil {
					s.logger.Error("cannot close consumer: "+topic, zap.Error(err))
				}
				return
			default:
				if msg != nil {
					s.logger.Debug("msg consumed from topic: "+topic, zap.Error(err))
					ch <- msg.Value
				}
			}
		}
	}()
	return nil
}
