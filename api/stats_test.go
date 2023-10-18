package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	baseKafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	mockery "github.com/stretchr/testify/mock"

	"github.com/Kale-Grabovski/impay/api/mock"
	"github.com/Kale-Grabovski/impay/domain"
	"github.com/Kale-Grabovski/impay/kafka"
)

func TestStatsWalletCreatedDeleted(t *testing.T) {
	e := echo.New()

	loggerMock := &mock.LoggerMock{}
	baseConsumerMock := &mock.BaseConsumerMock{}

	statsAction := NewStatsAction(kafka.NewConsumer(baseConsumerMock, loggerMock), loggerMock)
	statsAction.chans = map[string]chan []byte{
		domain.TopicWalletCreated: make(chan []byte, 50),
	}
	baseConsumerMock.
		On("ReadMessage", domain.ConsumerTimeout).
		Maybe().
		Return(&baseKafka.Message{}, nil)
	baseConsumerMock.On("SubscribeTopics", []string{domain.TopicWalletCreated}, nil).Once().Return(nil)
	err := statsAction.InitConsumers()
	assert.NoError(t, err)
	mockery.AssertExpectationsForObjects(t, loggerMock, baseConsumerMock)

	testCases := []struct {
		req         domain.WalletMsg
		consume     func()
		logCallback func()
	}{
		{
			req: domain.WalletMsg{Amount: decimal.NewFromFloat(5.55)},
			consume: func() {
				baseConsumerMock.
					On("ReadMessage", domain.ConsumerTimeout).
					Maybe().
					Return(&baseKafka.Message{Value: nil}, nil)
			},
		},
	}

	for _, tc := range testCases {
		tc.consume()
		time.Sleep(time.Millisecond * 10)
		req := httptest.NewRequest(http.MethodGet, "/stats", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, statsAction.Get(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var resp statsWalletResp
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.True(t, resp.Total.GreaterThan(decimal.NewFromInt32(0)))
			assert.Equal(t, resp.Total, resp.Active)
			assert.True(t, resp.Inactive.Equals(decimal.NewFromInt32(0)))
			assert.True(t, resp.Deposited.Equals(decimal.NewFromInt32(0)))
			assert.True(t, resp.Withdrawn.Equals(decimal.NewFromInt32(0)))
			assert.True(t, resp.Transferred.Equals(decimal.NewFromInt32(0)))
		}
		mockery.AssertExpectationsForObjects(t, loggerMock, baseConsumerMock)
	}
}
