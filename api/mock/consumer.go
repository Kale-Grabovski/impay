package mock

import "github.com/stretchr/testify/mock"

type ConsumerMock struct {
	mock.Mock
}

func (p *ConsumerMock) Subscribe(topic string, ch chan<- []byte) error {
	return p.Called(topic, ch).Error(0)
}
