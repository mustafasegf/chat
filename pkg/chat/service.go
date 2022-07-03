package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/websocket/v2"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/zap"
)

type Service interface {
	SendMessage(topic string, rawMessage string) (status int, err error)
	ReadMessage(ctx context.Context, cancle context.CancelFunc, c *websocket.Conn, topic string)
	WriteMessage(ctx context.Context, cancel context.CancelFunc, c *websocket.Conn, topic string)
	PingClient(ctx context.Context, cancel context.CancelFunc, c *websocket.Conn)
	CheckAndCreateTopic(topic string) (err error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{
		repo: repo,
	}
}

func (service *service) SendMessage(topic string, rawMessage string) (status int, err error) {
	status = http.StatusOK
	id, err := gonanoid.New()
	if err != nil {
		return
	}

	err = service.repo.SendMessage(topic, id, rawMessage)
	if err != nil {
		status = websocket.CloseInternalServerErr
	}
	return
}

func (service *service) ReadMessage(ctx context.Context, cancle context.CancelFunc, c *websocket.Conn, topic string) {
	defer cancle()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			mt, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.TextMessage {
				service.SendMessage(topic, string(msg))
			} else if mt == websocket.CloseMessage {
				return
			}
		}
	}
}

func (service *service) WriteMessage(ctx context.Context, cancle context.CancelFunc, c *websocket.Conn, topic string) {
	defer cancle()
	consumer, errors, err := service.repo.Subscribe(topic)
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute * 15):
			zap.L().Info("deleting topic", zap.String("topic", topic))
			if err = service.repo.DeleteTopic(topic); err != nil {
				zap.L().Error("Error deleting topic", zap.Error(err))
			}

			return
		case ans := <-consumer:
			b, _ := json.Marshal(ans)
			if err = c.WriteMessage(websocket.TextMessage, b); err != nil {
				return
			}
		case err := <-errors:
			b, _ := json.Marshal(err)
			c.WriteMessage(websocket.CloseAbnormalClosure, b)
			return
		}
	}
}

func (service *service) PingClient(ctx context.Context, cancle context.CancelFunc, c *websocket.Conn) {
	defer cancle()
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.SetWriteDeadline(time.Now().Add(time.Second * 15))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (service *service) CheckAndCreateTopic(topic string) (err error) {
	ok, err := service.repo.CheckTopic(topic)
	if !ok {
		err = service.repo.CreateTopic(topic)
		if err != nil {
			return
		}
	}
	return
}
