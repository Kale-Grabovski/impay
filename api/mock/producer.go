package mock

import "github.com/stretchr/testify/mock"

type ProducerMock struct {
	mock.Mock
}

func (p *ProducerMock) Send(topic string, partition int32, msg any) error {
	return p.Called(topic, partition, msg).Error(0)
}
