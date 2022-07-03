package redpanda

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mustafasegf/chat/internal/logger"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	consumer sarama.Consumer
	producer sarama.SyncProducer
	broker   *sarama.Broker
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	logger.SetLogger()
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		zap.L().Fatal("Could not connect to docker", zap.Error(err))
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "vectorized/redpanda",
		Tag:        "v22.1.4",
		Hostname:   "redpanda",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8081/tcp":  {{HostPort: "8081/tcp"}},
			"8082/tcp":  {{HostPort: "8082/tcp"}},
			"9092/tcp":  {{HostPort: "9092/tcp"}},
			"28082/tcp": {{HostPort: "28082/tcp"}},
			"29092/tcp": {{HostPort: "29092/tcp"}},
		},
		ExposedPorts: []string{"8081/tcp", "8082/tcp", "9092/tcp", "28082/tcp", "29092/tcp"},
		Cmd: []string{
			"redpanda",
			"start",
			"--smp",
			"1",
			"--reserve-memory",
			"0M",
			"--overprovisioned",
			"--node-id",
			"0",
			"--kafka-addr",
			"PLAINTEXT://0.0.0.0:29092,OUTSIDE://0.0.0.0:9092",
			"--advertise-kafka-addr",
			"PLAINTEXT://redpanda:29092,OUTSIDE://localhost:9092",
			"--pandaproxy-addr",
			"PLAINTEXT://0.0.0.0:28082,OUTSIDE://0.0.0.0:8082",
			"--advertise-pandaproxy-addr",
			"PLAINTEXT://redpanda:28082,OUTSIDE://localhost:8082",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		zap.L().Fatal("Could not start resource", zap.Error(err))
	}

	port := resource.GetPort("9092/tcp")
	zap.L().Info("Connecting to database on: " + port)

	resource.Expire(120)
	pool.MaxWait = 120 * time.Second

	if err := pool.Retry(func() (err error) {
		consumer, producer, broker, err = NewConn([]string{"localhost:" + port})
		if err != nil {
			return err
		}
		return
	}); err != nil {
		zap.L().Fatal("Could not connect to database", zap.Error(err))
	}

	code := m.Run()

	// time.Sleep(time.Second * 10)
	if err := pool.Purge(resource); err != nil {
		zap.L().Fatal("Could not purge resource", zap.Error(err))
	}

	os.Exit(code)
}

func TestTopic(t *testing.T) {
	repo := NewRepo(consumer, producer, broker)
	t.Run("check no topic", func(t *testing.T) {
		topic := randSeq(10)
		ok, err := repo.CheckTopic(topic)
		assert.False(t, ok, expStr(false, ok))
		assert.NoError(t, err, expStr("<err>", err))
	})

	t.Run("check create", func(t *testing.T) {
		topic := randSeq(10)
		err := repo.CreateTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))

		ok, err := repo.CheckTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))
		assert.True(t, ok, expStr(true, ok))
	})

	t.Run("check delete", func(t *testing.T) {
		topic := randSeq(10)
		err := repo.CreateTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))

		ok, err := repo.CheckTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))
		assert.True(t, ok, expStr(true, ok))

		err = repo.DeleteTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))

		ok, err = repo.CheckTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))
		assert.False(t, ok, expStr(false, ok))
	})
}

func TestMessage(t *testing.T) {
	repo := NewRepo(consumer, producer, broker)
	t.Run("send message", func(t *testing.T) {
		topic := randSeq(10)
		err := repo.CreateTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))

		ok, err := repo.CheckTopic(topic)
		assert.NoError(t, err, expStr("<err>", err))
		assert.True(t, ok, expStr(true, ok))

		key := randSeq(6)
		err = repo.SendMessage(topic, key, "test")
		assert.NoError(t, err, expStr("<err>", err))
	})
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func expStr(exp, got interface{}) string {
	return fmt.Sprintf("Expected %v but got %v", exp, got)
}
