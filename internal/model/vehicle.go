package model

import (
	"gorm.io/gorm"
)

type Vehicle struct {
	gorm.Model
	Name      string `json:"name"`
	Brand     string `json:"brand"`
	Location  string `json:"location"` // New field
	Data      string `json:"data"`     // Storing as JSON string
	Enabled   bool   `json:"enabled"`
	AuthError bool   `json:"auth_error"` // True if token refresh failed
	QRCode    QRCode `json:"qr_code"`    // One-to-One relationship
}

type QRCode struct {
	gorm.Model
	VehicleID uint   `json:"vehicle_id" gorm:"uniqueIndex"`
	UUID      string `json:"uuid" gorm:"uniqueIndex"`
}
