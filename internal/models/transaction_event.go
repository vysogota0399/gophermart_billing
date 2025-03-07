package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/vysogota0399/gophermart_protos/utils/amount"
	"google.golang.org/genproto/googleapis/type/money"
)

type TransactionEvent struct {
	UUID string
	Name string
	Meta *TransactionEventMeta
}

type TransactionEventMeta struct {
	UUID            string    `json:"event_uuid,omitempty"`
	TransactionUUID string    `json:"transaction_uuid,omitempty"`
	OrderNumber     string    `json:"number,omitempty"`
	AccountID       int64     `json:"accountID,omitempty"`
	AmountUnits     int64     `json:"amoun_units,omitempty"`
	AmountNanos     int32     `json:"amoun_nanos,omitempty"`
	Operation       string    `json:"operation,omitempty"`
	ProcessedAt     time.Time `json:"processed_at"`
	Error           string    `json:"error,omitempty"`
}

func (m *TransactionEventMeta) Scan(value interface{}) error {
	if value == nil {
		*m = TransactionEventMeta{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("transaction_outbox/models/event: meta invalid format error, expected json")
	}

	return json.Unmarshal(b, &m)
}

func (m *TransactionEventMeta) Amount() *amount.Amount {
	return amount.New(&money.Money{Nanos: m.AmountNanos, Units: m.AmountUnits})
}
