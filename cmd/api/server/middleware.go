package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) AddMiddlewares() {
	s.router.Use(addHeaders())
	s.router.Use(addSessionMiddleware())
	s.router.Use(gin.Recovery())
}

func addHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Use the following header to allow WebSocket connections
		ctx.Header("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws://localhost:8080;")
		ctx.Next()
	}
}

func addSessionMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionID, err := ctx.Cookie("session_id")

		if err != nil || sessionID == "" {
			newSessionID := generateSessionID()

			ctx.SetCookie("session_id", newSessionID, 3600, "/", "", false, true)
			ctx.Set("session_id", newSessionID)
		} else {
			ctx.Set("session_id", sessionID)
		}

		s, e := ctx.Get("session_id")
		if e == false {
			log.Fatal(e)
		}
		fmt.Printf("session_id:%s", s)

		ctx.Next()
	}
}

func generateSessionID() string {
	return uuid.NewString()
}
