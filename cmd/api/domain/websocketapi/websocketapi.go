package websocketapi

import (
	"encoding/json"
	"log"
	"net/http"
	"ticket-booking-backend/cmd/api/websocket"
	"ticket-booking-backend/dto"

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

		// Read the message sent from frontend over websocket
		for {
			messageType, message, err := wsConn.ReadMessage() //messageType:json
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			var msgData map[string]string
			if err := json.Unmarshal(message, &msgData); err == nil {
				if action, ok := msgData["action"]; ok && action == "reconnect" {
					log.Printf("Reconnect with reservation ID: %s\n", msgData["reservationId"])
					payload, err := cm.GetReservationData(ctx, msgData["reservationId"])
					if err != nil {
						log.Printf("Error parsing reservation data for ID %s: %v\n", msgData["reservationId"], err)
						return
					}

					if err := wsConn.WriteMessage(messageType, websocket.GetWebSocketMessageBytes(dto.TypeReservation, payload)); err != nil {
						log.Println("Write error:", err)
						return
					}
				}
			}

		}
	}
}
