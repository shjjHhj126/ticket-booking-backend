package server

import (
	"ticket-booking-backend/domain/user"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/tool/rabbitmq"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5"
	redislib "github.com/redis/go-redis/v9"
)

type Server struct {
	router      *gin.Engine // the core of the server
	redisClient *redislib.Client
	mq          *rabbitmq.RabbitMQ
	db          *sqlx.DB
	services    Services
}

type Services struct {
	// ticketService *ticket.TicketService
	venueService *venue.VenueService
	userService  *user.UserService
}
