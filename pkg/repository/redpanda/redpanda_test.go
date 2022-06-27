package redpanda

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mustafasegf/chat/internal/logger"
	"github.com/mustafasegf/chat/pkg/chat"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
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
			zap.L().Warn("Could not connect to database", zap.Error(err))
			return err
		}
		return
	}); err != nil {
		zap.L().Fatal("Could not connect to database", zap.Error(err))
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		zap.L().Fatal("Could not purge resource", zap.Error(err))
	}

	os.Exit(code)
}

func TestRepository_SendMessage(t *testing.T) {
	type fields struct {
		consumer sarama.Consumer
		producer sarama.SyncProducer
		broker   *sarama.Broker
	}
	type args struct {
		topic   string
		key     string
		message chat.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				consumer: consumer,
				producer: producer,
				broker:   broker,
			},
			args: args{
				topic: "test",
				key:   "test",
				message: chat.Message{
					Key:       "test",
					Text:      "test",
					CreatedAt: time.Now(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepository(tt.fields.consumer, tt.fields.producer, tt.fields.broker)
			zap.L().Info("Sending message", zap.String("topic", tt.args.topic), zap.String("key", tt.args.key), zap.String("message", tt.args.message.Text))
			if err := r.SendMessage(tt.args.topic, tt.args.key, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("Repository.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
