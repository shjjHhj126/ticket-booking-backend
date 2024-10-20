package ticketapi

import (
	"log"
	"net/http"
	"strconv"
	"ticket-booking-backend/domain/event"
	"ticket-booking-backend/domain/ticket"
	"ticket-booking-backend/domain/venue"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// use form instead of json
type TicketQuery struct {
	Number    int `form:"number" validate:"required,min=1,max=6"`
	LowPrice  int `form:"low_price" validate:"required,min=0"`
	HighPrice int `form:"high_price" validate:"required,min=0,gtfield=LowPrice"`
	Page      int `form:"page" validate:"required,min=1"`
	PageSize  int `form:"page_size" validate:"required,min=1,max=100"`
}

func GetTicketsHandler(ticketService *ticket.TicketService,
	venueService *venue.VenueService,
	eventService *event.EventService,
	validator *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		log.Println("in handler")

		// verify event exist
		eventIDStr := ctx.Param("event_id")
		eventID, err := strconv.Atoi(eventIDStr) // Convert event_id to int
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
			return
		}
		existEvent, err := eventService.Exist(eventID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check event existence: " + err.Error()})
			return
		}
		if !existEvent {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "event does not exist"})
			return
		}

		log.Println("event exists")

		// get Queries
		var query TicketQuery
		if err := ctx.ShouldBindQuery(&query); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
			return
		}

		log.Printf("got query:%+v\n", query)

		if err := validator.Struct(query); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Println("verify query")

		tickets, err := ticketService.GetTickets(ctx, eventID, query.Number, query.LowPrice, query.HighPrice, query.Page, query.PageSize, venueService, eventService)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tickets"})
			log.Fatal(err)
			return
		}
		log.Printf("Retrieved Tickets: %+v\n", tickets)
	}
}

func ReserveHandler(ticketService *ticket.TicketService, validator *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//
	}
}
