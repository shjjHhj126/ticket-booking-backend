package util

import (
	"log"
	"os"
	"strconv"
)

func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("%s not set, using default: %s", key, defaultValue)
		return defaultValue
	}
	return value
}

func GetEnvIntOrDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		log.Printf("%s not set, using default: %d", key, defaultValue)
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid %s: %s. Using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}
