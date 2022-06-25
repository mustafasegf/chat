package redpanda

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mustafasegf/chat/pkg/chat"
)

func NewConn(brokers []string) (consumer sarama.Consumer, producer sarama.SyncProducer, broker *sarama.Broker, err error) {
	config := sarama.NewConfig()
	config.ClientID = "chat-client"
	config.Consumer.Return.Errors = true
	config.Producer.Return.Successes = true

	consumer, err = sarama.NewConsumer(brokers, config)
	if err != nil {
		return
	}
	producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return
	}

	broker = sarama.NewBroker(brokers[0])
	err = broker.Open(config)
	if err != nil {
		return
	}

	ok, err := broker.Connected()
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("broker is not connected")
	}

	return
}

type repo struct {
	consumer sarama.Consumer
	producer sarama.SyncProducer
	broker   *sarama.Broker
}

func NewRepository(consumer sarama.Consumer, producer sarama.SyncProducer, broker *sarama.Broker) chat.Repo {
	return &repo{
		consumer: consumer,
		producer: producer,
		broker:   broker,
	}
}

func (repo *repo) SendMessage(topic string, key string, message chat.Message) (err error) {
	ok, err := repo.CheckTopic(topic)
	if !ok {
		err = repo.CreateTopic(topic)
		if err != nil {
			return
		}
	}

	if err != nil {
		return
	}

	producerMessage := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.StringEncoder(message.Text),
	}

	_, _, err = repo.producer.SendMessage(producerMessage)
	return
}

func (repo *repo) CheckTopic(topic string) (bool, error) {
	topics, err := repo.consumer.Topics()
	if err != nil {
		return false, err
	}
	for _, t := range topics {
		if t == topic {
			return true, nil
		}
	}
	return false, nil
}

func (repo *repo) CreateTopic(topic string) (err error) {
	detail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	req := &sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: detail,
		},
		Timeout: 10 * time.Second,
	}

	fmt.Printf("Creating topic %s\n\n", topic)
	fmt.Printf("repo %#v\n\n", repo)
	fmt.Printf("req %#v\n\n", req)

	repo.broker.CreateTopics(req)
	return
}
