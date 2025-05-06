package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	// Removed direct dependency imports, they are handled in bootstrap
	"github.com/yigit/unisphere/internal/bootstrap"
	"github.com/yigit/unisphere/internal/config"
)

// Server holds the state for the HTTP server.
type Server struct {
	config *config.Config
	router *gin.Engine
	dbPool *pgxpool.Pool
	logger zerolog.Logger
	http   *http.Server
}

// NewServer creates and initializes a new server instance by calling bootstrap functions.
func NewServer() (*Server, error) {
	cfg, lgr, err := bootstrap.LoadConfigAndSetupLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to load config or setup logger: %w", err)
	}

	dbPool, err := bootstrap.SetupDatabase(cfg, lgr)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	deps, err := bootstrap.BuildDependencies(cfg, dbPool, lgr)
	if err != nil {
		// Attempt to close DB pool if DI fails
		dbPool.Close()
		return nil, fmt.Errorf("failed to setup dependencies: %w", err)
	}

	router := bootstrap.SetupRouter(cfg, deps, lgr)

	s := &Server{
		config: cfg,
		router: router,
		dbPool: dbPool,
		logger: lgr,
	}

	return s, nil
}

// Run starts the HTTP server and handles graceful shutdown.
func (s *Server) Run() error {
	s.logger.Info().Str("port", s.config.Server.Port).Msg("Starting server...")

	s.http = &http.Server{
		Addr:         ":" + s.config.Server.Port,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for errors starting the server
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		s.logger.Info().Str("addr", s.http.Addr).Msg("HTTP server listening")
		serverErrors <- s.http.ListenAndServe()
	}()

	// Channel to listen for OS signals
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive either a server error or an OS signal
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("error starting server: %w", err)
		}
	case sig := <-osSignals:
		s.logger.Info().Str("signal", sig.String()).Msg("Received OS signal, initiating shutdown...")
	}

	// Perform graceful shutdown
	return s.Shutdown(context.Background())
}

// Shutdown gracefully stops the server and closes resources.
func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // Increased timeout slightly
	defer cancel()

	shutdownError := false

	// Shutdown HTTP server
	if s.http != nil {
		s.logger.Info().Msg("Shutting down HTTP server...")
		if err := s.http.Shutdown(ctx); err != nil {
			s.logger.Error().Err(err).Msg("HTTP server shutdown error")
			shutdownError = true
		} else {
			s.logger.Info().Msg("HTTP server gracefully stopped.")
		}
	}

	// Close database pool
	if s.dbPool != nil {
		s.logger.Info().Msg("Closing database connection pool...")
		s.dbPool.Close()
		s.logger.Info().Msg("Database connection pool closed.")
	}

	s.logger.Info().Msg("Server shutdown process complete.")
	if shutdownError {
		return errors.New("server shutdown completed with errors")
	}
	return nil
}
