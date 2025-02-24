package models

import (
	"encoding/json"
	"fmt"
)

type OrderEvent struct {
	UUID string
	Name string
	Meta Meta
}

type Meta struct {
	UUID            string `json:"event_uuid,omitempty"`
	OrderUUID       string `json:"uuid,omitempty"`
	OrderNumber     string `json:"number,omitempty"`
	OrderUploadedAt string `json:"uploaded_at,omitempty"`
	OrderState      string `json:"state,omitempty"`
	OrderAccountID  int64  `json:"account_id,omitempty"`
	Error           string `json:"error,omitempty"`
}

func (m *Meta) Scan(value interface{}) error {
	if value == nil {
		*m = Meta{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("transaction_outbox/models/event: meta invalid format error, expected json")
	}

	return json.Unmarshal(b, &m)
}
