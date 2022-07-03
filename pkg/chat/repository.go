package chat

type Repo interface {
	CheckTopic(topic string) (ok bool, err error)
	CreateTopic(topic string) (err error)
	DeleteTopic(topic string) (err error)
	SendMessage(topic, key, text string) (err error)
	Subscribe(topic string) (consumers chan Message, errors chan error, err error)
}
