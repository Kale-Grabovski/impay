package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/impay/domain"
)

const doneChanLen = 100

type statsWalletResp struct {
	Deposited   decimal.Decimal `json:"deposited"`
	Withdrawn   decimal.Decimal `json:"withdrawn"`
	Transferred decimal.Decimal `json:"transferred"`
	Total       decimal.Decimal `json:"total"`
	Active      decimal.Decimal `json:"active"`
	Inactive    decimal.Decimal `json:"inactive"`
}

type StatsAction struct {
	sync.RWMutex
	Deposited   decimal.Decimal
	Withdrawn   decimal.Decimal
	Transferred decimal.Decimal
	Total       decimal.Decimal
	Active      decimal.Decimal
	Inactive    decimal.Decimal

	logger    domain.Logger
	consumers []*kafka.Consumer
	done      chan struct{}
}

type Consumer interface {
	Subscribe(topic string, ch chan<- []byte) error
}

func NewStatsAction(
	consumerSvc Consumer,
	logger domain.Logger,
) (*StatsAction, error) {
	ret := &StatsAction{
		logger: logger,
		done:   make(chan struct{}, doneChanLen), // actually we only need 5 len (number of topics)
	}
	return ret, ret.initConsumers(consumerSvc)
}

func (s *StatsAction) Get(c echo.Context) (err error) {
	s.RLock()
	defer s.RUnlock()
	return c.JSON(http.StatusOK, statsWalletResp{
		Total:       s.Total,
		Active:      s.Active,
		Inactive:    s.Inactive,
		Deposited:   s.Deposited,
		Withdrawn:   s.Withdrawn,
		Transferred: s.Withdrawn,
	})
}

func (s *StatsAction) CloseConsumers() {
	for _, c := range s.consumers {
		if err := c.Close(); err != nil {
			s.logger.Error("cannot close consumer", zap.Error(err))
		}
	}
	for i := 0; i < doneChanLen; i++ {
		s.done <- struct{}{}
	}
	s.logger.Debug("consumers closed")
}

func (s *StatsAction) initConsumers(svc Consumer) error {
	subs := map[string]func([]byte){
		domain.TopicWalletCreated: func(m []byte) {
			s.Lock()
			s.Total = s.Total.Add(decimal.NewFromInt32(1))
			s.Active = s.Active.Add(decimal.NewFromInt32(1))
			s.Unlock()
		},
		domain.TopicWalletDeleted: func(m []byte) {
			s.Lock()
			s.Active = s.Active.Add(decimal.NewFromInt32(1))
			s.Inactive = s.Inactive.Add(decimal.NewFromInt32(-1))
			s.Unlock()
		},
		domain.TopicWalletDeposited: func(m []byte) {
			var msg domain.WalletMsg
			_ = json.Unmarshal(m, &msg)
			s.Lock()
			s.Deposited = s.Deposited.Add(msg.Amount)
			s.Unlock()
		},
		domain.TopicWalletWithdrawn: func(m []byte) {
			var msg domain.WalletMsg
			_ = json.Unmarshal(m, &msg)
			s.Lock()
			s.Withdrawn = s.Withdrawn.Add(msg.Amount)
			s.Unlock()
		},
		domain.TopicWalletTransferred: func(m []byte) {
			var msg domain.WalletMsg
			_ = json.Unmarshal(m, &msg)
			s.Lock()
			s.Transferred = s.Transferred.Add(msg.Amount)
			s.Unlock()
		},
	}

	for topic, callback := range subs {
		err := s.subscribe(svc, topic, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsAction) subscribe(svc Consumer, topic string, callback func([]byte)) error {
	ch := make(chan []byte)
	err := svc.Subscribe(topic, ch)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case m := <-ch:
				callback(m)
			case <-s.done:
				close(ch)
				return
			}
		}
	}()
	return nil
}
