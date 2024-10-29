package server

import (
	"fmt"
	"net/http"
	"ticket-booking-backend/cmd/api/session"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) AddMiddlewares() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://anotherdomain.com"}, // Specify allowed origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},           // Allowed methods
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},           // Allowed headers
		ExposeHeaders:    []string{"Content-Length"},                                    // Expose headers
		AllowCredentials: true,                                                          // Allow cookies/auth headers
		MaxAge:           12 * time.Hour,                                                // Cache preflight response
	}))
	s.router.Use(addHeaders())
	s.router.Use(addSessionMiddleware(s.sessionManager))
	s.router.Use(gin.Recovery())
}

func addHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Use the following header to allow WebSocket connections
		ctx.Header("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws://localhost:8080;")
		ctx.Next()
	}
}

func addSessionMiddleware(sessionManager *session.SessionManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionID, err := ctx.Cookie("session_id")
		fmt.Printf("Initial cookie check - sessionID: %s, err: %v\n", sessionID, err)

		if err != nil || sessionID == "" {
			newSessionID := generateSessionID()
			fmt.Printf("Setting new sessionID: %s\n", newSessionID)

			// Set SameSite attribute and be more explicit with cookie settings
			ctx.SetSameSite(http.SameSiteLaxMode)
			ctx.SetCookie(
				"session_id",
				newSessionID,
				int(sessionManager.SessionTTL),
				"/",
				"",    // empty domain for same-origin
				false, // secure
				true,  // httpOnly
			)

			ctx.Set("session_id", newSessionID)
			sessionManager.RedisClient.Set(ctx, "session:"+sessionID, "", sessionManager.SessionTTL)
		} else {
			ctx.Set("session_id", sessionID)
			info, err := sessionManager.RedisClient.Get(ctx, "session:"+sessionID).Result()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			sessionManager.RedisClient.Set(ctx, "session:"+sessionID, info, sessionManager.SessionTTL)
		}

		ctx.Next()
	}
}

func generateSessionID() string {
	return uuid.NewString()
}
