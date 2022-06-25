package chat

import (
	"time"
)

type Message struct {
	CreatedAt time.Time `json:"created_at"`
  Text      string    `json:"text"`
  User      string    `json:"user"`
}
