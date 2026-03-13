package router

import (
	"AED-QR/internal/config"
	"AED-QR/internal/handler"
	"AED-QR/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	if config.AppConfig.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(gin.Recovery())
	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// Custom logger middleware to use Zap
	r.Use(middleware.Logger())

	// Public routes
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// Public AED routes
	public := r.Group("/public")
	public.GET("/aed/:uuid", handler.GetAEDByUUID)
	public.POST("/aed/:uuid/open", handler.OpenAED)
	public.POST("/aed/:uuid/lock", handler.LockAED)
	public.GET("/config", handler.GetPublicConfig)
	public.POST("/open/aed", handler.OpenAED)

	// Admin routes
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.RateLimiter())
	{
		// Public routes
		adminGroup.GET("/captch", handler.GetCaptcha) // Kept "captch" as per user request, but maybe should be "captcha"
		adminGroup.POST("/login", handler.Login)

		// Protected routes
		protected := adminGroup.Group("/")
		protected.Use(middleware.JWT())
		{
			protected.GET("/check_token", handler.CheckToken)
			protected.POST("/logout", handler.Logout)

			// Vehicle routes
			protected.GET("/vehicles", handler.GetVehicles)
			protected.POST("/vehicles", handler.CreateVehicle)
			protected.PUT("/vehicles/:id", handler.UpdateVehicle)
			protected.DELETE("/vehicles/:id", handler.DeleteVehicle)

			// Brand routes
			protected.GET("/brands", handler.GetBrands)

			// Tesla Auth routes
			protected.GET("/tesla/auth-url", handler.GenerateTeslaAuthURL)
			protected.POST("/tesla/exchange", handler.ExchangeTeslaToken)
			protected.POST("/tesla/products", handler.ListTeslaVehicles)

			// Future authenticated routes will go here
			protected.GET("/me", func(c *gin.Context) {
				username, _ := c.Get("username")
				c.JSON(200, gin.H{"username": username})
			})
		}
	}

	return r
}
