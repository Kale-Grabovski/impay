package mock

import (
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

type LoggerMock struct {
	mock.Mock
}

func (p *LoggerMock) Debug(msg string, fields ...zapcore.Field) {
	p.Called(msg, fields)
}

func (p *LoggerMock) Info(msg string, fields ...zapcore.Field) {
	p.Called(msg, fields)
}

func (p *LoggerMock) Warn(msg string, fields ...zapcore.Field) {
	p.Called(msg, fields)
}

func (p *LoggerMock) Error(msg string, fields ...zapcore.Field) {
	p.Called(msg, fields)
}

func (p *LoggerMock) Panic(msg string, fields ...zapcore.Field) {
	p.Called(msg, fields)
}
