package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"ticket-booking-backend/dto"

	// "ticket-booking-backend/domain/ticket"

	websocketlib "github.com/gorilla/websocket"
	redislib "github.com/redis/go-redis/v9"
)

type ConnectionManager struct {
	redisClient     *redislib.Client
	upgrader        *websocketlib.Upgrader
	activeConns     map[string]*websocketlib.Conn // Map
	activeConnsLock sync.RWMutex                  // Protects concurrent access to the map
}

func NewConnectionManager(redisClient *redislib.Client) *ConnectionManager {
	return &ConnectionManager{
		redisClient: redisClient,
		upgrader: &websocketlib.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Todo: Add security checks here
			},
		},
		activeConns: make(map[string]*websocketlib.Conn),
	}
}

func (cm *ConnectionManager) AddConnection(sessionID string, wsConn *websocketlib.Conn) error {
	cm.activeConnsLock.Lock()
	defer cm.activeConnsLock.Unlock()

	if _, exists := cm.activeConns[sessionID]; exists {
		return fmt.Errorf("connection already exists for session ID: %s", sessionID)
	}

	cm.activeConns[sessionID] = wsConn
	return nil
}

func (cm *ConnectionManager) CreateConnection(w http.ResponseWriter, r *http.Request, h http.Header) (*websocketlib.Conn, error) {
	return cm.upgrader.Upgrade(w, r, h)
}

func (cm *ConnectionManager) RemoveConnection(sessionID string) {
	cm.activeConnsLock.Lock()
	defer cm.activeConnsLock.Unlock()

	if conn, exists := cm.activeConns[sessionID]; exists {
		conn.Close()
		delete(cm.activeConns, sessionID)
	}
}

func (cm *ConnectionManager) GetConnectionInfo(sessionID string) (*websocketlib.Conn, error) {
	cm.activeConnsLock.RLock()
	defer cm.activeConnsLock.RUnlock()

	conn, exists := cm.activeConns[sessionID]
	if !exists {
		return nil, fmt.Errorf("no active connection found for session ID: %s", sessionID)
	}
	return conn, nil
}

func (cm *ConnectionManager) GetAllConnections() map[string]*websocketlib.Conn {
	cm.activeConnsLock.RLock()
	defer cm.activeConnsLock.RUnlock()

	// Return a copy of the active connections map
	connectionsCopy := make(map[string]*websocketlib.Conn)
	for sessionID, conn := range cm.activeConns {
		connectionsCopy[sessionID] = conn
	}
	return connectionsCopy
}

func (cm *ConnectionManager) BroadcastReservation(data []byte) error {
	// var broadcastMsg dto.BroadcastMsg
	// if err := json.Unmarshal(data, &broadcastMsg); err != nil {
	// 	return fmt.Errorf("failed to unmarshal msg, error", err)
	// }

	connections := cm.GetAllConnections()

	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)

		go func(c *websocketlib.Conn) {
			defer wg.Done()

			err := c.WriteMessage(websocketlib.TextMessage, data)
			if err != nil {
				log.Printf("Failed to send message to connection: %v", err)
				// Todo: handle reconnection or cleanup
			}
		}(conn)
	}

	wg.Wait()

	return nil
}

func (cm *ConnectionManager) NotifyReservation(data []byte) error {
	var notificationMsg dto.NotificationMsg
	if err := json.Unmarshal(data, &notificationMsg); err != nil {
		return fmt.Errorf("failed to unmarshal msg:%+w", err)
	}

	sessionID := notificationMsg.SessionID

	conn, err := cm.GetConnectionInfo(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get connection info:%+w", err)
	}

	err = conn.WriteMessage(websocketlib.TextMessage, data)
	if err != nil {
		log.Printf("Failed to send message to connection: %v", err)
		// Todo: handle reconnection or cleanup
		return err
	}

	return nil
}
