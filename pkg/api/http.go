package api

import (
	"os"

	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
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

	repo := redpanda.NewRepository(s.Consumer, s.Producer, s.Broker)
	service := chat.NewService(repo)
	handler := NewHandler(service)

	s.Router.Get("/chat/subscribe", handler.Subscribe)
}

// @Summary Subscribe to a topic
// @Description Endpoint for connecting to a topic
// @Tags /chat
// @Produce json
// @Param topic query string true "id"
// @Router /chat/topic [get]
func (h *Handler) Subscribe(c *fiber.Ctx) error {
	topic := c.Query("topic")

	if topic == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "topic can't be empty",
		})
	}

	message := c.Query("message")
	if message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "message can't be empty",
		})
	}

	res, status, err := h.Service.SendMessage(topic, message)
	if err != nil {
		return c.Status(status).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(res)
}
