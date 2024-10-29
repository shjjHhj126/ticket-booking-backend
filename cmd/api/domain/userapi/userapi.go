package userapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"ticket-booking-backend/domain/user"
	"ticket-booking-backend/dto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserHandler(service *user.UserService, validator *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var postUser dto.PostUser
		if err := ctx.ShouldBindJSON(&postUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := user.DtoToModel(postUser) //hashes password
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := service.CreateUser(&user); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		sessionID, err := ctx.Cookie("session_id")
		if err != nil || sessionID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "No active session found"})
			return
		}

		err = service.RedisClient.Set(ctx, "session:"+sessionID, "authorized:"+strconv.Itoa(user.ID), service.SessionManager.SessionTTL).Err() // 30 min session expiry
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set session data"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}

func LoginHandler(service *user.UserService, validator *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var loginUser dto.LoginUser
		if err := ctx.ShouldBindJSON(&loginUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := service.GetUserByEmail(loginUser.Email)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("User with email %s not found", loginUser.Email)})
			return
		}

		err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(loginUser.Password))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Wrong password"})
			return
		}

		sessionID, err := ctx.Cookie("session_id")
		if err != nil || sessionID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "No active session found"})
			return
		}

		// Store the sessionID -> userID mapping in Redis and renew expiry
		err = service.RedisClient.Set(ctx, "session:"+sessionID, "authorized:"+strconv.Itoa(user.ID), service.SessionManager.SessionTTL).Err() // 30 min session expiry
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set session data"})
			return
		}
		info, err := service.RedisClient.Get(ctx, "session:"+sessionID).Result() // 30 min session expiry
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get session data"})
			return
		}
		fmt.Printf("info:%s", info)

		//Todo: renew session, also renew corresponding data in redis
		ctx.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
	}
}

func CheckLoginStatusHandler(service *user.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionID, err := ctx.Cookie("session_id")
		if err != nil || sessionID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"loggedIn": false, "message": "Not logged in"})
			return
		}

		redisKey := "session:" + sessionID
		userID, err := service.RedisClient.Get(ctx, redisKey).Result()
		if err != nil || !strings.HasPrefix(userID, "authorized:") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"loggedIn": false, "message": "Not logged in"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"loggedIn": true, "message": "User is logged in"})
	}
}
