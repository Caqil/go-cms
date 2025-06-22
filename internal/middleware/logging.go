package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorRecoveryMiddleware handles panics and provides better error logging
func ErrorRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("[PANIC] %s %s - Panic recovered: %v", c.Request.Method, c.Request.URL.Path, recovered)
		log.Printf("[PANIC] Stack trace: %s", debug.Stack())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Internal server error occurred",
			"timestamp": time.Now().Unix(),
		})
	})
}

// DetailedRequestLogger provides enhanced request logging
func DetailedRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Log request start for upload endpoints
		if path == "/api/v1/admin/plugins/upload" {
			log.Printf("[REQUEST_START] %s %s - Content-Length: %s, Content-Type: %s",
				method, path,
				c.Request.Header.Get("Content-Length"),
				c.Request.Header.Get("Content-Type"))
		}

		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Enhanced logging for plugin uploads and errors
		if path == "/api/v1/admin/plugins/upload" || statusCode >= 400 {
			log.Printf("[REQUEST_END] %s %s - Status: %d, Duration: %v, Size: %d bytes",
				method, path, statusCode, duration, c.Writer.Size())

			// Log any errors that were set
			if len(c.Errors) > 0 {
				for _, err := range c.Errors {
					log.Printf("[REQUEST_ERROR] %s %s - Error: %v", method, path, err.Error())
				}
			}
		}
	}
}

// FileUploadMiddleware handles file upload specific configurations
func FileUploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to upload endpoints
		if c.Request.URL.Path == "/api/v1/admin/plugins/upload" {
			// Set longer timeout for uploads
			c.Request.Header.Set("X-Upload-Timeout", "300") // 5 minutes

			// Log upload start
			contentLength := c.Request.Header.Get("Content-Length")
			log.Printf("[UPLOAD_START] Content-Length: %s bytes", contentLength)
		}

		c.Next()
	}
}

// RequestLogger creates a request logging middleware
func RequestLogger() gin.HandlerFunc {
	return RequestLoggerWithConfig(DefaultLogConfig())
}

// RequestLoggerWithConfig creates a request logging middleware with custom configuration
func RequestLoggerWithConfig(config LogConfig) gin.HandlerFunc {
	var logger *log.Logger

	if config.Output != nil {
		logger = log.New(config.Output, config.Prefix, log.LstdFlags)
	} else {
		logger = log.New(os.Stdout, config.Prefix, log.LstdFlags)
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get method
		method := c.Request.Method

		// Get status code
		statusCode := c.Writer.Status()

		// Get response size
		bodySize := c.Writer.Size()

		// Build query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Get user agent
		userAgent := c.Request.UserAgent()

		// Format log message based on configuration
		var logMessage string
		if config.CustomFormat != "" {
			logMessage = formatCustomLog(config.CustomFormat, LogData{
				ClientIP:   clientIP,
				Method:     method,
				Path:       path,
				StatusCode: statusCode,
				Latency:    latency,
				BodySize:   bodySize,
				UserAgent:  userAgent,
			})
		} else {
			logMessage = fmt.Sprintf("%s %s %s %d %s %d %s",
				clientIP,
				method,
				path,
				statusCode,
				latency,
				bodySize,
				userAgent,
			)
		}

		// Log based on status code if filtering is enabled
		if config.FilterByStatus {
			if statusCode >= 400 {
				logger.Printf("[ERROR] %s", logMessage)
			} else if statusCode >= 300 {
				logger.Printf("[WARN] %s", logMessage)
			} else {
				logger.Printf("[INFO] %s", logMessage)
			}
		} else {
			logger.Printf("%s", logMessage)
		}

		// Additional logging for errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Printf("[ERROR] %s: %s", path, err.Error())
			}
		}
	}
}

type LogConfig struct {
	Output         io.Writer
	Prefix         string
	CustomFormat   string
	FilterByStatus bool
	SkipPaths      []string
}

type LogData struct {
	ClientIP   string
	Method     string
	Path       string
	StatusCode int
	Latency    time.Duration
	BodySize   int
	UserAgent  string
}

// DefaultLogConfig returns a default logging configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Output:         os.Stdout,
		Prefix:         "[CMS] ",
		FilterByStatus: true,
		SkipPaths:      []string{"/health", "/metrics"},
	}
}

// formatCustomLog formats log message using custom format string
func formatCustomLog(format string, data LogData) string {
	// Simple format replacement (in a real implementation, you might use text/template)
	formatted := format

	replacements := map[string]string{
		"${client_ip}":   data.ClientIP,
		"${method}":      data.Method,
		"${path}":        data.Path,
		"${status_code}": fmt.Sprintf("%d", data.StatusCode),
		"${latency}":     data.Latency.String(),
		"${body_size}":   fmt.Sprintf("%d", data.BodySize),
		"${user_agent}":  data.UserAgent,
		"${timestamp}":   time.Now().Format(time.RFC3339),
	}

	for placeholder, value := range replacements {
		formatted = replaceAll(formatted, placeholder, value)
	}

	return formatted
}

// Simple string replacement function
func replaceAll(s, old, new string) string {
	// Simple implementation; in production, you'd use strings.ReplaceAll
	result := ""
	for {
		index := findString(s, old)
		if index == -1 {
			result += s
			break
		}
		result += s[:index] + new
		s = s[index+len(old):]
	}
	return result
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// StructuredLogger creates a structured JSON logger middleware
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// Create structured log entry
		logEntry := map[string]interface{}{
			"timestamp":   time.Now().Format(time.RFC3339),
			"client_ip":   c.ClientIP(),
			"method":      c.Request.Method,
			"path":        path,
			"status_code": c.Writer.Status(),
			"latency_ms":  latency.Milliseconds(),
			"body_size":   c.Writer.Size(),
			"user_agent":  c.Request.UserAgent(),
		}

		// Add user information if available
		if userID, exists := c.Get("user_id"); exists {
			logEntry["user_id"] = userID
		}
		if username, exists := c.Get("username"); exists {
			logEntry["username"] = username
		}

		// Add errors if any
		if len(c.Errors) > 0 {
			var errors []string
			for _, err := range c.Errors {
				errors = append(errors, err.Error())
			}
			logEntry["errors"] = errors
		}

		// In a real implementation, you'd use a proper JSON logger like logrus or zap
		fmt.Printf("LOG: %+v\n", logEntry)
	}
}
