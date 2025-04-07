package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/project-exam/pkg/interface/api/response"
)

// contextKey is a custom type to prevent context key collisions
type contextKey string

// Constants for context keys
const (
	RequestIDKey contextKey = "requestID"
	StartTimeKey contextKey = "startTime"
)

// RateLimiter represents a simple IP-based rate limiter
type RateLimiter struct {
	ips          map[string][]time.Time
	limit        int
	window       time.Duration
	mu           sync.Mutex
	whitelistIPs map[string]bool // IPs that are exempt from rate limiting
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		ips:          make(map[string][]time.Time),
		limit:        limit,
		window:       window,
		whitelistIPs: make(map[string]bool),
	}

	// Start a cleanup goroutine to prevent memory leaks
	go func() {
		for {
			time.Sleep(window)
			limiter.cleanup()
		}
	}()

	return limiter
}

// AddToWhitelist adds an IP to the whitelist
func (rl *RateLimiter) AddToWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.whitelistIPs[ip] = true
}

// RemoveFromWhitelist removes an IP from the whitelist
func (rl *RateLimiter) RemoveFromWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.whitelistIPs, ip)
}

// Limit returns a middleware for rate limiting
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()

		// Skip rate limiting for whitelisted IPs
		if rl.whitelistIPs[ip] {
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Initialize if this is the first request from this IP
		if _, exists := rl.ips[ip]; !exists {
			rl.ips[ip] = make([]time.Time, 0, rl.limit)
		}

		// Remove timestamps outside the window
		now := time.Now()
		windowStart := now.Add(-rl.window)

		validRequests := rl.ips[ip][:0]
		for _, t := range rl.ips[ip] {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}

		// Check if limit has been reached
		if len(validRequests) >= rl.limit {
			rl.mu.Unlock()
			response.TooManyRequests(c)
			c.Abort()
			return
		}

		// Add current request
		rl.ips[ip] = append(validRequests, now)
		rl.mu.Unlock()

		c.Next()
	}
}

// cleanup removes old IP entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, times := range rl.ips {
		var validTimes []time.Time

		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}

		if len(validTimes) == 0 {
			delete(rl.ips, ip) // Remove the IP if no recent requests
		} else {
			rl.ips[ip] = validTimes
		}
	}
}

// RequestLogger logs information about each request
func RequestLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique ID for this request
		requestID := uuid.New().String()

		// Store the request ID and start time in the context
		c.Set(string(RequestIDKey), requestID)
		startTime := time.Now()
		c.Set(string(StartTimeKey), startTime)

		// Add request ID to response headers
		c.Header("X-Request-ID", requestID)

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Create log entry
		logEntry := logger.WithFields(logrus.Fields{
			"request_id":  requestID,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      c.Writer.Status(),
			"client_ip":   c.ClientIP(),
			"duration_ms": duration.Milliseconds(),
			"user_agent":  c.Request.UserAgent(),
			"referer":     c.Request.Referer(),
		})

		// Log based on status code
		if c.Writer.Status() >= 500 {
			logEntry.Error("Server error")
		} else if c.Writer.Status() >= 400 {
			logEntry.Warn("Client error")
		} else {
			logEntry.Info("Request processed")
		}
	}
}

// Timeout middleware aborts requests that take too long to process
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal when the request is complete
		done := make(chan struct{})

		// Process the request in a goroutine
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for the request to complete or timeout
		select {
		case <-done:
			// Request completed in time
			return
		case <-ctx.Done():
			// Request timed out
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"status":  "error",
				"message": "Request timed out",
			})
			return
		}
	}
}

// Recovery middleware handles panics and recovers gracefully
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := string(debug.Stack())

				// Log the error
				requestID, _ := c.Get(string(RequestIDKey))
				logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err,
					"stack":      stack,
				}).Error("Panic recovered")

				// Respond with internal server error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Internal server error",
				})
			}
		}()

		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}

// APIKeyAuth middleware for API key authentication
type APIKeyAuth struct {
	apiKeys map[string]string // map[apiKey]userID
	mu      sync.RWMutex
}

// NewAPIKeyAuth creates a new API key authentication middleware
func NewAPIKeyAuth() *APIKeyAuth {
	return &APIKeyAuth{
		apiKeys: make(map[string]string),
	}
}

// AddAPIKey adds an API key for a user
func (a *APIKeyAuth) AddAPIKey(apiKey, userID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.apiKeys[apiKey] = userID
}

// RemoveAPIKey removes an API key
func (a *APIKeyAuth) RemoveAPIKey(apiKey string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.apiKeys, apiKey)
}

// Authenticate checks if a valid API key is provided
func (a *APIKeyAuth) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header or query parameter
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "API key is required",
			})
			return
		}

		a.mu.RLock()
		userID, exists := a.apiKeys[apiKey]
		a.mu.RUnlock()

		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid API key",
			})
			return
		}

		// Store user ID in context
		c.Set("userID", userID)

		c.Next()
	}
}

// CacheControl sets cache control headers
func CacheControl(maxAge time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set cache control headers
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
		c.Header("Expires", time.Now().Add(maxAge).Format(http.TimeFormat))

		c.Next()
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ContentTypeEnforcer ensures correct content types for requests
func ContentTypeEnforcer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET, HEAD, OPTIONS
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")

		// Check if Content-Type header contains application/json
		if contentType == "" || !strings.Contains(strings.ToLower(contentType), "application/json") {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
				"status":  "error",
				"message": "Content-Type must be application/json",
			})
			return
		}

		c.Next()
	}
}

// RequestSizeLimiter limits the size of incoming requests
func RequestSizeLimiter(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}
