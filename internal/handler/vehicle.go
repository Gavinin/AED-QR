package handler

import (
	"AED-QR/internal/config"
	"AED-QR/internal/initial"
	"AED-QR/internal/model"
	"AED-QR/internal/services"
	"AED-QR/internal/services/tesla"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateVehicle
func CreateVehicle(c *gin.Context) {
	var vehicle model.Vehicle
	if err := c.ShouldBindJSON(&vehicle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	err := initial.DB.Transaction(func(tx *gorm.DB) error {
		// Create Vehicle
		if err := tx.Create(&vehicle).Error; err != nil {
			return err
		}

		// Generate UUID
		newUUID := uuid.New().String()

		// Create QRCode record
		qrCode := model.QRCode{
			VehicleID: vehicle.ID,
			UUID:      newUUID,
		}
		if err := tx.Create(&qrCode).Error; err != nil {
			return err
		}

		// Attach QRCode to response object (optional, for immediate use)
		vehicle.QRCode = qrCode
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vehicle"})
		return
	}

	// Update Cache
	services.UpdateVehicleCache(vehicle)

	c.JSON(http.StatusOK, vehicle)
}

// GetVehicles
func GetVehicles(c *gin.Context) {
	var vehicles []model.Vehicle
	if err := initial.DB.Preload("QRCode").Find(&vehicles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vehicles"})
		return
	}

	// Optionally construct the full QR URL here if not doing it on frontend
	// But frontend is better suited for display logic.
	// We'll just return the data including UUID.

	c.JSON(http.StatusOK, vehicles)
}

// UpdateVehicle
func UpdateVehicle(c *gin.Context) {
	id := c.Param("id")
	var vehicle model.Vehicle

	if err := initial.DB.Preload("QRCode").First(&vehicle, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}

	var input model.Vehicle
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	vehicle.Name = input.Name
	vehicle.Brand = input.Brand
	vehicle.Location = input.Location
	vehicle.Data = input.Data
	vehicle.Enabled = input.Enabled

	if err := initial.DB.Save(&vehicle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vehicle"})
		return
	}

	// Update Cache
	services.UpdateVehicleCache(vehicle)

	c.JSON(http.StatusOK, vehicle)
}

// DeleteVehicle
func DeleteVehicle(c *gin.Context) {
	id := c.Param("id")
	var vehicle model.Vehicle

	// Fetch first to get UUID for cache removal
	if err := initial.DB.Preload("QRCode").First(&vehicle, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}

	// Because of foreign key constraints or logic, we should probably delete the QRCode too.
	// GORM's cascading delete depends on definition, but let's be explicit or rely on Hooks.
	// For now, standard delete.
	if err := initial.DB.Select("QRCode").Delete(&vehicle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vehicle"})
		return
	}

	// Remove from Cache
	services.RemoveVehicleFromCache(vehicle.QRCode.UUID)

	c.JSON(http.StatusOK, gin.H{"message": "Vehicle deleted successfully"})
}

// GetBrands
func GetBrands(c *gin.Context) {
	c.JSON(http.StatusOK, config.AppConfig.Brands)
}

// GetAEDByUUID (Public)
func GetAEDByUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	var qrCode model.QRCode

	if err := initial.DB.Where("uuid = ?", uuidStr).First(&qrCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "AED not found"})
		return
	}

	var vehicle model.Vehicle
	if err := initial.DB.First(&vehicle, qrCode.VehicleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}

	if !vehicle.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "AED is disabled"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":     vehicle.Name,
		"brand":    vehicle.Brand,
		"location": vehicle.Location,
		"data":     vehicle.Data,
	})
}

// LockAED (Public Action)
func LockAED(c *gin.Context) {
	uuidStr := c.Param("uuid")

	// Use Cache
	vehicle, ok := services.GetVehicleByUUID(uuidStr)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "AED not found"})
		return
	}

	if !vehicle.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "AED is disabled"})
		return
	}

	var adapter services.IVehicle
	var err error

	switch strings.ToLower(vehicle.Brand) {
	case "tesla":
		adapter, err = tesla.NewTesla(&vehicle)
	default:
		// For now, if brand is not implemented, just return error
		err = fmt.Errorf("unsupported brand: %s", vehicle.Brand)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := adapter.Lock(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to lock vehicle: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vehicle locked successfully"})
}

// OpenAED (Public Action)
func OpenAED(c *gin.Context) {
	uuidStr := c.Param("uuid")

	// Use Cache
	vehicle, ok := services.GetVehicleByUUID(uuidStr)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "AED not found"})
		return
	}

	if !vehicle.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "AED is disabled"})
		return
	}

	var adapter services.IVehicle
	var err error

	switch strings.ToLower(vehicle.Brand) {
	case "tesla":
		adapter, err = tesla.NewTesla(&vehicle)
	default:
		// For now, if brand is not implemented, just return error
		err = fmt.Errorf("unsupported brand: %s", vehicle.Brand)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// We assume IsRunning check might be part of Open logic or pre-check
	if adapter.IsRunning() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Vehicle is running, operation prohibited"})
		return
	}

	if err := adapter.Open(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to open vehicle: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "AED opened successfully"})
}

// GetPublicConfig (Public) - Optional, if frontend needs base domain
func GetPublicConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"domain": config.AppConfig.Server.Domain,
	})
}

func getVehicleAdapter(vehicle *model.Vehicle) (services.IVehicle, error) {
	var adapter services.IVehicle
	var err error

	switch strings.ToLower(vehicle.Brand) {
	case "tesla":
		adapter, err = tesla.NewTesla(vehicle)
	default:
		err = fmt.Errorf("unsupported brand: %s", vehicle.Brand)
	}
	return adapter, err
}
