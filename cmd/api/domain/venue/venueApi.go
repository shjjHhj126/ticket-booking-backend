package venueApi

import (
	"fmt"
	"net/http"

	"ticket-booking-backend/domain/venue"

	"github.com/gin-gonic/gin"
)

func CreateVenueHandler(service *venue.VenueService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var v venue.Venue
		if err := ctx.ShouldBindJSON(&v); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Printf("New Venue Structure: %+v\n", v)

		if err := service.CreateVenue(&v); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Venue created successfully"})
	}
}
