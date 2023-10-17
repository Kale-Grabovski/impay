package domain

import "github.com/shopspring/decimal"

const (
	TopicWalletCreated     = "Wallet_Created"
	TopicWalletDeleted     = "Wallet_Deleted"
	TopicWalletDeposited   = "Wallet_Deposited"
	TopicWalletTransferred = "Wallet_Transferred"
	TopicWalletWithdrawn   = "Wallet_Withdrawn"
)

type WalletMsg struct {
	Amount decimal.Decimal `json:"amount,omitempty"`
}
