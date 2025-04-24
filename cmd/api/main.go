package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/yigit/unisphere/docs" // Import generated swagger docs
	"github.com/yigit/unisphere/internal/app/auth"
	"github.com/yigit/unisphere/internal/app/controllers"
	"github.com/yigit/unisphere/internal/app/migrations"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/app/routes"
	"github.com/yigit/unisphere/internal/app/services"
	"github.com/yigit/unisphere/internal/config"
	"github.com/yigit/unisphere/internal/db"
	"github.com/yigit/unisphere/internal/middleware"
	pkgauth "github.com/yigit/unisphere/internal/pkg/auth"
	"github.com/yigit/unisphere/internal/pkg/logger"
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
	// Load configuration file
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load configuration")
		os.Exit(1)
	}

	// Configure logger based on environment
	logLevel := logger.InfoLevel
	if cfg.Server.Mode == "development" {
		logLevel = logger.DebugLevel
	}

	logger.Configure(logger.Config{
		Level:  logLevel,
		Pretty: cfg.Server.Mode != "production",
	})

	// Mode setting
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	logger.Info().Msg("Establishing database connection...")
	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		os.Exit(1)
	}
	defer database.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.Pool.Ping(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to test database connection")
		os.Exit(1)
	}
	logger.Info().Msg("Database connection successfully established.")

	// Run migrations
	logger.Info().Msg("Creating database schemas...")
	migrator := migrations.NewMigrator(database.Pool)

	// Check migration directory
	migrationPath := filepath.Join("migrations", "init.sql")
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		logger.Error().Str("path", migrationPath).Msg("Migration file not found")
		os.Exit(1)
	}

	if err := migrator.MigrateFromFile(migrationPath); err != nil {
		logger.Error().Err(err).Msg("Database migration error")
		os.Exit(1)
	}
	logger.Info().Msg("Database schemas successfully created.")

	// Create Gin router
	router := gin.Default()

	// Setup Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Repositories
	userRepo := repositories.NewUserRepository(database.Pool)
	tokenRepo := repositories.NewTokenRepository(database.Pool)
	facultyRepo := repositories.NewFacultyRepository(database.Pool)
	departmentRepo := repositories.NewDepartmentRepository(database.Pool)
	pastExamRepo := repositories.NewPastExamRepository(database.Pool)

	// JWT Service
	jwtService := pkgauth.NewJWTService(pkgauth.JWTConfig{
		SecretKey:       cfg.JWT.Secret,
		AccessTokenExp:  parseDuration(cfg.JWT.AccessTokenExpiration, 1*time.Hour),    // 1 hour default
		RefreshTokenExp: parseDuration(cfg.JWT.RefreshTokenExpiration, 720*time.Hour), // 30 days default
		TokenIssuer:     cfg.JWT.Issuer,
	})

	// Authorization Service
	authorizationService := auth.NewAuthorizationService(userRepo, pastExamRepo)

	// Services
	authService := services.NewAuthService(userRepo, tokenRepo, jwtService)
	instructorService := services.NewInstructorService(userRepo, departmentRepo)
	facultyService := services.NewFacultyService(facultyRepo)
	departmentService := services.NewDepartmentService(departmentRepo, facultyRepo)
	pastExamService := services.NewPastExamService(pastExamRepo, authorizationService)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Controllers
	authController := controllers.NewAuthController(authService, jwtService)
	facultyController := controllers.NewFacultyController(facultyService)
	departmentController := controllers.NewDepartmentController(departmentService)
	instructorController := controllers.NewInstructorController(instructorService)
	pastExamController := controllers.NewPastExamController(pastExamService)

	// Configure Router
	routes.SetupRouter(router, authController, facultyController, departmentController, instructorController, pastExamController, authMiddleware)

	// Create default data if needed
	createDefaultData(ctx, facultyRepo, departmentRepo)

	// Test endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":  "pong",
			"status":   "success",
			"database": "connected",
		})
	})

	// Start server
	logger.Info().Str("port", cfg.Server.Port).Msg("Starting server...")

	// For graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in a separate goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server start error")
		}
	}()

	// Wait for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutting down server...")

	// Timed shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server shutdown error")
	}

	logger.Info().Msg("Server successfully shut down.")
}

// parseDuration parses a duration string, returns default duration on error
func parseDuration(durationStr string, defaultDuration time.Duration) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return defaultDuration
	}
	return duration
}

// createDefaultData creates default faculties and departments if they don't exist
func createDefaultData(ctx context.Context, facultyRepo *repositories.FacultyRepository, departmentRepo *repositories.DepartmentRepository) {
	// Create Engineering Faculty
	engineeringFaculty := &models.Faculty{
		Name: "Engineering Faculty",
		Code: "ENG",
	}

	engineeringID, err := facultyRepo.CreateFaculty(ctx, engineeringFaculty)
	if err != nil && err != repositories.ErrFacultyAlreadyExists {
		logger.Error().Err(err).Msg("Error creating engineering faculty")
	}

	if err == repositories.ErrFacultyAlreadyExists {
		faculties, err := facultyRepo.GetAllFaculties(ctx)
		if err == nil {
			for _, f := range faculties {
				if f.Code == "ENG" {
					engineeringID = f.ID
					break
				}
			}
		}
	}

	if engineeringID > 0 {
		// Create Computer Engineering Department
		computerEngDept := &models.Department{
			FacultyID: engineeringID,
			Name:      "Computer Engineering",
			Code:      "CENG",
		}

		err = departmentRepo.Create(ctx, computerEngDept)
		if err != nil && err != repositories.ErrDepartmentAlreadyExists {
			logger.Error().Err(err).Msg("Error creating computer engineering department")
		}

		// Create Electrical Engineering Department
		electricalEngDept := &models.Department{
			FacultyID: engineeringID,
			Name:      "Electrical Engineering",
			Code:      "EEE",
		}

		err = departmentRepo.Create(ctx, electricalEngDept)
		if err != nil && err != repositories.ErrDepartmentAlreadyExists {
			logger.Error().Err(err).Msg("Error creating electrical engineering department")
		}
	}

	// Create Science Faculty
	scienceFaculty := &models.Faculty{
		Name: "Science Faculty",
		Code: "SCI",
	}

	scienceID, err := facultyRepo.CreateFaculty(ctx, scienceFaculty)
	if err != nil && err != repositories.ErrFacultyAlreadyExists {
		logger.Error().Err(err).Msg("Error creating science faculty")
	}

	if err == repositories.ErrFacultyAlreadyExists {
		faculties, err := facultyRepo.GetAllFaculties(ctx)
		if err == nil {
			for _, f := range faculties {
				if f.Code == "SCI" {
					scienceID = f.ID
					break
				}
			}
		}
	}

	if scienceID > 0 {
		// Create Mathematics Department
		mathDept := &models.Department{
			FacultyID: scienceID,
			Name:      "Mathematics",
			Code:      "MATH",
		}

		err = departmentRepo.Create(ctx, mathDept)
		if err != nil && err != repositories.ErrDepartmentAlreadyExists {
			logger.Error().Err(err).Msg("Error creating mathematics department")
		}

		// Create Physics Department
		physicsDept := &models.Department{
			FacultyID: scienceID,
			Name:      "Physics",
			Code:      "PHYS",
		}

		err = departmentRepo.Create(ctx, physicsDept)
		if err != nil && err != repositories.ErrDepartmentAlreadyExists {
			logger.Error().Err(err).Msg("Error creating physics department")
		}
	}
}
