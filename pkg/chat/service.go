package chat

import (
	"net/http"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Service interface {
	GetAllMesages(topic string) ([]Message, error)
	GetChannelMessages(topic string) (chan Message, error)
	SendMessage(topic string, message string) (res Message, status int, err error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{
		repo: repo,
	}
}

func (service *service) GetAllMesages(topic string) ([]Message, error) {
	panic("not implemented") // TODO: Implement
}

func (service *service) GetChannelMessages(topic string) (chan Message, error) {
	panic("not implemented") // TODO: Implement
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
