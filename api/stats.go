package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"

	"github.com/Kale-Grabovski/impay/domain"
)

const doneChanLen = 50

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

	logger      domain.Logger
	consumerSvc Consumer
	ctx         context.Context
	cancel      context.CancelFunc
	chans       map[string]chan []byte
}

type Consumer interface {
	Subscribe(ctx context.Context, topic string, ch chan []byte) error
}

func NewStatsAction(
	consumerSvc Consumer,
	logger domain.Logger,
) *StatsAction {
	ctx, cancel := context.WithCancel(context.Background())
	return &StatsAction{
		logger:      logger,
		consumerSvc: consumerSvc,
		ctx:         ctx,
		cancel:      cancel,
		chans: map[string]chan []byte{
			domain.TopicWalletCreated:     make(chan []byte, doneChanLen),
			domain.TopicWalletDeleted:     make(chan []byte, doneChanLen),
			domain.TopicWalletTransferred: make(chan []byte, doneChanLen),
			domain.TopicWalletDeposited:   make(chan []byte, doneChanLen),
			domain.TopicWalletWithdrawn:   make(chan []byte, doneChanLen),
		},
	}
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
	s.cancel()
	time.Sleep(domain.ConsumerTimeout + 10*time.Millisecond)
}

func (s *StatsAction) InitConsumers() error {
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

	for topic := range s.chans {
		err := s.subscribe(topic, subs[topic])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsAction) subscribe(topic string, callback func([]byte)) error {
	err := s.consumerSvc.Subscribe(s.ctx, topic, s.chans[topic])
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case m := <-s.chans[topic]:
				callback(m)
			case <-s.ctx.Done():
				close(s.chans[topic])
				return
			}
		}
	}()
	return nil
}
