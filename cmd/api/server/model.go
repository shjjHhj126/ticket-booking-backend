package server

import (
	"ticket-booking-backend/cmd/api/session"
	"ticket-booking-backend/cmd/api/websocket"
	"ticket-booking-backend/domain/artist"
	"ticket-booking-backend/domain/event"
	"ticket-booking-backend/domain/ticket"
	"ticket-booking-backend/domain/user"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/tool/rabbitmq"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	redislib "github.com/redis/go-redis/v9"
)

type Server struct {
	router            *gin.Engine // the core of the server
	redisClient       *redislib.Client
	mq                *rabbitmq.RabbitMQ
	db                *sql.DB
	services          Services
	validator         *validator.Validate
	sessionManager    *session.SessionManager
	ConnectionManager *websocket.ConnectionManager
}
type Services struct {
	ticketService *ticket.TicketService
	venueService  *venue.VenueService
	userService   *user.UserService
	artistService *artist.ArtistService
	eventService  *event.EventService
}
