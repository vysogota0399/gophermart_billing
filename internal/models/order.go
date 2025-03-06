package models

import "time"

type Order struct {
	UUID       string    `json:"uuid"`
	Number     string    `json:"number"`
	State      int32     `json:"state"`
	AccountID  int64     `json:"account_id"`
	CreatedAt  time.Time `json:"created_at"`
	UploadedAt time.Time `json:"uploaded_at"`
}
