package initial

import (
	"AED-QR/internal/config"
	"AED-QR/internal/log"
	"AED-QR/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	source := config.AppConfig.Database.Source

	log.Infof("Connecting to database %s (sqlite3)", source)

	DB, err = gorm.Open(sqlite.Open(source), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Auto Migrate
	err = DB.AutoMigrate(&model.Vehicle{}, &model.QRCode{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Info("Database connection established and migrated successfully")
}
