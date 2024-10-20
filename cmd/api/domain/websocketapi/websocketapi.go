package websocketapi

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true //Todo: security
	},
}

func WebsocketHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		wsConn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
			return
		}
		defer wsConn.Close()

		for {
			// Handle incoming messages from clients if needed
			_, msg, err := wsConn.ReadMessage()
			if err != nil {
				break // Handle error or break the loop
			}
			// Process the message if necessary
			log.Println(msg)
		}
	}
}
