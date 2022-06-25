package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mustafasegf/chat/internal/logger"
	"github.com/mustafasegf/chat/pkg/api"
	"github.com/mustafasegf/chat/pkg/repository/redpanda"
	"go.uber.org/zap"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	godotenv.Load()
	logger.SetLogger()
	brokers := []string{os.Getenv("KAFKA_HOST")}
	consumer, producer, broker, err := redpanda.NewConn(brokers)
	if err != nil {
		zap.L().Fatal("Failed to connect to Kafka", zap.Error(err))
	}
	s := api.MakeServer(consumer, producer, broker)
	s.RunServer()
}
