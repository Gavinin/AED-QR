package handler

import (
	"AED-QR/internal/config"
	"AED-QR/internal/log"
	"AED-QR/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Captcha
	if !store.Verify(req.CaptchaID, req.Captcha, true) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid captcha"})
		return
	}

	// Verify Username and Password
	if req.Username != config.AppConfig.Admin.Username || req.Password != config.AppConfig.Admin.Password {
		log.Warnf("Failed login attempt for user: %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT
	token, err := middleware.GenerateToken(req.Username)
	if err != nil {
		log.Errorf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Infof("User %s logged in successfully", req.Username)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GetCaptcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		log.Errorf("Failed to generate captcha: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate captcha"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"captcha_id": id,
		"captcha":    b64s,
	})
}

func CheckToken(c *gin.Context) {
	c.Status(http.StatusOK)
}

func Logout(c *gin.Context) {
	c.Status(http.StatusOK)
}
