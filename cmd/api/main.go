package main

import (
	"log"
	"ticket-booking-backend/cmd/api/server"

	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ginServer := server.NewServer()

	ginServer.InitServices()
	ginServer.SetupRoutes()

	// // Start worker goroutines
	// for i := 0; i < 5; i++ {
	// 	go startWorker(server.rmq, server.rdb, server.db)
	// }

	ginServer.Run(":8080")
}
