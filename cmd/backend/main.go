package main

import (
	"AED-QR/internal/config"
	"AED-QR/internal/initial"
	"AED-QR/internal/log"
	"AED-QR/internal/model"
	"AED-QR/internal/router"
	"AED-QR/internal/services"
)

func main() {
	// 1. Load configuration
	config.LoadConfig("config/config.yml")

	// 2. Initialize Logger
	log.Init(config.AppConfig.Log)
	defer log.Sync()

	log.Info("Starting application...")

	// 3. Initialize Database
	initial.InitDB()
	// defer initial.DB.Close() // GORM manages connection pool differently

	// 4. Load Cache
	var vehicles []model.Vehicle
	if err := initial.DB.Preload("QRCode").Find(&vehicles).Error; err != nil {
		log.Fatalf("Failed to load vehicles into cache: %v", err)
	}
	services.InitCache(vehicles)
	log.Infof("Loaded %d vehicles into cache", len(vehicles))

	// 5. Initialize Router
	r := router.Init()

	// 6. Start Server
	port := config.AppConfig.Server.Port
	log.Infof("Server starting on port %s", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
