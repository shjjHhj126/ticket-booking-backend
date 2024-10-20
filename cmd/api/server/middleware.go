package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) AddMiddlewares() {
	s.router.Use(addSessionMiddleware())
	s.router.Use(gin.Recovery())
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

		ctx.Next()
	}
}

func generateSessionID() string {
	return uuid.NewString()
}
