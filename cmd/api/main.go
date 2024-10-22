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
	defer ginServer.Close()

	ginServer.InitServices()
	ginServer.AddMiddlewares() //before routes!
	ginServer.SetupRoutes()

	ginServer.StartConsumers() //running in background

	if err := ginServer.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
