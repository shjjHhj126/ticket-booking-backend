package sqldb

import (
	"database/sql"
	"fmt"
	"log"

	"ticket-booking-backend/tool/util"
	"time"

	_ "github.com/lib/pq"
)

const (
	defaultDBHost     = "localhost"
	defaultDBPort     = "5432"
	defaultDBUser     = "postgres"
	defaultDBPassword = "password"
	defaultDBName     = "ticket-booking-db"
	defaultDBMaxConn  = 10
)

func InitPostgres() *sql.DB {
	dbHost := util.GetEnvOrDefault("POSTGRES_HOST", defaultDBHost)
	dbPort := util.GetEnvOrDefault("POSTGRES_PORT", defaultDBPort)
	dbUser := util.GetEnvOrDefault("POSTGRES_USER", defaultDBUser)
	dbPassword := util.GetEnvOrDefault("POSTGRES_PASSWORD", defaultDBPassword)
	dbName := util.GetEnvOrDefault("POSTGRES_DATABASE", defaultDBName)
	dbMaxConn := util.GetEnvIntOrDefault("POSTGRES_MAX_CONN", defaultDBMaxConn)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Attempting to connect to PostgreSQL with %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	db.SetMaxOpenConns(dbMaxConn)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	log.Printf("Successfully connected to PostgreSQL. Max open connections: %d", dbMaxConn)

	cleanupExpiredSessions(db)

	return db
}

func cleanupExpiredSessions(db *sql.DB) {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < NOW();")
	if err != nil {
		log.Printf("Error cleaning up expired sessions: %v", err)
	} else {
		log.Println("Expired sessions cleaned up successfully.")
	}
}
