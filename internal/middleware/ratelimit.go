package middleware

import (
	"AED-QR/internal/config"
	"AED-QR/internal/log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipRecord struct {
	requestCount int
	firstRequest time.Time
	blockedUntil time.Time
}

var (
	ipStore = make(map[string]*ipRecord)
	mutex   sync.Mutex
)

// RateLimiter middleware limits requests based on IP address
func RateLimiter() gin.HandlerFunc {
	// Start cleanup routine if not already started
	go cleanupIPStore()

	return func(c *gin.Context) {
		// Only apply rate limiting to login and captcha endpoints
		path := c.Request.URL.Path
		if path != "/admin/login" && path != "/admin/captch" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		cfg := config.AppConfig.RateLimit
		window := time.Duration(cfg.Window) * time.Second
		blockDuration := time.Duration(cfg.BlockDuration) * time.Second

		mutex.Lock()
		defer mutex.Unlock()

		record, exists := ipStore[ip]
		now := time.Now()

		if !exists {
			ipStore[ip] = &ipRecord{
				requestCount: 1,
				firstRequest: now,
			}
			c.Next()
			return
		}

		// Check if currently blocked
		if !record.blockedUntil.IsZero() {
			if now.Before(record.blockedUntil) {
				log.Warnf("IP %s is blocked until %s", ip, record.blockedUntil)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Too many requests. Please try again later.",
					"retry_after": int(time.Until(record.blockedUntil).Seconds()),
				})
				c.Abort()
				return
			}
			// Block expired, reset record
			record.blockedUntil = time.Time{}
			record.requestCount = 0
			record.firstRequest = now
		}

		// Check if window expired
		if now.Sub(record.firstRequest) > window {
			// Window expired, reset counter
			record.requestCount = 1
			record.firstRequest = now
			c.Next()
			return
		}

		// Increment count
		record.requestCount++

		// Check limit
		if record.requestCount > cfg.MaxRequests {
			record.blockedUntil = now.Add(blockDuration)
			log.Warnf("IP %s exceeded rate limit. Blocking for %v", ip, blockDuration)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests. Please try again later.",
				"retry_after": int(blockDuration.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanupIPStore removes old entries to save memory
func cleanupIPStore() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		mutex.Lock()
		for ip, record := range ipStore {
			// Remove records that haven't been active for a while (e.g., 2 * window)
			// And are not currently blocked
			if config.AppConfig == nil {
				continue
			}
			window := time.Duration(config.AppConfig.RateLimit.Window) * time.Second
			if time.Since(record.firstRequest) > 2*window && record.blockedUntil.IsZero() {
				delete(ipStore, ip)
			}
		}
		mutex.Unlock()
	}
}
