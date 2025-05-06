package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yigit/unisphere/internal/app/controllers"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/middleware"
)

// SetupRouter configures all application routes
func SetupRouter(
	router *gin.Engine,
	authController *controllers.AuthController,
	facultyController *controllers.FacultyController,
	departmentController *controllers.DepartmentController,
	pastExamController *controllers.PastExamController,
	classNoteController *controllers.ClassNoteController,
	communityController *controllers.CommunityController,
	userController *controllers.UserController,
	authMiddleware *middleware.AuthMiddleware,
) {
	// API version group
	v1 := router.Group("/api/v1")

	// Setup different route groups
	setupPublicRoutes(v1, facultyController, departmentController)
	setupAuthRoutes(v1, authController)
	setupUserRoutes(v1, userController, authMiddleware)
	setupContentRoutes(v1, pastExamController, classNoteController, communityController, authMiddleware, departmentController, facultyController)

	// Health check endpoint (public)
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, dto.APIResponse{
			Data: gin.H{"status": "ok"},
		})
	})

	// Test endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "status": "success"})
	})
}

// setupPublicRoutes configures public routes for faculties and departments
func setupPublicRoutes(
	v1 *gin.RouterGroup,
	facultyController *controllers.FacultyController,
	departmentController *controllers.DepartmentController,
) {
	// Faculty routes (public access)
	faculties := v1.Group("/faculties")
	{
		faculties.GET("", facultyController.GetAllFaculties)
		faculties.GET("/:id", facultyController.GetFacultyByID)
	}

	// Department routes (public access)
	departments := v1.Group("/departments")
	{
		departments.GET("", departmentController.GetAllDepartments)
		departments.GET("/:id", departmentController.GetDepartmentByID)
	}
}

// setupAuthRoutes configures authentication related routes
func setupAuthRoutes(
	v1 *gin.RouterGroup,
	authController *controllers.AuthController,
) {
	// --- Public Auth routes ---
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)
		auth.GET("/verify-email", authController.VerifyEmail)
		auth.POST("/resend-verification", authController.ResendVerificationEmail)
		auth.POST("/forgot-password", authController.ForgotPassword)
		auth.POST("/reset-password", authController.ResetPassword)
	}
}

// setupUserRoutes configures user profile and management routes
func setupUserRoutes(
	v1 *gin.RouterGroup,
	userController *controllers.UserController,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Create authenticated group
	authenticated := v1.Group("")
	authenticated.Use(authMiddleware.JWTAuth())

	// Routes available to authenticated users, even without email verification
	users := authenticated.Group("/users")
	{
		// Profile routes available without email verification
		users.GET("/profile", userController.GetUserProfile)
		users.PUT("/profile", userController.UpdateUserProfile)
		users.POST("/profile/photo", userController.UpdateProfilePhoto)
		users.DELETE("/profile/photo", userController.DeleteProfilePhoto)
	}

	// Routes that require email verification
	authenticatedWithEmailVerified := authenticated.Group("")
	authenticatedWithEmailVerified.Use(authMiddleware.EmailVerificationRequired())

	// User endpoints that require email verification
	usersVerified := authenticatedWithEmailVerified.Group("/users")
	{
		usersVerified.GET("", userController.GetAllUsers)
		usersVerified.GET("/:id", userController.GetUserByID)
	}

	// Admin protected routes
	adminProtected := authenticatedWithEmailVerified.Group("/admin")
	adminProtected.Use(authMiddleware.RoleRequired(string(models.RoleAdmin)))
	{
		// User management (Admin only)
		adminProtected.GET("/users", userController.GetAllUsers)
		adminProtected.DELETE("/users/:id", userController.DeleteUser)
	}

	// Use a different URL pattern to avoid conflicts with /departments/:id endpoint
	authenticated.GET("/department-users/:departmentId", userController.GetUsersByDepartment)
}

// setupContentRoutes configures routes for content like past exams, class notes and communities
func setupContentRoutes(
	v1 *gin.RouterGroup,
	pastExamController *controllers.PastExamController,
	classNoteController *controllers.ClassNoteController,
	communityController *controllers.CommunityController,
	authMiddleware *middleware.AuthMiddleware,
	departmentController *controllers.DepartmentController,
	facultyController *controllers.FacultyController,
) {
	// Create authenticated group with email verification
	authenticated := v1.Group("")
	authenticated.Use(authMiddleware.JWTAuth())

	// Files endpoint (global access to file details) - available without email verification
	authenticated.GET("/files/:fileId", classNoteController.GetFileDetails)

	// Routes that require email verification
	authenticatedWithEmailVerified := authenticated.Group("")
	authenticatedWithEmailVerified.Use(authMiddleware.EmailVerificationRequired())

	// Faculty protected routes
	facultiesProtected := authenticatedWithEmailVerified.Group("/faculties")
	{
		// Role-protected routes within faculties
		facultiesInstructorProtected := facultiesProtected.Group("")
		facultiesInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
		{
			facultiesInstructorProtected.POST("", facultyController.CreateFaculty)
			facultiesInstructorProtected.PUT("/:id", facultyController.UpdateFaculty)
			facultiesInstructorProtected.DELETE("/:id", facultyController.DeleteFaculty)
		}
	}

	// Department protected routes
	departmentsProtected := authenticatedWithEmailVerified.Group("/departments")
	{
		// Role-protected routes within departments
		departmentsInstructorProtected := departmentsProtected.Group("")
		departmentsInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
		{
			departmentsInstructorProtected.POST("", departmentController.CreateDepartment)
			departmentsInstructorProtected.PUT("/:id", departmentController.UpdateDepartment)
			departmentsInstructorProtected.DELETE("/:id", departmentController.DeleteDepartment)
		}
	}

	// Past Exam routes - Endpoints for accessing and managing past examination materials
	pastExams := authenticatedWithEmailVerified.Group("/past-exams")
	{
		// Public routes accessible to all authenticated users (students and instructors)
		pastExams.GET("", pastExamController.GetAllPastExams)     // List all past exams with optional filtering
		pastExams.GET("/:id", pastExamController.GetPastExamByID) // Retrieve a specific past exam by ID

		// Instructor-only routes - Protected by role-based middleware
		// These routes are restricted to users with the Instructor role
		pastExamsInstructorProtected := pastExams.Group("")
		pastExamsInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
		{
			// CRUD operations for past exam resources
			pastExamsInstructorProtected.POST("", pastExamController.CreatePastExam)       // Create a new past exam
			pastExamsInstructorProtected.PUT("/:id", pastExamController.UpdatePastExam)    // Update an existing past exam
			pastExamsInstructorProtected.DELETE("/:id", pastExamController.DeletePastExam) // Delete a past exam

			// File management for past exams
			pastExamsInstructorProtected.POST("/:id/files", pastExamController.AddFileToPastExam)                // Upload and attach files to a past exam
			pastExamsInstructorProtected.DELETE("/:id/files/:fileId", pastExamController.DeleteFileFromPastExam) // Remove a file from a past exam
		}
	}

	// Class Notes routes
	classNotes := authenticatedWithEmailVerified.Group("/class-notes")
	{
		classNotes.GET("", classNoteController.GetAllNotes)
		classNotes.GET("/:noteId", classNoteController.GetNoteByID)

		// Both students and instructors can create class notes
		classNotesAuthProtected := classNotes.Group("")
		{
			classNotesAuthProtected.POST("", classNoteController.CreateNote)
			classNotesAuthProtected.PUT("/:noteId", classNoteController.UpdateNote)
			classNotesAuthProtected.DELETE("/:noteId", classNoteController.DeleteNote)
			classNotesAuthProtected.POST("/:noteId/files", classNoteController.AddFilesToNote)
			classNotesAuthProtected.DELETE("/:noteId/files/:fileId", classNoteController.DeleteFileFromNote)
		}
	}

	// Community routes - Endpoints for accessing and managing communities
	communities := authenticatedWithEmailVerified.Group("/communities")
	{
		// Public routes accessible to all authenticated users
		communities.GET("", communityController.GetAllCommunities)    // List all communities with optional filtering
		communities.GET("/:id", communityController.GetCommunityByID) // Retrieve a specific community by ID

		// Routes that require authentication
		communitiesAuthProtected := communities.Group("")
		{
			// CRUD operations for communities
			communitiesAuthProtected.POST("", communityController.CreateCommunity)       // Create a new community
			communitiesAuthProtected.PUT("/:id", communityController.UpdateCommunity)    // Update an existing community
			communitiesAuthProtected.DELETE("/:id", communityController.DeleteCommunity) // Delete a community

			// File management for communities
			communitiesAuthProtected.POST("/:id/files", communityController.AddFileToCommunity)                // Upload and attach files
			communitiesAuthProtected.DELETE("/:id/files/:fileId", communityController.DeleteFileFromCommunity) // Remove a file

			// Profile photo management
			communitiesAuthProtected.POST("/:id/profile-photo", communityController.UpdateProfilePhoto)   // Update profile photo
			communitiesAuthProtected.DELETE("/:id/profile-photo", communityController.DeleteProfilePhoto) // Delete profile photo

			// Participant management
			communitiesAuthProtected.GET("/:id/participants", communityController.GetCommunityParticipants) // Get all participants
			communitiesAuthProtected.POST("/:id/participants", communityController.JoinCommunity)           // Join community
			communitiesAuthProtected.DELETE("/:id/participants", communityController.LeaveCommunity)        // Leave community
			communitiesAuthProtected.GET("/:id/participants/check", communityController.IsUserParticipant)  // Check if user is participant
		}
	}
}
