package models

import "bytes"

type File struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	ContentType string `json:"content_type" db:"content_type"`
	Public      bool   `json:"public" db:"public"`
	SenderID    string `json:"sender_id" db:"sender_id"`
	RecipientID string `json:"recipient_id,omitempty" db:"recipient_id,omitempty"`
	Size        int64  `json:"size" db:"size"`
	Data        bytes.Buffer
}
