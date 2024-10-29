package dto

const (
	TypeBroadcast   = "broadcast"   // tickets change in a row
	TypeReservation = "reservation" // the ticket that successfully reserved
	TypeError       = "error"       // indicate error in reservation or payment
)

type BaseMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type BroadcastPayload struct {
	Messages []struct {
		EventID   int `json:"event_id"`
		SectionID int `json:"section_id"`
		RowID     int `json:"row_id"`
		Price     int `json:"price"`
		MaxLength int `json:"max_length"`
	} `json:"messages"`
}

type ReservationPayload struct {
	EventID            int    `json:"event_id"`
	SectionID          int    `json:"section_id"`
	RowID              int    `json:"row_id"`
	Price              int    `json:"price"`
	Length             int    `json:"length"`
	StripeClientSecret string `json:"stripe_client_secret"`
}

type ErrorPayload struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Domain    string `json:"domain"` //"reservation" / "payment"
}
