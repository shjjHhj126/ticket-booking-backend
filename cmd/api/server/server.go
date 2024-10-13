package server

import (
	userApi "ticket-booking-backend/cmd/api/domain/user"
	venueApi "ticket-booking-backend/cmd/api/domain/venue"
	"ticket-booking-backend/domain/user"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/tool/rabbitmq"
	"ticket-booking-backend/tool/redis"
	"ticket-booking-backend/tool/sqldb"

	"github.com/gin-gonic/gin"
)

func NewServer() *Server {
	return &Server{
		router:      gin.Default(),
		redisClient: redis.InitRedis(),
		db:          sqldb.InitPostgres(),
		mq:          rabbitmq.InitRabbitMQ(),
	}
}

func (s *Server) InitServices() {
	s.services = Services{
		// ticketService: ticket.NewTicketService(s.redisClient, s.mq, s.db),
		venueService: venue.NewVenueService(s.db),
		// userService:  user.NewUserService(s.db),
	}
}

func (s *Server) SetupRoutes() {
	// s.router.POST("/book", ticket.InitiateBookingHandler(s.services.ticketService))
	s.router.POST("/venues", venueApi.CreateVenueHandler(s.services.venueService))
	// s.router.POST("/users", userApi.CreateUserHandler(s.services.userService))
}

func (s *Server) Run(port string) {
	s.router.Run(port)
}
