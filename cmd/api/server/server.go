package server

import (
	"log"
	"ticket-booking-backend/cmd/api/domain/artistapi"
	"ticket-booking-backend/cmd/api/domain/eventapi"
	"ticket-booking-backend/cmd/api/domain/ticketapi"
	"ticket-booking-backend/cmd/api/domain/userapi"
	"ticket-booking-backend/cmd/api/domain/venueapi"
	"ticket-booking-backend/cmd/api/domain/websocketapi"
	"ticket-booking-backend/cmd/api/session"
	"ticket-booking-backend/cmd/api/websocket"
	"ticket-booking-backend/domain/artist"
	"ticket-booking-backend/domain/event"
	"ticket-booking-backend/domain/ticket"
	"ticket-booking-backend/domain/user"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/tool/rabbitmq"
	"ticket-booking-backend/tool/redis"
	"ticket-booking-backend/tool/sqldb"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func NewServer() *Server {
	redisClient := redis.InitRedis()
	return &Server{
		router:            gin.Default(),
		redisClient:       redisClient,
		db:                sqldb.InitPostgres(),
		mq:                rabbitmq.InitRabbitMQ(),
		sessionManager:    session.NewSessionManager(redisClient, time.Minute*30),
		validator:         validator.New(),
		ConnectionManager: websocket.NewConnectionManager(redisClient),
	}
}

func (s *Server) InitServices() {
	s.services = Services{
		ticketService: ticket.NewTicketService(s.redisClient, s.mq, s.db),
		venueService:  venue.NewVenueService(s.db),
		userService:   user.NewUserService(s.db),
		artistService: artist.NewArtistService(s.db),
		eventService:  event.NewEventService(s.db),
	}
}

func (s *Server) SetupRoutes() {
	s.router.POST("/events/:event_id/seats/set-price", eventapi.SetSeatsPriceHandler(s.services.eventService, s.services.venueService, s.validator))
	s.router.GET("/events/:event_id/tickets", ticketapi.GetTicketsHandler(s.services.ticketService, s.services.venueService, s.services.eventService, s.validator))
	s.router.POST("/events/:event_id/tickets/reserve", ticketapi.ReserveHandler(s.services.ticketService, s.services.eventService, s.validator))
	s.router.POST("/venues", venueapi.CreateVenueHandler(s.services.venueService, s.validator))
	s.router.POST("/artists", artistapi.CreateArtistHandler(s.services.artistService, s.validator))
	s.router.POST("/events", eventapi.CreateEventHandler(s.services.eventService, s.services.venueService, s.services.artistService, s.validator))
	s.router.POST("/users", userapi.CreateUserHandler(s.services.userService, s.validator))
	s.router.GET("/ws", websocketapi.WebsocketHandler(s.ConnectionManager)) // get notification: tickets unavailable/available, ticket reserved
}

func (s *Server) Run(port string) error {
	return s.router.Run(port)
}

func (s *Server) Close() {
	if err := s.redisClient.Close(); err != nil {
		log.Println("Error closing Redis client:", err)
	}

	if err := s.db.Close(); err != nil {
		log.Println("Error closing database:", err)
	}

	if err := s.mq.Close(); err != nil {
		log.Println("Error closing RabbitMQ connection:", err)
	}

	if err := s.sessionManager.Close(); err != nil {
		log.Println("Error closing Session Manager:", err)
	}

}
