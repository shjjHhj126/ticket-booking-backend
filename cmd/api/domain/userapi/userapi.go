package userapi

import (
	"net/http"
	"ticket-booking-backend/domain/user"
	"ticket-booking-backend/dto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func CreateUserHandler(service *user.UserService, validator *validator.Validate) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var postUser dto.PostUser
		if err := ctx.ShouldBindJSON(&postUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := user.DtoToModel(postUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := service.CreateUser(&user); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}
