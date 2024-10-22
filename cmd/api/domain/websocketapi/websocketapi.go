package websocketapi

import (
	"log"
	"net/http"
	"ticket-booking-backend/cmd/api/websocket"

	"github.com/gin-gonic/gin"
)

func WebsocketHandler(cm *websocket.ConnectionManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionIDAny, exists := ctx.Get("session_id")
		if !exists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required in cookie"})
			return
		}
		sessionID := sessionIDAny.(string)

		wsConn, err := cm.CreateConnection(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
			return
		}
		defer wsConn.Close()

		err = cm.AddConnection(sessionID, wsConn)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register WebSocket connection"})
			return
		}
		defer cm.RemoveConnection(sessionID)

		// Read the message sent from frontend over websocket (for testing)
		for {
			messageType, p, err := wsConn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			// Echo the received message back to the client
			if err := wsConn.WriteMessage(messageType, p); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}
}
