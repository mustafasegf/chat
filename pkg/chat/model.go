package chat

import (
	"time"
)

type Message struct {
	CreatedAt time.Time `json:"created_at"`
	Text      string    `json:"text"`
	Key       string    `json:"key"`
	User      string    `json:"user"`
}
