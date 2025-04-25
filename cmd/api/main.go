package main

import (
	"os"

	"github.com/yigit/unisphere/internal/pkg/logger" // Still needed for initial error logging
	"github.com/yigit/unisphere/internal/server"
)

// @title UniSphere API
// @version 1.0
// @description API for UniSphere university social platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://www.unisphere.com/support
// @contact.email support@unisphere.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token for authorization

func main() {
	// Initialize the server with all its dependencies
	// NewServer now orchestrates LoadConfigAndSetupLogger, SetupDatabase, SetupDependencies, SetupRouter
	srv, err := server.NewServer()
	if err != nil {
		// Use the default logger setup by the logger package's init
		// Error details are logged within NewServer's setup functions
		logger.Error().Err(err).Msg("Failed to initialize server")
		os.Exit(1)
	}

	// Run the server (this blocks until shutdown signal)
	if err := srv.Run(); err != nil {
		// Log potential errors during server run or shutdown
		// Run() logs fatal on startup failure. Shutdown() logs its errors.
		logger.Error().Err(err).Msg("Server execution failed or shutdown encountered errors")
		os.Exit(1) // Exit with error code if Run returns an error
	}

	// If Run completes without error, it means graceful shutdown was successful.
	logger.Info().Msg("Application finished gracefully.")
	os.Exit(0)
}
