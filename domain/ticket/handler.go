package ticket

// const reservationExpireDuration = 30 * time.Minute // Example expiry time for the reservation

// func InitiateBookingHandler(c *gin.Context) {
// 	var req BookingRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 		return
// 	}

// 	service := TicketService{} // Initialize your service

// 	// Check ticket availability in Redis
// 	available, err := service.CheckTicketAvailability(req.TicketID)
// 	if err != nil || !available {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket not available"})
// 		return
// 	}

// 	// Temporary reservation in Redis with an expiry time
// 	err = service.TempReserveTicket(req.TicketID, req.UserID, reservationExpireDuration)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reserve ticket"})
// 		return
// 	}

// 	// Respond to the user indicating that the booking process has started
// 	c.JSON(http.StatusAccepted, gin.H{"message": "Seats reserved, waiting for payment"})
// }
