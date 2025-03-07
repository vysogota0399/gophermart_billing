package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/gophermart_protos/utils/amount"
)

type Transaction struct {
	UUID        string         `json:"uuid"`
	Amount      *amount.Amount `json:"amount"`
	Operation   string         `json:"operation"`
	OrderNumber string         `json:"order_number"`
	AccountID   int64          `json:"accountID"`
	CreatedAt   time.Time      `json:"creataed_at"`
	ProcessedAt time.Time      `json:"processed_at"`
}

const (
	Debit  = "debit"
	Credit = "credit"
)

func NewDebit(amount *amount.Amount, accountID int64, orderNumber string) *Transaction {
	return &Transaction{
		UUID:        uuid.NewV4().String(),
		Amount:      amount,
		AccountID:   accountID,
		Operation:   Debit,
		OrderNumber: orderNumber,
		ProcessedAt: time.Now().Local(),
	}
}

func NewCredit(amount *amount.Amount, accountID int64, orderNumber string) *Transaction {
	return &Transaction{
		UUID:        uuid.NewV4().String(),
		Amount:      amount,
		AccountID:   accountID,
		Operation:   Credit,
		OrderNumber: orderNumber,
		ProcessedAt: time.Now().Local(),
	}
}
