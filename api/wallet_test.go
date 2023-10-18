package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	mockery "github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/Kale-Grabovski/impay/api/mock"
	"github.com/Kale-Grabovski/impay/domain"
)

func TestDeposit(t *testing.T) {
	e := echo.New()

	// Create one wallet
	producerMock := &mock.ProducerMock{}
	loggerMock := &mock.LoggerMock{}
	walletAction := NewWalletAction(producerMock, loggerMock)

	req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{"name": "xxx"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	producerMock.
		On("Send", domain.TopicWalletCreated, int32(0), domain.WalletMsg{}).
		Once().
		Return(nil)

	err := walletAction.Create(c)
	assert.NoError(t, err)

	var walletResp *createWalletResp
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &walletResp))

	mockery.AssertExpectationsForObjects(t, loggerMock, producerMock)

	testCases := []struct {
		id               string
		req              string
		err              string
		producerCallback func()
		logCallback      func()
		respCode         int
	}{
		{
			id:       walletResp.ID,
			req:      `{"amount": 5.55}`,
			respCode: http.StatusOK,
			producerCallback: func() {
				producerMock.
					On(
						"Send",
						domain.TopicWalletDeposited,
						int32(0),
						domain.WalletMsg{Amount: decimal.NewFromFloat(5.55)},
					).
					Once().
					Return(nil)
			},
		},
		{
			id:               "666",
			err:              "wallet not found",
			respCode:         http.StatusNotFound,
			producerCallback: func() {},
		},
		{
			id:               walletResp.ID,
			err:              "wrong amount",
			req:              `{"amount": -5.55}`,
			respCode:         http.StatusBadRequest,
			producerCallback: func() {},
		},
		{
			id:       walletResp.ID,
			respCode: http.StatusOK,
			req:      `{"amount": 5.55}`,
			producerCallback: func() {
				producerMock.
					On(
						"Send",
						domain.TopicWalletDeposited,
						int32(0),
						domain.WalletMsg{Amount: decimal.NewFromFloat(5.55)},
					).
					Once().
					Return(errors.New("publish failed"))
			},
			logCallback: func() {
				loggerMock.
					On(
						"Error",
						"cannot publish event to topic "+domain.TopicWalletDeposited,
						[]zap.Field{zap.Error(errors.New("publish failed"))},
					).
					Once().
					Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.req))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetPath("/wallets/:id/deposit")
		c.SetParamNames("id")
		c.SetParamValues(tc.id)

		tc.producerCallback()
		if tc.logCallback != nil {
			tc.logCallback()
		}

		if assert.NoError(t, walletAction.Deposit(c)) {
			assert.Equal(t, tc.respCode, rec.Code)
			var resp financeResp
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			if tc.err != "" {
				assert.Equal(t, tc.err, resp.Err)
			} else {
				assert.True(t, resp.Success)
				assert.Equal(t, resp.Amount, decimal.NewFromFloat(5.55))
			}
		}
		mockery.AssertExpectationsForObjects(t, loggerMock, producerMock)
	}

	// Withdraw
	testCases = []struct {
		id               string
		req              string
		err              string
		producerCallback func()
		logCallback      func()
		respCode         int
	}{
		{
			id:               walletResp.ID,
			respCode:         http.StatusBadRequest,
			req:              `{"amount": 666}`,
			err:              "not enough money to withdraw",
			producerCallback: func() {},
		},
		{
			id:       walletResp.ID,
			req:      `{"amount": 5.55}`,
			respCode: http.StatusOK,
			producerCallback: func() {
				producerMock.
					On(
						"Send",
						domain.TopicWalletWithdrawn,
						int32(0),
						domain.WalletMsg{Amount: decimal.NewFromFloat(5.55)},
					).
					Once().
					Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.req))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetPath("/wallets/:id/withdraw")
		c.SetParamNames("id")
		c.SetParamValues(tc.id)

		tc.producerCallback()
		if tc.logCallback != nil {
			tc.logCallback()
		}

		if assert.NoError(t, walletAction.Withdraw(c)) {
			assert.Equal(t, tc.respCode, rec.Code)
			var resp financeResp
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			if tc.err != "" {
				assert.Equal(t, tc.err, resp.Err)
			} else {
				assert.True(t, resp.Success)
				assert.Equal(t, resp.Amount, decimal.NewFromFloat(5.55))
			}
		}
		mockery.AssertExpectationsForObjects(t, loggerMock, producerMock)
	}
}

func TestDelete(t *testing.T) {
	e := echo.New()

	// Create one wallet
	producerMock := &mock.ProducerMock{}
	loggerMock := &mock.LoggerMock{}
	walletAction := NewWalletAction(producerMock, loggerMock)

	req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{"name": "xxx"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	producerMock.
		On("Send", domain.TopicWalletCreated, int32(0), domain.WalletMsg{}).
		Twice().
		Return(nil)

	err := walletAction.Create(c)
	assert.NoError(t, err)

	var walletResp *createWalletResp
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &walletResp))

	// Create second wallet
	req = httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{"name": "xxx2"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = walletAction.Create(c)
	assert.NoError(t, err)

	var walletResp2 *createWalletResp
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &walletResp2))

	testCases := []struct {
		id               string
		err              string
		producerCallback func()
		logCallback      func()
		respCode         int
	}{
		{
			id:       walletResp.ID,
			respCode: http.StatusOK,
			producerCallback: func() {
				producerMock.
					On("Send", domain.TopicWalletDeleted, int32(0), domain.WalletMsg{}).
					Once().
					Return(nil)
			},
		},
		{
			id:               "666",
			err:              "wallet not found",
			respCode:         http.StatusNotFound,
			producerCallback: func() {},
		},
		{
			id:       walletResp2.ID,
			respCode: http.StatusOK,
			producerCallback: func() {
				producerMock.
					On("Send", domain.TopicWalletDeleted, int32(0), domain.WalletMsg{}).
					Once().
					Return(errors.New("publish failed"))
			},
			logCallback: func() {
				loggerMock.
					On("Error", "cannot publish delete event to kafka", []zap.Field{zap.Error(errors.New("publish failed"))}).
					Once().
					Return(nil)
			},
		},
		{
			id:               walletResp.ID,
			respCode:         http.StatusBadRequest,
			err:              "wallet already deleted",
			producerCallback: func() {},
		},
	}

	for _, tc := range testCases {
		req = httptest.NewRequest(http.MethodDelete, "/", nil)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetPath("/wallets/:id")
		c.SetParamNames("id")
		c.SetParamValues(tc.id)

		tc.producerCallback()
		if tc.logCallback != nil {
			tc.logCallback()
		}

		if assert.NoError(t, walletAction.Delete(c)) {
			assert.Equal(t, tc.respCode, rec.Code)
			var resp updateDeleteWalletResp
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			if tc.err != "" {
				assert.Equal(t, tc.err, resp.Err)
			} else {
				assert.True(t, resp.Success)
				assert.Equal(t, tc.id, resp.ID)
			}
		}
		mockery.AssertExpectationsForObjects(t, loggerMock, producerMock)
	}
}

func TestGetAll(t *testing.T) {
	e := echo.New()

	// Create one wallet
	producerMock := &mock.ProducerMock{}
	loggerMock := &mock.LoggerMock{}
	walletAction := NewWalletAction(producerMock, loggerMock)

	req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{"name": "xxx"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	producerMock.
		On("Send", domain.TopicWalletCreated, int32(0), domain.WalletMsg{}).
		Once().
		Return(nil)

	err := walletAction.Create(c)
	assert.NoError(t, err)

	var walletResp *createWalletResp
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &walletResp))

	// Getting the list
	req = httptest.NewRequest(http.MethodGet, "/wallets", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, walletAction.GetAll(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		var resp []*domain.Wallet
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

		assert.Equal(t, 1, len(resp))
		assert.Equal(t, walletResp.ID, resp[0].ID)
		assert.Equal(t, domain.StatusActive, resp[0].Status)
		assert.Equal(t, decimal.NewFromFloat(0), resp[0].Balance)
		assert.Equal(t, walletResp.Name, resp[0].Name)
	}
}

func TestGetById(t *testing.T) {
	walletAction, walletResp := create(t)
	if walletResp == nil {
		t.Fatal("cannot create wallet")
	}

	testCases := []struct {
		id       string
		err      string
		respCode int
	}{
		{
			id:       walletResp.ID,
			respCode: http.StatusOK,
		},
		{
			id:       "666",
			err:      "wallet not found",
			respCode: http.StatusNotFound,
		},
	}

	e := echo.New()
	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/wallets/:id")
		c.SetParamNames("id")
		c.SetParamValues(tc.id)

		if assert.NoError(t, walletAction.GetById(c)) {
			assert.Equal(t, tc.respCode, rec.Code)
			var resp getWalletResp
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			if tc.err != "" {
				assert.Equal(t, tc.err, resp.Err)
			} else {
				assert.True(t, resp.Success)
				assert.Equal(t, tc.id, resp.ID)
				assert.Equal(t, domain.StatusActive, resp.Status)
				assert.Equal(t, decimal.NewFromFloat(0), resp.Balance)
				assert.Equal(t, "", resp.Name)
			}
		}
	}
}

func TestUpdate(t *testing.T) {
	walletAction, walletResp := create(t)
	if walletResp == nil {
		t.Fatal("cannot create wallet")
	}

	testCases := []struct {
		id       string
		req      string
		err      string
		respCode int
	}{
		{
			id:       walletResp.ID,
			req:      `{"name": "sss"}`,
			respCode: http.StatusOK,
		},
		{
			id:       walletResp.ID,
			req:      `{"name": ""}`,
			err:      "empty name passed",
			respCode: http.StatusBadRequest,
		},
		{
			id:       walletResp.ID,
			req:      `{"name": 666}`,
			err:      "wrong name passed",
			respCode: http.StatusBadRequest,
		},
		{
			id:       "666",
			req:      `{"name": "addf"}`,
			err:      "wallet not found",
			respCode: http.StatusNotFound,
		},
	}

	e := echo.New()
	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tc.req))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/wallets/:id")
		c.SetParamNames("id")
		c.SetParamValues(tc.id)

		if assert.NoError(t, walletAction.Update(c)) {
			assert.Equal(t, tc.respCode, rec.Code)
			var resp updateDeleteWalletResp
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			if tc.err != "" {
				assert.Equal(t, tc.err, resp.Err)
			} else {
				assert.True(t, resp.Success)
			}
		}
	}
}

func create(t *testing.T) (*WalletAction, *createWalletResp) {
	e := echo.New()
	producerMock := &mock.ProducerMock{}
	loggerMock := &mock.LoggerMock{}
	walletAction := NewWalletAction(producerMock, loggerMock)

	producerOk := func() {
		producerMock.
			On("Send", domain.TopicWalletCreated, int32(0), domain.WalletMsg{}).
			Once().
			Return(nil)
	}

	testCases := []struct {
		req              string
		respCode         int
		producerCallback func()
		logCallback      func()
	}{
		{
			req:              `{"name": ""}`,
			respCode:         http.StatusOK,
			producerCallback: producerOk,
		},
		{
			req:              `{"name": "sss"}`,
			respCode:         http.StatusOK,
			producerCallback: producerOk,
		},
		{
			req:              `{"name": 666}`,
			respCode:         http.StatusBadRequest,
			producerCallback: func() {},
		},
		{
			req:      `{"name": "abc"}`,
			respCode: http.StatusOK,
			producerCallback: func() {
				producerMock.
					On("Send", domain.TopicWalletCreated, int32(0), domain.WalletMsg{}).
					Once().
					Return(errors.New("publish failed"))
			},
			logCallback: func() {
				loggerMock.
					On("Error", "cannot publish create event to kafka", []zap.Field{zap.Error(errors.New("publish failed"))}).
					Once().
					Return(nil)
			},
		},
	}

	var wallet *createWalletResp
	for i, tc := range testCases {
		req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(tc.req))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		tc.producerCallback()
		if tc.logCallback != nil {
			tc.logCallback()
		}

		if assert.NoError(t, walletAction.Create(c)) {
			assert.Equal(t, tc.respCode, rec.Code)

			var resp *createWalletResp
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

			if rec.Code == http.StatusOK {
				assert.NotEmpty(t, resp.ID)
			} else {
				assert.Empty(t, resp.ID)
			}
			if i == 0 {
				wallet = resp
			}
		}
		mockery.AssertExpectationsForObjects(t, producerMock, loggerMock)
	}
	return walletAction, wallet
}
