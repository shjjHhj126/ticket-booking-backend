package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	if existingConn, ok := cm.activeConns[sessionID]; ok {
		// Close the old connection to prevent stale references and free up resources
		existingConn.Close()
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

func GetWebSocketMessageBytes(msgType string, payload interface{}) []byte {
	msg := dto.BaseMessage{
		Type:    msgType,
		Payload: payload,
	}
	bytes, _ := json.Marshal(msg)
	return bytes
}

//------------------------------------------------------------------

func (cm *ConnectionManager) BroadcastReservation(data []byte) error {
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

func (cm *ConnectionManager) NotifyError(data []byte, sessionID string) error {

	conn, err := cm.GetConnectionInfo(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get connection info: %+w", err)
	}

	err = conn.WriteMessage(websocketlib.TextMessage, data) // Assuming Message is a field in the struct
	if err != nil {
		log.Printf("Failed to send message to connection: %v", err)
		// Todo: handle reconnection or cleanup
		return err
	}

	return nil
}

func (cm *ConnectionManager) NotifyReservation(data []byte, sessionID string) error {
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

// ----------------------------------------------------

func (cm *ConnectionManager) GetReservationData(ctx context.Context, reservationID string) (dto.ReservationPayload, error) {
	redisKey := fmt.Sprintf("reservation:%s", reservationID)
	data, err := cm.redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		return dto.ReservationPayload{}, fmt.Errorf("Error fetching reservation from Redis for ID %s: %v\n", reservationID, err)
	}

	//  paymentintent_id:client_secret:session_id:event_id:section_id:row_id:start_seat_number:length:price
	parts := strings.Split(data, ":")
	if len(parts) != 9 {
		return dto.ReservationPayload{}, fmt.Errorf("invalid reservation data format")
	}

	clientSecret := parts[1]

	eventID, err := strconv.Atoi(parts[3])
	if err != nil {
		return dto.ReservationPayload{}, fmt.Errorf("invalid event ID: %v", err)
	}

	sectionID, err := strconv.Atoi(parts[4])
	if err != nil {
		return dto.ReservationPayload{}, fmt.Errorf("invalid section ID: %v", err)
	}

	rowID, err := strconv.Atoi(parts[5])
	if err != nil {
		return dto.ReservationPayload{}, fmt.Errorf("invalid row ID: %v", err)
	}

	length, err := strconv.Atoi(parts[7])
	if err != nil {
		return dto.ReservationPayload{}, fmt.Errorf("invalid length: %v", err)
	}

	price, err := strconv.Atoi(parts[8])
	if err != nil {
		return dto.ReservationPayload{}, fmt.Errorf("invalid price: %v", err)
	}

	// Construct the ReservationPayload
	return dto.ReservationPayload{
		EventID:            eventID,
		SectionID:          sectionID,
		RowID:              rowID,
		Price:              price,
		Length:             length,
		StripeClientSecret: clientSecret,
	}, nil
}
