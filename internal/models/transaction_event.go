package models

import (
	"encoding/json"
	"fmt"
	"time"
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
	Amount          int64     `json:"amount,omitempty"`
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
