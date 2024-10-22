package server

import "log"

func (s *Server) StartConsumers() {
	go func() {
		if err := s.mq.ConsumeMessages("book", s.services.ticketService.HandleBookingMessage); err != nil {
			log.Printf("Failed to start booking consumer: %v", err)
		}
	}()

	// go func() {
	// 	if err := s.mq.ConsumeMessages("broadcast", s.services.ticketService.HandleBroadcastMessage); err != nil {
	// 		log.Printf("Failed to start broadcast consumer: %v", err)
	// 	}
	// }()
}
