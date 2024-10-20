package artistapi

import (
	"log"
	"net/http"
	"ticket-booking-backend/domain/artist"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func CreateArtistHandler(service *artist.ArtistService, validator *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var postArtist artist.Artist
		if err := ctx.ShouldBindJSON(&postArtist); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("new artist: %+v", postArtist)

		err := validator.Struct(postArtist)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := service.CreateArtist(&postArtist); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "Artist created successfully"})
	}
}
