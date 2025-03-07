package models

import (
	"encoding/json"
	"fmt"
)

const (
	AccrualNewState int32 = iota
	AccrualProcessingState
	AccrualFinishedState
	AccrualFailedState
)

type AccrualEvent struct {
	UUID  string            `json:"uuid"`
	Name  string            `json:"event_name"`
	State int32             `json:"event_state"`
	Meta  *AccrualEventMeta `json:"meta"`
}

type AccrualEventMeta struct {
	EventUUID   string `json:"event_uuid"`
	OrderUUID   string `json:"order_uuid"`
	OrderNumber string `json:"order_number"`
	Amount      int64  `json:"amount,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (m *AccrualEventMeta) Scan(value interface{}) error {
	if value == nil {
		*m = AccrualEventMeta{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("accruals/started/event: meta invalid format error, expected json")
	}

	return json.Unmarshal(b, &m)
}
