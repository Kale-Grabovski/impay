package api

import (
	"errors"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/impay/domain"
)

const walletLen = 8

type walletReq struct {
	Name string `json:"name"`
}

type createWalletResp struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Err     string `json:"err_code,omitempty"`
	Success bool   `json:"success"`
}

type updateDeleteWalletResp struct {
	ID      string `json:"id"`
	Err     string `json:"err_code,omitempty"`
	Success bool   `json:"success"`
}

type getWalletResp struct {
	Balance decimal.Decimal `json:"balance"`
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Status  string          `json:"status"`
	Err     string          `json:"err_code,omitempty"`
	Success bool            `json:"success"`
}

type financeReq struct {
	Amount     decimal.Decimal `json:"amount"`
	TransferTo string          `json:"transfer_to"`
}

type financeResp struct {
	Amount  decimal.Decimal `json:"amount"`
	Err     string          `json:"err_code,omitempty"`
	Success bool            `json:"success"`
}

type Producer interface {
	Send(topic string, partition int32, msg any) error
}

type WalletAction struct {
	sync.RWMutex
	wallets  map[string]*domain.Wallet
	producer Producer
	logger   domain.Logger
}

func NewWalletAction(
	producer Producer,
	logger domain.Logger,
) *WalletAction {
	return &WalletAction{
		wallets:  make(map[string]*domain.Wallet),
		producer: producer,
		logger:   logger,
	}
}

func (s *WalletAction) GetAll(c echo.Context) (err error) {
	s.RLock()
	defer s.RUnlock()
	wallets := make([]*domain.Wallet, 0, len(s.wallets))
	for _, w := range s.wallets {
		wallets = append(wallets, w)
	}
	return c.JSON(http.StatusOK, wallets)
}

func (s *WalletAction) GetById(c echo.Context) (err error) {
	s.RLock()
	defer s.RUnlock()

	if wallet, ok := s.wallets[c.Param("id")]; ok {
		return c.JSON(http.StatusOK, getWalletResp{
			Balance: wallet.Balance,
			ID:      wallet.ID,
			Name:    wallet.Name,
			Status:  wallet.Status,
			Success: true,
		})
	}

	return c.JSON(http.StatusNotFound, getWalletResp{
		Err: "wallet not found",
	})
}

func (s *WalletAction) Create(c echo.Context) (err error) {
	req := &walletReq{}
	if err = c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, createWalletResp{
			Err: "wrong name passed" + err.Error(),
		})
	}

	wallet, err := s.genWallet(req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, createWalletResp{
			Err: err.Error(),
		})
	}

	s.Lock()
	s.wallets[wallet.ID] = wallet
	s.Unlock()

	msg := domain.WalletMsg{}
	err = s.producer.Send(domain.TopicWalletCreated, 0, msg)
	if err != nil {
		s.logger.Error("cannot publish create event to kafka", zap.Error(err))
	}

	return c.JSON(http.StatusOK, createWalletResp{
		ID:      wallet.ID,
		Name:    wallet.Name,
		Status:  wallet.Status,
		Success: true,
	})
}

func (s *WalletAction) Update(c echo.Context) (err error) {
	req := &walletReq{}
	if err = c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, updateDeleteWalletResp{
			Err: "wrong name passed",
		})
	}
	if utf8.RuneCountInString(req.Name) == 0 {
		return c.JSON(http.StatusBadRequest, updateDeleteWalletResp{
			Err: "empty name passed",
		})
	}

	s.Lock()
	defer s.Unlock()

	if wallet, ok := s.wallets[c.Param("id")]; ok {
		wallet.Name = req.Name
		return c.JSON(http.StatusOK, updateDeleteWalletResp{
			ID:      wallet.ID,
			Success: true,
		})
	}

	return c.JSON(http.StatusNotFound, updateDeleteWalletResp{
		Err: "wallet not found",
	})
}

func (s *WalletAction) Delete(c echo.Context) (err error) {
	s.Lock()
	if wallet, ok := s.wallets[c.Param("id")]; ok {
		wallet.Status = domain.StatusInactive
		s.Unlock()
		return c.JSON(http.StatusOK, updateDeleteWalletResp{
			ID:      wallet.ID,
			Success: true,
		})
	}
	s.Unlock()

	msg := domain.WalletMsg{}
	err = s.producer.Send(domain.TopicWalletDeleted, 0, msg)
	if err != nil {
		s.logger.Error("cannot publish delete event to kafka", zap.Error(err))
	}

	return c.JSON(http.StatusNotFound, updateDeleteWalletResp{
		Err: "wallet not found",
	})
}

func (s *WalletAction) Deposit(c echo.Context) (err error) {
	return s.financeProcess(c, func(req *financeReq, wallet *domain.Wallet) error {
		s.Lock()
		wallet.Balance = wallet.Balance.Add(req.Amount)
		s.Unlock()

		s.financePublish(domain.TopicWalletDeposited, req.Amount)
		return c.JSON(http.StatusOK, financeResp{
			Success: true,
			Amount:  req.Amount,
		})
	})
}

func (s *WalletAction) Withdraw(c echo.Context) (err error) {
	return s.financeProcess(c, func(req *financeReq, wallet *domain.Wallet) error {
		s.Lock()
		if wallet.Balance.LessThan(req.Amount) {
			s.Unlock()
			return c.JSON(http.StatusBadRequest, financeResp{
				Err: "not enough money to withdraw",
			})
		}
		wallet.Balance = wallet.Balance.Add(req.Amount.Neg())
		s.Unlock()

		s.financePublish(domain.TopicWalletWithdrawn, req.Amount)
		return c.JSON(http.StatusOK, financeResp{
			Success: true,
			Amount:  req.Amount,
		})
	})
}

func (s *WalletAction) Transfer(c echo.Context) (err error) {
	return s.financeProcess(c, func(req *financeReq, wallet *domain.Wallet) error {
		s.Lock()
		if wallet.Balance.LessThan(req.Amount) {
			s.Unlock()
			return c.JSON(http.StatusBadRequest, financeResp{
				Err: "not enough money to transfer",
			})
		}
		walletTo, ok2 := s.wallets[req.TransferTo]
		if !ok2 {
			s.Unlock()
			return c.JSON(http.StatusNotFound, financeResp{
				Err: "target wallet not found",
			})
		}
		wallet.Balance = wallet.Balance.Add(req.Amount.Neg())
		walletTo.Balance = walletTo.Balance.Add(req.Amount)
		s.Unlock()

		s.financePublish(domain.TopicWalletTransferred, req.Amount)
		return c.JSON(http.StatusOK, financeResp{
			Success: true,
			Amount:  req.Amount,
		})
	})
}

func (s *WalletAction) financePublish(topic string, amount decimal.Decimal) {
	msg := domain.WalletMsg{
		Amount: amount,
	}
	err := s.producer.Send(topic, 0, msg)
	if err != nil {
		s.logger.Error("cannot publish event to topic "+topic, zap.Error(err))
	}
}

func (s *WalletAction) financeProcess(
	c echo.Context,
	callback func(*financeReq, *domain.Wallet) error,
) error {
	req := &financeReq{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, createWalletResp{
			Err: "wrong input params",
		})
	}
	if req.Amount.LessThan(decimal.NewFromInt32(0)) {
		return c.JSON(http.StatusBadRequest, financeResp{
			Err: "wrong amount",
		})
	}

	s.RLock()
	wallet, ok := s.wallets[c.Param("id")]
	s.RUnlock()

	if ok {
		return callback(req, wallet)
	}
	return c.JSON(http.StatusNotFound, financeResp{
		Err: "wallet not found",
	})
}

func (s *WalletAction) genWallet(name string) (wallet *domain.Wallet, err error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	letterRunes := []rune("123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	walletID := make([]rune, walletLen)

	s.RLock()
	defer s.RUnlock()

	timeout := time.NewTimer(20 * time.Millisecond)
	for {
		select {
		case <-timeout.C:
			return nil, errors.New("cannot generate wallet ID")
		default:
			for i := 0; i < walletLen; i++ {
				walletID[i] = letterRunes[rnd.Intn(len(letterRunes))]
			}
			id := string(walletID)
			if _, ok := s.wallets[id]; !ok {
				timeout.Stop()
				return domain.NewWallet(id, name), nil
			}
		}
	}
}
