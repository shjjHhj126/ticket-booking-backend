package server

import (
	"net/http"
	"ticket-booking-backend/cmd/api/session"
	"ticket-booking-backend/domain/artist"
	"ticket-booking-backend/domain/event"
	"ticket-booking-backend/domain/ticket"
	"ticket-booking-backend/domain/user"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/tool/rabbitmq"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	redislib "github.com/redis/go-redis/v9"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins, adjust this for better security
	},
}

type Server struct {
	router         *gin.Engine // the core of the server
	redisClient    *redislib.Client
	mq             *rabbitmq.RabbitMQ
	db             *sql.DB
	services       Services
	validator      *validator.Validate
	sessionManager *session.SessionManager
}
type Services struct {
	ticketService *ticket.TicketService
	venueService  *venue.VenueService
	userService   *user.UserService
	artistService *artist.ArtistService
	eventService  *event.EventService
}
