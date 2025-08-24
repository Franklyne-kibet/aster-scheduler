package main

import (
	"fmt"
	"log"

	"github.com/Franklyne-kibet/aster-scheduler/internal/config"
)


func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil { 
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Aster API Server Starting ...\n")
	fmt.Printf("Database URL: %s\n", cfg.DatabaseURL)
	fmt.Printf("API Port: %d\n", cfg.APIPort)
	fmt.Printf("Worker Pool Size: %d\n", cfg.WorkerPoolSize)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Leader Election TTL: %s\n", cfg.LeaderElectionTTL)

	// TODO: Add server here
	fmt.Println("Config loaded successfully!")
}
