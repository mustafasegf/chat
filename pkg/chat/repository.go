package chat

type Repo interface {
	SendMessage(topic string, key string, message Message) (err error)
  CheckTopic(topic string) (ok bool, err error)
  CreateTopic(topic string) (err error)
}
