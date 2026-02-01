package main

import (
	"context"
	"log"
	"os"

	"vn-admin-api/internal/config"
	"vn-admin-api/internal/crawler"
	"vn-admin-api/internal/database"
	"vn-admin-api/internal/logger"
)

func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Init Logger
	appLog := logger.New("logs/crawler.log", false) // Set debug=true if needed
	appLog.Info("Starting Application")

	// 3. Connect DB
	repo, err := database.Connect(cfg)
	if err != nil {
		appLog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// 4. Init Schema
	// Read schema file
	schemaBytes, err := os.ReadFile("internal/database/schema.sql")
	if err != nil {
		appLog.Error("Failed to read schema file", "error", err)
		os.Exit(1)
	}
	if err := repo.InitSchema(string(schemaBytes)); err != nil {
		appLog.Error("Failed to init schema", "error", err)
		os.Exit(1)
	}

	// 5. Run Crawler
	c := crawler.New(repo, appLog, cfg)
	if err := c.Run(context.Background()); err != nil {
		appLog.Error("Crawler failed", "error", err)
		os.Exit(1)
	}
}
