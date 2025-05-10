package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/yigit/unisphere/docs" // Import generated swagger docs
	appAuth "github.com/yigit/unisphere/internal/app/auth"
	appControllers "github.com/yigit/unisphere/internal/app/controllers"
	appMigrations "github.com/yigit/unisphere/internal/app/migrations"
	appRepos "github.com/yigit/unisphere/internal/app/repositories"
	appRoutes "github.com/yigit/unisphere/internal/app/routes"
	appServices "github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/config"
	"github.com/yigit/unisphere/internal/db"
	appMiddleware "github.com/yigit/unisphere/internal/middleware"
	pkgAuth "github.com/yigit/unisphere/internal/pkg/auth"
	"github.com/yigit/unisphere/internal/pkg/filestorage" // Import filestorage

	// Import the new helpers package
	"github.com/yigit/unisphere/internal/pkg/email" // Import email package
	"github.com/yigit/unisphere/internal/pkg/helpers"
	"github.com/yigit/unisphere/internal/pkg/logger"
	"github.com/yigit/unisphere/internal/pkg/websocket" // Import WebSocket package
	"github.com/yigit/unisphere/internal/seed" // Import the new seed package
)

// Dependencies holds all the application dependencies
type Dependencies struct {
	AuthService          appServices.AuthService       // Interface type
	UserService          appServices.UserService       // Interface type
	FacultyService       appServices.FacultyService    // Interface type
	DepartmentService    appServices.DepartmentService // Interface type
	PastExamService      appServices.PastExamService   // Interface type
	ClassNoteService     appServices.ClassNoteService  // Interface type
	CommunityService     appServices.CommunityService  // Interface type
	ChatService          appServices.ChatService       // Interface type
	AuthController       *appControllers.AuthController
	FacultyController    *appControllers.FacultyController
	DepartmentController *appControllers.DepartmentController
	UserController       *appControllers.UserController // User Controller
	PastExamController   *appControllers.PastExamController
	ClassNoteController  *appControllers.ClassNoteController
	CommunityController  *appControllers.CommunityController
	ChatController       *appControllers.ChatController
	AuthMiddleware       *appMiddleware.AuthMiddleware // Pointer to middleware struct
	Repos                *appRepos.Repositories        // Include the main repo container
	JWTService           *pkgAuth.JWTService
	AuthzService         *appAuth.AuthorizationService
	EmailService         email.EmailService
	Logger               zerolog.Logger
	FileStorage          *filestorage.LocalStorage // Add FileStorage
	WSHub                *websocket.Hub            // WebSocket hub for real-time communication
	WSHandler            *websocket.Handler        // WebSocket connection handler
}

// LoadConfigAndSetupLogger loads configuration and initializes the logger.
func LoadConfigAndSetupLogger() (*config.Config, zerolog.Logger, error) {
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load configuration")
		return nil, zerolog.Logger{}, err // Return zero logger and the error
	}

	logLevel := logger.LogLevel(strings.ToLower(cfg.Logging.Level))
	prettyLog := strings.ToLower(cfg.Logging.Format) == "text"

	logger.Configure(logger.Config{
		Level:  logLevel,
		Pretty: prettyLog,
	})

	lgr := log.Logger // Get the configured global logger
	lgr.Info().Str("logLevel", string(logLevel)).Str("logFormat", cfg.Logging.Format).Msg("Logger configured")
	return cfg, lgr, nil
}

// SetupDatabase establishes the database connection and runs migrations.
func SetupDatabase(cfg *config.Config, lgr zerolog.Logger) (*pgxpool.Pool, error) {
	lgr.Info().Msg("Establishing database connection...")
	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		lgr.Error().Err(err).Msg("Failed to connect to database")
		return nil, err
	}
	dbPool := database.Pool

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbPool.Ping(ctx); err != nil {
		lgr.Error().Err(err).Msg("Failed to ping database")
		dbPool.Close()
		return nil, err
	}
	lgr.Info().Msg("Database connection successfully established.")

	// Run migrations
	lgr.Info().Msg("Running database migrations...")
	migrator := appMigrations.NewMigrator(dbPool)

	// Use the migrations directory instead of just init.sql
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		lgr.Error().Str("path", migrationsDir).Msg("Migrations directory not found")
		return nil, fmt.Errorf("migrations directory not found at %s: %w", migrationsDir, err)
	}

	if err := migrator.MigrateFromDirectory(migrationsDir); err != nil {
		lgr.Error().Err(err).Msg("Database migration error")
		return nil, fmt.Errorf("database migrations failed: %w", err)
	}

	lgr.Info().Msg("Database migrations successfully applied.")

	// Create Default Data (after migrations)
	if err := seed.CreateDefaultData(context.Background(), dbPool, lgr); err != nil {
		// Log the error but don't necessarily fail the startup
		lgr.Error().Err(err).Msg("Failed to create default data, proceeding anyway...")
	}

	return dbPool, nil
}

// BuildDependencies initializes application repositories, services, and controllers.
func BuildDependencies(cfg *config.Config, dbPool *pgxpool.Pool, lgr zerolog.Logger) (*Dependencies, error) {
	deps := &Dependencies{Logger: lgr}

	deps.Repos = appRepos.NewRepositories(dbPool)

	// Initialize File Storage
	// Configure baseURL to match the static file serving endpoint
	var err error
	// Change the relative path to an absolute URL including host and port
	baseUrl := "http://localhost:" + cfg.Server.Port
	fileStorageBaseURL := baseUrl + "/uploads" // This must match the static file serving URL path
	deps.FileStorage, err = filestorage.NewLocalStorage(cfg.Server.StoragePath, fileStorageBaseURL)
	if err != nil {
		lgr.Error().Err(err).Msg("Failed to initialize file storage")
		return nil, fmt.Errorf("failed to initialize file storage: %w", err)
	}

	// Initialize services
	deps.AuthzService = appAuth.NewAuthorizationService(
		deps.Repos.UserRepository,
		deps.Repos.ClassNoteRepository,
		deps.Repos.PastExamRepository,
	)

	deps.JWTService = pkgAuth.NewJWTService(pkgAuth.JWTConfig{
		SecretKey:       cfg.JWT.Secret,
		AccessTokenExp:  helpers.ParseDuration(cfg.JWT.AccessTokenExpiration, 1*time.Hour),
		RefreshTokenExp: helpers.ParseDuration(cfg.JWT.RefreshTokenExpiration, 720*time.Hour),
		TokenIssuer:     cfg.JWT.Issuer,
	})

	// Initialize Email Service
	deps.EmailService = email.NewEmailService(email.SMTPConfig{
		Host:      cfg.SMTP.Host,
		Port:      cfg.SMTP.Port,
		Username:  cfg.SMTP.Username,
		Password:  cfg.SMTP.Password,
		FromName:  cfg.SMTP.FromName,
		FromEmail: cfg.SMTP.FromEmail,
		UseTLS:    cfg.SMTP.UseTLS,
		BaseURL:   baseUrl, // Use the same base URL as file storage
	}, lgr)

	deps.AuthService = appServices.NewAuthService(
		deps.Repos.UserRepository,
		deps.Repos.TokenRepository,
		deps.Repos.DepartmentRepository,
		deps.Repos.FacultyRepository,
		deps.Repos.FileRepository,
		deps.FileStorage,
		deps.Repos.VerificationTokenRepository,
		deps.Repos.PasswordResetTokenRepository,
		deps.EmailService,
		deps.JWTService,
		lgr,
	)

	deps.FacultyService = appServices.NewFacultyService(deps.Repos.FacultyRepository)
	deps.DepartmentService = appServices.NewDepartmentService(deps.Repos.DepartmentRepository, deps.Repos.FacultyRepository)

	// Initialize User Service
	deps.UserService = appServices.NewUserService(
		deps.Repos.UserRepository,
		deps.Repos.DepartmentRepository,
		deps.Repos.FileRepository,
		deps.FileStorage,
		deps.AuthService,
		deps.Logger,
	)

	deps.PastExamService = appServices.NewPastExamService(
		deps.Repos.PastExamRepository,
		deps.Repos.DepartmentRepository,
		deps.Repos.FileRepository,
		deps.FileStorage,
		deps.AuthzService,
		deps.Logger,
	)
	deps.ClassNoteService = appServices.NewClassNoteService(
		deps.Repos.ClassNoteRepository,
		deps.Repos.DepartmentRepository,
		deps.Repos.FileRepository,
		deps.FileStorage,
		deps.AuthzService,
		deps.Logger,
	)

	deps.CommunityService = appServices.NewCommunityService(
		deps.Repos.CommunityRepository,
		deps.Repos.CommunityParticipantRepository,
		deps.Repos.UserRepository,
		deps.Repos.FileRepository,
		deps.FileStorage,
		deps.AuthzService,
		deps.Logger,
	)
	
	// Initialize WebSocket Hub
	deps.WSHub = websocket.NewHub(deps.Logger)
	
	// Start the WebSocket Hub in a separate goroutine
	go deps.WSHub.Run()
	lgr.Info().Msg("WebSocket hub initialized and running")
	
	// Initialize WebSocket Handler
	deps.WSHandler = websocket.NewHandler(
		deps.WSHub,
		deps.Repos.CommunityParticipantRepository,
		deps.Logger,
	)
	lgr.Info().Msg("WebSocket handler initialized")
	
	// Initialize WebSocket Message Handler for database persistence
	messageHandler := websocket.NewMessageHandler(
		deps.Repos.ChatRepository,
		deps.Repos.UserRepository,
		deps.WSHub,
		deps.Logger,
	)
	
	// Start the message handler
	messageHandler.Start()
	lgr.Info().Msg("WebSocket message handler initialized and running")
	
	// Initialize Chat Service with WebSocket Hub
	deps.ChatService = appServices.NewChatService(
		deps.Repos.ChatRepository,
		deps.Repos.CommunityRepository,
		deps.Repos.CommunityParticipantRepository,
		deps.Repos.UserRepository,
		deps.Repos.FileRepository,
		deps.FileStorage,
		deps.WSHub,
		deps.Logger,
	)

	deps.AuthMiddleware = appMiddleware.NewAuthMiddleware(deps.JWTService, deps.Repos.UserRepository)

	deps.AuthController = appControllers.NewAuthController(
		deps.AuthService,
		deps.Repos.UserRepository,
		deps.JWTService,
		deps.Logger,
	)
	deps.FacultyController = appControllers.NewFacultyController(deps.FacultyService)
	deps.DepartmentController = appControllers.NewDepartmentController(deps.DepartmentService)
	deps.UserController = appControllers.NewUserController(deps.UserService, deps.FileStorage)
	deps.PastExamController = appControllers.NewPastExamController(deps.PastExamService, deps.FileStorage)
	deps.ClassNoteController = appControllers.NewClassNoteController(deps.ClassNoteService, deps.FileStorage)
	deps.CommunityController = appControllers.NewCommunityController(deps.CommunityService, deps.FileStorage)
	deps.ChatController = appControllers.NewChatController(deps.ChatService)

	return deps, nil
}

// SetupRouter configures the Gin engine with middleware and routes.
func SetupRouter(cfg *config.Config, deps *Dependencies, lgr zerolog.Logger) *gin.Engine {
	if strings.ToLower(cfg.Server.Mode) == "production" {
		gin.SetMode(gin.ReleaseMode)
		lgr.Info().Msg("Setting Gin mode to release")
	} else {
		gin.SetMode(gin.DebugMode)
		lgr.Info().Msg("Setting Gin mode to debug")
	}

	router := gin.Default()

	// Apply CORS middleware to all routes
	router.Use(appMiddleware.CORSMiddleware())
	lgr.Info().Msg("CORS middleware enabled: Allowing all origins")

	// Setup Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json"), ginSwagger.DefaultModelsExpandDepth(1)))

	// Setup static files for frontend
	router.Static("/public", "./public")

	// Serve uploaded files
	router.Static("/uploads", cfg.Server.StoragePath)

	// Setup all API routes
	appRoutes.SetupRouter(router,
		deps.AuthController,
		deps.FacultyController,
		deps.DepartmentController,
		deps.PastExamController,
		deps.ClassNoteController,
		deps.CommunityController,
		deps.UserController,
		deps.ChatController,
		deps.WSHandler,
		deps.AuthMiddleware,
	)

	return router
}
