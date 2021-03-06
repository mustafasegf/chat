package redpanda

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mustafasegf/chat/pkg/chat"
)

var BrokerNotConnectedErr = fmt.Errorf("broker is not connected")

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
		err = BrokerNotConnectedErr
	}
	return
}

type repo struct {
	consumer sarama.Consumer
	producer sarama.SyncProducer
	broker   *sarama.Broker
}

func NewRepo(consumer sarama.Consumer, producer sarama.SyncProducer, broker *sarama.Broker) chat.Repo {
	return &repo{
		consumer: consumer,
		producer: producer,
		broker:   broker,
	}
}

func (repo *repo) CreateTopic(topic string) (err error) {
	req := &sarama.CreateTopicsRequest{
		TopicDetails: map[string]*sarama.TopicDetail{
			topic: {
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		},
		Timeout: 10 * time.Second,
	}

	_, err = repo.broker.CreateTopics(req)
	return
}

func (repo *repo) CheckTopic(topic string) (ok bool, err error) {
	ok = false
	var topics []string

	topics, err = repo.consumer.Topics()
	if err != nil {
		return
	}
	for _, t := range topics {
		if t == topic {
			ok = true
			return
		}
	}
	return
}

func (repo *repo) DeleteTopic(topic string) (err error) {
	req := &sarama.DeleteTopicsRequest{
		Topics:  []string{topic},
		Timeout: 10 * time.Second,
	}
	_, err = repo.broker.DeleteTopics(req)
	return
}

func (repo *repo) SendMessage(topic, key, text string) (err error) {
	producerMessage := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.StringEncoder(text),
	}

	_, _, err = repo.producer.SendMessage(producerMessage)
	return
}

func (repo *repo) Subscribe(topic string) (consumers chan chat.Message, errors chan error, err error) {
	consumers = make(chan chat.Message)
	errors = make(chan error)

	partitions, _ := repo.consumer.Partitions(topic)
	consumer, err := repo.consumer.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
	if err != nil {
		return
	}

	go func(topic string, consumer sarama.PartitionConsumer) {
		for {
			select {
			case msg := <-consumer.Messages():
				var chatMsg chat.Message
				json.Unmarshal(msg.Value, &chatMsg)
				chatMsg.Key = string(msg.Key)
				chatMsg.CreatedAt = msg.Timestamp
				consumers <- chatMsg

			case err := <-consumer.Errors():
				errors <- err
			}
		}
	}(topic, consumer)

	return
}
