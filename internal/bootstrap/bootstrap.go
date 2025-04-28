package bootstrap

import (
	"context"
	"fmt"
	"net/http"
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
	"github.com/yigit/unisphere/internal/pkg/helpers"
	"github.com/yigit/unisphere/internal/pkg/logger"
	"github.com/yigit/unisphere/internal/seed" // Import the new seed package
)

// Dependencies holds all the application dependencies
type Dependencies struct {
	AuthService          appServices.AuthService       // Interface type
	InstructorService    appServices.InstructorService // Interface type
	FacultyService       appServices.FacultyService    // Interface type
	DepartmentService    appServices.DepartmentService // Interface type
	PastExamService      appServices.PastExamService   // Interface type
	ClassNoteService     appServices.ClassNoteService  // Interface type
	AuthController       *appControllers.AuthController
	FacultyController    *appControllers.FacultyController
	DepartmentController *appControllers.DepartmentController
	InstructorController *appControllers.InstructorController
	PastExamController   *appControllers.PastExamController
	ClassNoteController  *appControllers.ClassNoteController
	AuthMiddleware       *appMiddleware.AuthMiddleware // Pointer to middleware struct
	Repos                *appRepos.Repositories        // Include the main repo container
	JWTService           *pkgAuth.JWTService
	AuthzService         *appAuth.AuthorizationService
	Logger               zerolog.Logger
	FileStorage          *filestorage.LocalStorage // Add FileStorage
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

	deps.AuthService = appServices.NewAuthService(
		deps.Repos.UserRepository,
		deps.Repos.TokenRepository,
		deps.Repos.DepartmentRepository,
		deps.Repos.FacultyRepository,
		deps.Repos.FileRepository,
		deps.FileStorage,
		deps.JWTService,
		lgr,
	)

	deps.FacultyService = appServices.NewFacultyService(deps.Repos.FacultyRepository)
	deps.DepartmentService = appServices.NewDepartmentService(deps.Repos.DepartmentRepository, deps.Repos.FacultyRepository)
	deps.InstructorService = appServices.NewInstructorService(deps.Repos.UserRepository, deps.Repos.DepartmentRepository)
	deps.PastExamService = appServices.NewPastExamService(deps.Repos.PastExamRepository, deps.Repos.DepartmentRepository, deps.AuthzService)
	deps.ClassNoteService = appServices.NewClassNoteService(deps.Repos.ClassNoteRepository, deps.Repos.DepartmentRepository, deps.AuthzService)

	deps.AuthMiddleware = appMiddleware.NewAuthMiddleware(deps.JWTService)

	deps.AuthController = appControllers.NewAuthController(
		deps.AuthService,
		deps.Repos.UserRepository,
		deps.JWTService,
		deps.Logger,
	)
	deps.FacultyController = appControllers.NewFacultyController(deps.FacultyService)
	deps.DepartmentController = appControllers.NewDepartmentController(deps.DepartmentService)
	deps.InstructorController = appControllers.NewInstructorController(deps.InstructorService)
	deps.PastExamController = appControllers.NewPastExamController(deps.PastExamService, deps.FileStorage)
	deps.ClassNoteController = appControllers.NewClassNoteController(deps.ClassNoteService, deps.FileStorage)

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

	// Setup Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json"), ginSwagger.DefaultModelsExpandDepth(1)))

	// Setup API routes using the dependencies
	appRoutes.SetupRouter(router,
		deps.AuthController,
		deps.FacultyController,
		deps.DepartmentController,
		deps.InstructorController,
		deps.PastExamController,
		deps.ClassNoteController,
		deps.AuthMiddleware, // Pass the middleware struct itself
	)

	// Test endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "status": "success"})
	})

	return router
}
