package server

import (
	"context"
	"log"
	"ticket-booking-backend/cmd/api/redisexpirescanner"
	"time"
)

func (s *Server) StartBackgroundRoutines() {
	go func() {
		if err := s.mq.ConsumeMessages("book", s.services.ticketService.HandleReservationMessage); err != nil {
			log.Printf("Failed to start booking consumer: %v", err)
		}
	}()

	go func() {
		if err := s.mq.ConsumeMessages("pay", s.services.paymentService.HandleBookedMessage); err != nil {
			log.Printf("Failed to start payment consumer: %v", err)
		}
	}()

	go func() {
		scanner := redisexpirescanner.NewExpirationScanner(s.redisClient, 10*time.Minute, s.services.paymentService)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := scanner.Start(ctx); err != nil {
			log.Printf("Scanner error: %v", err)
		}
	}()

}
