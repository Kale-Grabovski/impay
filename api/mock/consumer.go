package mock

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/mock"
)

type ConsumerMock struct {
	mock.Mock
}

func (p *ConsumerMock) Subscribe(ctx context.Context, chans map[string]chan []byte) error {
	return p.Called(ctx, chans).Error(0)
}

type BaseConsumerMock struct {
	mock.Mock
}

func (p *BaseConsumerMock) SubscribeTopics(topic []string, rebalanceCb kafka.RebalanceCb) error {
	return p.Called(topic, nil).Error(0)
}
func (p *BaseConsumerMock) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	args := p.Called(timeout)
	return args.Get(0).(*kafka.Message), args.Error(1)
}
