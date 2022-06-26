package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/websocket/v2"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Service interface {
	SendMessage(topic string, message string) (res Message, status int, err error)
	ReadMessage(ctx context.Context, c *websocket.Conn, topic string)
	WriteMessage(ctx context.Context, c *websocket.Conn, topic string)
	PingClient(ctx context.Context, c *websocket.Conn, cancel context.CancelFunc)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{
		repo: repo,
	}
}

func (service *service) SendMessage(topic string, rawMessage string) (res Message, status int, err error) {
	status = http.StatusOK
	id, err := gonanoid.New()
	if err != nil {
		return
	}

	message := Message{
		Text:      rawMessage,
		CreatedAt: time.Now(),
	}

	err = service.repo.SendMessage(topic, id, message)
	res = message
	return
}

func (service *service) ReadMessage(ctx context.Context, c *websocket.Conn, topic string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			if mt == websocket.TextMessage {
				service.SendMessage(topic, string(msg))
			} else if mt == websocket.CloseMessage {
				ctx.Done()
			}
		}
	}
}

func (service *service) WriteMessage(ctx context.Context, c *websocket.Conn, topic string) {
	consumer, errors, err := service.repo.Subscribe(topic)
	go func() {
    ctx.Done()
  }()
  for {
		select {
		case <-ctx.Done():
			return
		case ans := <-consumer:
			b, _ := json.Marshal(ans)
			if err = c.WriteMessage(websocket.TextMessage, b); err != nil {
				break
			}
		case err := <-errors:
			b, _ := json.Marshal(err)
			c.WriteMessage(websocket.CloseAbnormalClosure, b)
			return
		}
	}
}

func (service *service) PingClient(ctx context.Context, c *websocket.Conn, cancle context.CancelFunc) {
}
