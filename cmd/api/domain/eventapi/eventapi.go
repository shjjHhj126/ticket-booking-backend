package eventapi

import (
	"log"
	"net/http"
	"strconv"
	"ticket-booking-backend/domain/artist"
	"ticket-booking-backend/domain/event"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/dto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func CreateEventHandler(eventService *event.EventService, venueService *venue.VenueService, artistService *artist.ArtistService, validator *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var postEvent dto.PostEventDTO
		if err := ctx.ShouldBindJSON(&postEvent); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := validator.Struct(postEvent)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		existArtist, err := artistService.Exist(postEvent.ArtistID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check artist existence: " + err.Error()})
			return
		}
		if !existArtist {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "artist does not exist"})
			return
		}

		// check venue existence
		existVenue, err := venueService.Exist(postEvent.VenueID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check venue existence: " + err.Error()})
			return
		}
		if !existVenue {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "venue does not exist"})
			return
		}

		eventModel := event.Event{
			Name:        postEvent.Name,
			StartTime:   postEvent.StartTime,
			EndTime:     postEvent.EndTime,
			Status:      postEvent.Status,
			VenueID:     postEvent.VenueID,
			ArtistID:    postEvent.ArtistID,
			Description: postEvent.Description,
		}

		if err := eventService.CreateEvent(&eventModel); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Event created successfully"})
	}
}

func SetSeatsPriceHandler(eventService *event.EventService, venueService *venue.VenueService, validator *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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

		log.Printf("event_id = %d exists", eventID)

		var setSeatsPriceDTO dto.SetSeatsPriceDTO
		if err := ctx.ShouldBindJSON(&setSeatsPriceDTO); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = validator.Struct(setSeatsPriceDTO)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var seatPriceModels []event.EventSeatPrice
		for _, seatID := range setSeatsPriceDTO.SeatIDs {
			if isExists, err := venueService.SeatExist(seatID); err != nil || !isExists {
				if !isExists {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				} else {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}

			seatPriceModel := event.EventSeatPrice{
				EventID: eventID,
				SeatID:  seatID,
				Price:   setSeatsPriceDTO.Price,
			}
			seatPriceModels = append(seatPriceModels, seatPriceModel)
		}

		if err := eventService.SetEventSeatsPrice(seatPriceModels); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Seats Prices set successfully"})
	}
}
