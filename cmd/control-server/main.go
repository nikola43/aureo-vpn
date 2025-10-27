package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikola43/aureo-vpn/internal/control"
	"github.com/nikola43/aureo-vpn/pkg/database"
)

func main() {
	log.Println("Starting Aureo VPN Control Server...")

	// Load configuration
	config := loadConfig()

	// Connect to database
	dbConfig := database.Config{
		Host:     config.DBHost,
		Port:     config.DBPort,
		User:     config.DBUser,
		Password: config.DBPassword,
		DBName:   config.DBName,
		SSLMode:  config.DBSSLMode,
		TimeZone: "UTC",
	}

	if err := database.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create and start control server
	controlServer := control.NewServer()
	if err := controlServer.Start(); err != nil {
		log.Fatalf("Failed to start control server: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Control Server is running. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("Shutting down Control Server...")
	if err := controlServer.Stop(); err != nil {
		log.Printf("Error stopping control server: %v", err)
	}

	log.Println("Control Server stopped successfully")
}

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func loadConfig() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "aureo_vpn"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return defaultValue
	}
	return value
}
