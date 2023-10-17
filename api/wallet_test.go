package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/Kale-Grabovski/impay/api/mock"
	"github.com/Kale-Grabovski/impay/domain"
)

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

	testCases := []struct {
		req      string
		respCode int
	}{
		{
			req:      `{"name": ""}`,
			respCode: http.StatusOK,
		},
		{
			req:      `{"name": "sss"}`,
			respCode: http.StatusOK,
		},
		{
			req:      `{"name": 666}`,
			respCode: http.StatusBadRequest,
		},
	}

	var wallet *createWalletResp
	for i, tc := range testCases {
		req := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(tc.req))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		producerMock.
			On("Send", domain.TopicWalletCreated, int32(0), domain.WalletMsg{}).
			Once().
			Return(nil)

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
	}
	return walletAction, wallet
}
