package middleware

import (
	"AED-QR/internal/config"
	"AED-QR/internal/log"
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Determine log level
		// Safe check for AppConfig
		level := "info"
		if config.AppConfig != nil {
			level = strings.ToLower(config.AppConfig.Log.Level)
		}
		isDebug := level == "debug"

		var requestBody []byte
		if isDebug {
			if c.Request.Body != nil {
				var err error
				requestBody, err = io.ReadAll(c.Request.Body)
				if err == nil {
					// Restore the io.ReadCloser to its original state
					c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				}
			}
		}
		if string(requestBody) == "" {
			requestBody = []byte("{}")
		}

		// Wrap ResponseWriter to capture response body
		var w *responseBodyWriter
		if isDebug {
			w = &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = w
		}

		// Process request
		c.Next()

		// Log details
		latency := time.Since(start)
		method := c.Request.Method
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		if raw != "" {
			path = path + "?" + raw
		}

		if isDebug {
			respBody := ""
			if w != nil {
				respBody = w.body.String()
			}

			log.Debugf("[GIN] %3d | %13v | %15s | %-7s %s\nRequest Body: %s\nResponse Body: %s",
				statusCode,
				latency,
				clientIP,
				method,
				path,
				string(requestBody),
				respBody,
			)
		} else {
			log.Infof("[GIN] %3d | %13v | %15s | %-7s %s",
				statusCode,
				latency,
				clientIP,
				method,
				path,
			)
		}
	}
}
