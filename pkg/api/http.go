package api

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/mustafasegf/chat/docs"
	"github.com/mustafasegf/chat/internal/logger"
	"github.com/mustafasegf/chat/pkg/chat"
	"github.com/mustafasegf/chat/pkg/repository/redpanda"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"go.uber.org/zap"
)

type Server struct {
	Router   *fiber.App
	Consumer sarama.Consumer
	Producer sarama.SyncProducer
	Broker   *sarama.Broker
}

func MakeServer(consumer sarama.Consumer, producer sarama.SyncProducer, broker *sarama.Broker) Server {
	r := fiber.New()
	server := Server{
		Router:   r,
		Consumer: consumer,
		Producer: producer,
		Broker:   broker,
	}
	return server
}

type Handler struct {
	Service chat.Service
}

func NewHandler(service chat.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (s *Server) RunServer() {
	s.SetupSwagger()
	s.SetupRouter()

	port := os.Getenv("PORT")
	err := s.Router.Listen(":" + port)
	if err != nil {
		zap.L().Fatal("Failed to listen port "+port, zap.Error(err))
	}
}

func (s *Server) SetupSwagger() {
	docs.SwaggerInfo.Title = "Swagger API"
	docs.SwaggerInfo.Description = "Cards"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = os.Getenv("SWAGGER_HOST")
	docs.SwaggerInfo.Schemes = []string{"http"}

	s.Router.Get("/swagger/*", fiberSwagger.WrapHandler)
}

func (s *Server) SetupRouter() {
	s.Router.Use(logger.MiddleWare())
	s.Router.Use(recover.New())

	repo := redpanda.NewRepo(s.Consumer, s.Producer, s.Broker)
	service := chat.NewService(repo)
	handler := NewHandler(service)

	s.Router.Get("/chat/subscribe", websocket.New(handler.Subscribe))
}

// @Summary Subscribe to a topic
// @Description Endpoint for connecting to a topic
// @Tags /chat
// @Produce json
// @Param topic query string true "topic"
// @Router /chat/topic [get]
func (h *Handler) Subscribe(c *websocket.Conn) {
	topic := c.Query("topic")

	if topic == "" {
		b, err := json.Marshal(fiber.Map{"message": "topic can't be empty"})
		if err != nil {
			zap.L().Error("Failed to marshal error message", zap.Error(err))
		}
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(b)))
		c.Close()
		return
	}
	if err := h.Service.CheckAndCreateTopic(topic); err != nil {
		zap.L().Error("Failed to check and create topic", zap.Error(err))
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error()))
		c.Close()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go h.Service.ReadMessage(ctx, cancel, c, topic)
	go h.Service.WriteMessage(ctx, cancel, c, topic)
	go h.Service.PingClient(ctx, cancel, c)

	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		}
	}
}
