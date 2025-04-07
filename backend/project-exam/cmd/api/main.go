package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/project-exam/pkg/infrastructure/config"
	"github.com/project-exam/pkg/infrastructure/ethereum"
	"github.com/project-exam/pkg/infrastructure/persistence"
	"github.com/project-exam/pkg/interface/api/handler"
	"github.com/project-exam/pkg/interface/api/router"
	"github.com/project-exam/pkg/interface/validator"
	"github.com/project-exam/pkg/usecase"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set up logger
	logger := setupLogger(cfg.Log)
	logger.Info("Starting Ethereum Data API")

	// Initialize Ethereum client
	logger.Info("Connecting to Ethereum node...")
	ethClient, err := ethereum.NewClient(&cfg.Ethereum)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Ethereum client")
	}
	defer ethClient.Close()

	// Check connection to Ethereum node
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	blockNumber, err := ethClient.EthClient.BlockNumber(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Ethereum node")
	}
	logger.WithField("blockNumber", blockNumber).Info("Successfully connected to Ethereum node")

	// Initialize repository layer
	ethereumRepo := persistence.NewEthereumRepository(ethClient)

	// Initialize use case layer
	ethereumUseCase := usecase.NewEthereumUseCase(ethereumRepo)

	// Initialize interface layer
	ethereumValidator := validator.NewEthereumValidator()
	ethereumHandler := handler.NewEthereumHandler(ethereumUseCase, ethereumValidator)

	// Create router
	router := router.NewRouter(cfg, ethereumHandler, logger)

	// Start server in a goroutine
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Server starting")
		if err := router.Run(); err != nil {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	// Kill (no param) default sends syscall.SIGTERM
	// Kill -2 is syscall.SIGINT
	// Kill -9 is syscall.SIGKILL but can't be caught, so don't need to specify it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline for the shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close resources
	ethClient.Close()

	logger.Info("Server exited properly")
}

// setupLogger configures the logger based on configuration
func setupLogger(cfg config.LogConfig) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Configure formatter
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Configure output
	switch cfg.OutputPath {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	default:
		// Try to open file for writing logs
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Failed to open log file, using stdout: %v\n", err)
			logger.SetOutput(os.Stdout)
		} else {
			// Use MultiWriter to write logs to both file and stdout
			mw := io.MultiWriter(os.Stdout, file)
			logger.SetOutput(mw)
		}
	}

	return logger
}
