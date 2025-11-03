package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/nikola43/aureo-vpn/internal/api"
	"github.com/nikola43/aureo-vpn/pkg/auth"
	"github.com/nikola43/aureo-vpn/pkg/blockchain"
	"github.com/nikola43/aureo-vpn/pkg/config"
	"github.com/nikola43/aureo-vpn/pkg/database"
	apperrors "github.com/nikola43/aureo-vpn/pkg/errors"
	"github.com/nikola43/aureo-vpn/pkg/logger"
	"github.com/nikola43/aureo-vpn/pkg/metrics"
	"github.com/nikola43/aureo-vpn/pkg/middleware"
	"github.com/nikola43/aureo-vpn/pkg/operator"
	"github.com/nikola43/aureo-vpn/pkg/rewards"
)

const version = "1.0.0"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	log := logger.New(logger.Config{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		AddSource:   cfg.Logging.AddSource,
		Service:     "api-gateway",
		Version:     version,
		Environment: cfg.Logging.Environment,
	})
	logger.SetGlobal(log)

	log.Info("starting Aureo VPN API Gateway",
		"version", version,
		"environment", cfg.Logging.Environment,
	)

	// Connect to database with retry logic
	if err := connectDatabase(cfg, log); err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Error("error closing database", "error", err)
		}
	}()

	// Run database migrations
	log.Info("running database migrations")
	if err := database.AutoMigrate(); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	log.Info("database migrations completed")

	// Initialize authentication service
	tokenService := auth.NewTokenService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)
	authService := auth.NewService(tokenService)

	// Initialize blockchain service (optional - can be nil for development)
	// In production, configure with real RPC endpoints and private keys
	var blockchainService *blockchain.Service
	// Example configuration (commented out for security):
	// blockchainCfg := blockchain.Config{
	// 	EthereumRPCURL:     os.Getenv("ETHEREUM_RPC_URL"),
	// 	EthereumPrivateKey: os.Getenv("ETHEREUM_PRIVATE_KEY"),
	// 	EthereumChainID:    1, // Mainnet
	// 	BitcoinRPCURL:      os.Getenv("BITCOIN_RPC_URL"),
	// 	BitcoinRPCUser:     os.Getenv("BITCOIN_RPC_USER"),
	// 	BitcoinRPCPassword: os.Getenv("BITCOIN_RPC_PASSWORD"),
	// }
	// blockchainService, err = blockchain.NewService(blockchainCfg, log)
	// if err != nil {
	// 	log.Warn("failed to initialize blockchain service, using mock mode", "error", err)
	// }

	// Initialize reward service
	rewardService := rewards.NewRewardService(log, blockchainService)

	// Initialize reward tiers
	if err := rewardService.InitializeRewardTiers(); err != nil {
		log.Warn("failed to initialize reward tiers", "error", err)
	}

	// Initialize operator service
	operatorService := operator.NewService(log, rewardService)

	// Initialize handlers
	handlers := api.NewHandlers(authService, operatorService)

	// Create Fiber app with production configuration
	app := fiber.New(fiber.Config{
		AppName:               "Aureo VPN API Gateway v" + version,
		ServerHeader:          "", // Hide server header for security
		ErrorHandler:          customErrorHandler(log),
		BodyLimit:             cfg.Server.BodyLimit,
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		DisableStartupMessage: true, // We'll log our own message
		EnableTrustedProxyCheck: len(cfg.Security.TrustedProxies) > 0,
		TrustedProxies:        cfg.Security.TrustedProxies,
	})

	// Global middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
	}))
	app.Use(requestid.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Request logging middleware
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path}\n",
		Output: os.Stdout,
	}))

	// CORS middleware with proper configuration
	if cfg.Security.CORS.Enabled {
		corsOrigins := "*"
		if cfg.IsProduction() && len(cfg.Security.CORS.AllowedOrigins) > 0 {
			corsOrigins = cfg.Security.CORS.AllowedOrigins[0]
			for i := 1; i < len(cfg.Security.CORS.AllowedOrigins); i++ {
				corsOrigins += ", " + cfg.Security.CORS.AllowedOrigins[i]
			}
		}

		app.Use(cors.New(cors.Config{
			AllowOrigins:     corsOrigins,
			AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
			AllowCredentials: cfg.Security.CORS.AllowCredentials,
			MaxAge:           cfg.Security.CORS.MaxAge,
		}))
	}

	// Metrics middleware
	if cfg.Metrics.Enabled {
		app.Use(metrics.RecordHTTPMetrics())
	}

	// Rate limiting middleware
	if cfg.Security.RateLimit.Enabled {
		rateLimiter := middleware.NewSimpleRateLimiter(
			cfg.Security.RateLimit.MaxRequests,
			cfg.Security.RateLimit.WindowSize,
		)
		app.Use(rateLimiter.Middleware())
	}

	// Health check endpoint (no auth required)
	app.Get("/health", handlers.HealthCheck)
	app.Get("/ready", handlers.ReadinessCheck)

	// Metrics endpoint (if enabled)
	if cfg.Metrics.Enabled {
		app.Get(cfg.Metrics.Path, metrics.PrometheusHandler())
	}

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Public routes
	authRoutes := v1.Group("/auth")
	authRoutes.Post("/register", handlers.Register)
	authRoutes.Post("/login", handlers.Login)
	authRoutes.Post("/refresh", handlers.RefreshToken)

	// Protected routes (require authentication)
	authMiddleware := middleware.AuthMiddleware(tokenService)

	userRoutes := v1.Group("/user", authMiddleware)
	userRoutes.Get("/profile", handlers.GetProfile)
	userRoutes.Put("/profile", handlers.UpdateProfile)
	userRoutes.Get("/sessions", handlers.GetActiveSessions)
	userRoutes.Get("/stats", handlers.GetStats)
	userRoutes.Put("/password", handlers.ChangePassword)

	nodeRoutes := v1.Group("/nodes", authMiddleware)
	nodeRoutes.Get("/", handlers.ListNodes)
	nodeRoutes.Get("/best", handlers.GetBestNode)
	nodeRoutes.Get("/:id", handlers.GetNode)

	sessionRoutes := v1.Group("/sessions", authMiddleware)
	sessionRoutes.Post("/", handlers.CreateSession)
	sessionRoutes.Delete("/:id", handlers.DisconnectSession)
	sessionRoutes.Get("/:id", handlers.GetSession)

	configRoutes := v1.Group("/config", authMiddleware)
	configRoutes.Post("/generate", handlers.GenerateConfig)
	configRoutes.Get("/:id", handlers.GetConfig)
	configRoutes.Get("/", handlers.ListConfigs)

	// Operator routes (require authentication)
	operatorRoutes := v1.Group("/operator", authMiddleware)
	operatorRoutes.Post("/register", handlers.RegisterOperator)
	operatorRoutes.Post("/nodes", handlers.CreateOperatorNode)
	operatorRoutes.Get("/nodes", handlers.GetOperatorNodes)
	operatorRoutes.Get("/stats", handlers.GetOperatorStats)
	operatorRoutes.Get("/earnings", handlers.GetOperatorEarnings)
	operatorRoutes.Get("/payouts", handlers.GetOperatorPayouts)
	operatorRoutes.Post("/payout/request", handlers.RequestOperatorPayout)
	operatorRoutes.Get("/dashboard", handlers.GetOperatorDashboard)

	// Public operator routes (no auth required)
	v1.Get("/operator/rewards/tiers", handlers.GetRewardTiers)

	// Admin routes (require admin role)
	adminMiddleware := middleware.AdminOnlyMiddleware()
	adminRoutes := v1.Group("/admin", authMiddleware, adminMiddleware)

	adminRoutes.Get("/nodes", handlers.ListAllNodes)
	adminRoutes.Post("/nodes", handlers.CreateNode)
	adminRoutes.Put("/nodes/:id", handlers.UpdateNode)
	adminRoutes.Delete("/nodes/:id", handlers.DeleteNode)

	adminRoutes.Get("/users", handlers.ListAllUsers)
	adminRoutes.Get("/users/:id", handlers.GetUser)
	adminRoutes.Put("/users/:id", handlers.UpdateUser)
	adminRoutes.Delete("/users/:id", handlers.DeleteUser)

	adminRoutes.Get("/stats", handlers.GetSystemStats)
	adminRoutes.Get("/sessions", handlers.GetAllSessions)

	// Admin operator routes
	adminRoutes.Put("/operators/:id/verify", handlers.VerifyOperator)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "route not found",
		})
	})

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
		log.Info("api gateway listening", "address", addr)

		if cfg.Server.TLS.Enabled {
			serverErrors <- app.ListenTLS(addr, cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		} else {
			serverErrors <- app.Listen(addr)
		}
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Error("server error", "error", err)
		os.Exit(1)

	case sig := <-shutdown:
		log.Info("shutting down", "signal", sig.String())

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Error("graceful shutdown failed", "error", err)
			if err := app.Shutdown(); err != nil {
				log.Error("forced shutdown failed", "error", err)
			}
		}

		log.Info("server stopped gracefully")
	}
}

// connectDatabase connects to the database with retry logic
func connectDatabase(cfg *config.Config, log *logger.Logger) error {
	maxRetries := 5
	retryDelay := 2 * time.Second

	dbConfig := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		TimeZone:        cfg.Database.TimeZone,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		LogLevel:        cfg.Database.LogLevel,
	}

	for i := 0; i < maxRetries; i++ {
		if err := database.Connect(dbConfig); err != nil {
			log.Warn("database connection failed, retrying",
				"attempt", i+1,
				"max_retries", maxRetries,
				"error", err,
			)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
			continue
		}
		log.Info("database connected successfully")
		return nil
	}

	return fmt.Errorf("failed to connect to database after %d attempts", maxRetries)
}

// customErrorHandler handles all errors globally
func customErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default to 500 Internal Server Error
		code := fiber.StatusInternalServerError
		message := "internal server error"
		errorCode := apperrors.ErrCodeInternal
		var details map[string]interface{}

		// Check if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Check if it's an AppError
		if appErr := apperrors.GetAppError(err); appErr != nil {
			code = appErr.StatusCode
			message = appErr.Message
			errorCode = appErr.Code
			details = appErr.Details
		}

		// Log the error
		log.WithField("request_id", c.GetRespHeader("X-Request-ID")).
			LogError("request error",
				err,
				"method", c.Method(),
				"path", c.Path(),
				"status", code,
				"ip", c.IP(),
			)

		// Prepare response
		response := fiber.Map{
			"error": fiber.Map{
				"code":    errorCode,
				"message": message,
			},
		}

		if details != nil {
			response["error"].(fiber.Map)["details"] = details
		}

		// Include request ID in response
		if requestID := c.GetRespHeader("X-Request-ID"); requestID != "" {
			response["request_id"] = requestID
		}

		return c.Status(code).JSON(response)
	}
}
