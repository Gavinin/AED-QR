package handler

import (
	"AED-QR/internal/services/tesla"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ListTeslaVehiclesRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
	APIRegion    string `json:"api_region" binding:"required"` // auth.tesla.cn or auth.tesla.com
}

// ListTeslaVehicles lists vehicles for selection
func ListTeslaVehicles(c *gin.Context) {
	var req ListTeslaVehiclesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	products, newAccess, newRefresh, err := tesla.ListVehicles(req.AccessToken, req.RefreshToken, req.APIRegion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":      products,
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	})
}
