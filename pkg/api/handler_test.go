package api

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/mustafasegf/chat/internal/logger"
	"github.com/mustafasegf/chat/pkg/repository/redpanda"
	"github.com/stretchr/testify/assert"
)

func init() {
	godotenv.Load("../../.env")
	rand.Seed(time.Now().UTC().UnixNano())
	godotenv.Load()
	logger.SetLogger()
}

func TestSubscribeChat(t *testing.T) {
	brokers := []string{os.Getenv("KAFKA_HOST")}

	consumer, producer, broker, err := redpanda.NewConn(brokers)
  if err != nil {
    panic(err)
  }

	assert.NoError(t, err, expectedStr(nil, err))

	s := MakeServer(consumer, producer, broker)
	s.SetupRouter()

	t.Run("SubscribeChat", func(t *testing.T) {
    url := fmt.Sprintf("/chat/subscribe?topic=%s&message=%s", "topic", randSeq(10))
		resp, err := s.Router.Test(httptest.NewRequest("GET", url, nil))
		defer resp.Body.Close()
	  b, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(b))

		assert.NoError(t, err, expectedStr(nil, err))
		assert.Equal(t, 200, resp.StatusCode, expectedStr(200, resp.StatusCode))
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

func expectedStr(expected, got interface{}) string {
	return fmt.Sprintf("Expected %v but got %v", expected, got)
}
