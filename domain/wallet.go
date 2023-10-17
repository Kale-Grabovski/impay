package domain

import "github.com/shopspring/decimal"

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Wallet struct {
	Balance decimal.Decimal `json:"balance"`
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Status  string          `json:"status"`
}

func NewWallet(id, name string) *Wallet {
	return &Wallet{
		ID:     id,
		Name:   name,
		Status: StatusActive,
	}
}
