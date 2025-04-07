package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/project-exam/pkg/infrastructure/config"
	"github.com/project-exam/pkg/interface/api/handler"
	"github.com/project-exam/pkg/interface/api/middleware"
)

// Router manages the routes for the API
type Router struct {
	config          *config.Config
	engine          *gin.Engine
	ethereumHandler *handler.EthereumHandler
	logger          *logrus.Logger
}

// NewRouter creates a new router with the given configuration and handlers
func NewRouter(cfg *config.Config, ethereumHandler *handler.EthereumHandler, logger *logrus.Logger) *Router {
	// Set Gin mode based on configuration
	if gin.Mode() == gin.DebugMode && cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New() // Don't use Default() as we're adding our own middleware

	router := &Router{
		config:          cfg,
		engine:          engine,
		ethereumHandler: ethereumHandler,
		logger:          logger,
	}

	// Register middleware and routes
	router.setupMiddleware()
	router.registerRoutes()

	return router
}

// setupMiddleware adds global middleware to the router
func (r *Router) setupMiddleware() {
	// Recovery middleware
	r.engine.Use(middleware.Recovery(r.logger))

	// Request logging
	r.engine.Use(middleware.RequestLogger(r.logger))

	// Security headers
	r.engine.Use(middleware.SecurityHeaders())

	// CORS middleware
	r.engine.Use(middleware.CORS())

	// Request size limiter (10MB)
	r.engine.Use(middleware.RequestSizeLimiter(10 * 1024 * 1024))

	// Request timeout (30 seconds global timeout)
	r.engine.Use(middleware.Timeout(30 * time.Second))
}

// registerRoutes registers all API routes
func (r *Router) registerRoutes() {
	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(
		r.config.Server.RateLimit.Limit,
		r.config.Server.RateLimit.Window,
	)

	// Add localhost to rate limiter whitelist for development
	if gin.Mode() == gin.DebugMode {
		rateLimiter.AddToWhitelist("127.0.0.1")
		rateLimiter.AddToWhitelist("::1")
	}

	// Create API key auth middleware if enabled
	var apiKeyAuth *middleware.APIKeyAuth
	if r.config.Server.Auth.Enabled {
		apiKeyAuth = middleware.NewAPIKeyAuth()

		// Add development API key if in debug mode
		if gin.Mode() == gin.DebugMode {
			apiKeyAuth.AddAPIKey("development-api-key", "dev-user")
		}

		// Add configured API keys
		for key, userID := range r.config.Server.Auth.APIKeys {
			apiKeyAuth.AddAPIKey(key, userID)
		}
	}

	// Health check endpoint - no rate limiting or auth
	r.engine.GET("/health", r.ethereumHandler.HealthCheck)

	// API routes group
	api := r.engine.Group("/api")

	// Apply rate limiting to all API routes
	api.Use(rateLimiter.Limit())

	// Apply API key authentication if enabled
	if r.config.Server.Auth.Enabled {
		api.Use(apiKeyAuth.Authenticate())
	}

	// Apply content type enforcer for non-GET requests
	api.Use(middleware.ContentTypeEnforcer())

	// Ethereum routes
	ethereum := api.Group("/ethereum")
	{
		// Cache GET requests for 5 seconds
		ethereum.GET("/:address", middleware.CacheControl(5*time.Second), r.ethereumHandler.GetAddressInfo)
	}

	// Other potential groups
	if gin.Mode() == gin.DebugMode {
		// Debug endpoints only available in debug mode
		debug := r.engine.Group("/debug")
		debug.GET("/ping", func(c *gin.Context) {
			c.String(200, "pong")
		})
	}
}

// Engine returns the underlying Gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// Run starts the HTTP server
func (r *Router) Run() error {
	return r.engine.Run(":" + r.config.Server.Port)
}
