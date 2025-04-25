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
	"github.com/yigit/unisphere/internal/pkg/helpers" // Import the new helpers package
	"github.com/yigit/unisphere/internal/pkg/logger"
	"github.com/yigit/unisphere/internal/seed" // Import the new seed package
)

// Dependencies holds the core application components (DI container).
type Dependencies struct {
	AuthService          *appServices.AuthService       // Pointer type
	InstructorService    *appServices.InstructorService // Assuming it returns a pointer
	FacultyService       *appServices.FacultyService    // Assuming it returns a pointer
	DepartmentService    *appServices.DepartmentService // Assuming it returns a pointer
	PastExamService      *appServices.PastExamService   // Assuming it returns a pointer
	ClassNoteService     appServices.ClassNoteService   // Assuming interface implementation (value type)
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
	migrationPath := filepath.Join("migrations", "init.sql")
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		lgr.Error().Str("path", migrationPath).Msg("Migration file not found")
		return nil, fmt.Errorf("migration file not found at %s: %w", migrationPath, err)
	}
	if err := migrator.MigrateFromFile(migrationPath); err != nil {
		lgr.Error().Err(err).Msg("Database migration error")
		return nil, fmt.Errorf("database migration failed: %w", err)
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

	deps.JWTService = pkgAuth.NewJWTService(pkgAuth.JWTConfig{
		SecretKey:       cfg.JWT.Secret,
		AccessTokenExp:  helpers.ParseDuration(cfg.JWT.AccessTokenExpiration, 1*time.Hour),
		RefreshTokenExp: helpers.ParseDuration(cfg.JWT.RefreshTokenExpiration, 720*time.Hour),
		TokenIssuer:     cfg.JWT.Issuer,
	})

	deps.AuthzService = appAuth.NewAuthorizationService(deps.Repos.UserRepository, deps.Repos.PastExamRepository, deps.Repos.ClassNoteRepository)

	deps.AuthService = appServices.NewAuthService(
		deps.Repos.UserRepository,
		deps.Repos.TokenRepository,
		deps.Repos.DepartmentRepository,
		deps.Repos.FacultyRepository,
		deps.JWTService,
		lgr,
	)
	deps.InstructorService = appServices.NewInstructorService(deps.Repos.UserRepository, deps.Repos.DepartmentRepository)
	deps.FacultyService = appServices.NewFacultyService(deps.Repos.FacultyRepository)
	deps.DepartmentService = appServices.NewDepartmentService(deps.Repos.DepartmentRepository, deps.Repos.FacultyRepository)
	deps.PastExamService = appServices.NewPastExamService(deps.Repos.PastExamRepository, deps.AuthzService)
	deps.ClassNoteService = appServices.NewClassNoteService(deps.Repos.ClassNoteRepository, deps.Repos.DepartmentRepository, deps.AuthzService)

	deps.AuthMiddleware = appMiddleware.NewAuthMiddleware(deps.JWTService)

	deps.AuthController = appControllers.NewAuthController(deps.AuthService, deps.JWTService)
	deps.FacultyController = appControllers.NewFacultyController(deps.FacultyService)
	deps.DepartmentController = appControllers.NewDepartmentController(deps.DepartmentService)
	deps.InstructorController = appControllers.NewInstructorController(deps.InstructorService)
	deps.PastExamController = appControllers.NewPastExamController(deps.PastExamService)
	deps.ClassNoteController = appControllers.NewClassNoteController(deps.ClassNoteService)

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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

// --- Helper Functions (Private to bootstrap package) ---
