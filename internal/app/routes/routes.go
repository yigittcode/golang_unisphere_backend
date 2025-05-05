package routes

import (
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
	authMiddleware *middleware.AuthMiddleware,
) {
	// API version group
	v1 := router.Group("/api/v1")

	// --- Public Faculty and Department routes ---
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
		// Profile route moved to authenticated group
	}

	// --- Authenticated Routes Group ---
	authenticated := v1.Group("")               // Create a new group for all authenticated routes
	authenticated.Use(authMiddleware.JWTAuth()) // Apply JWT Auth middleware to this group

	// Apply email verification middleware to all routes except user profile
	// This ensures users can at least access their profile even if email is not verified
	authenticatedWithEmailVerified := authenticated.Group("")
	authenticatedWithEmailVerified.Use(authMiddleware.EmailVerificationRequired())
	{
		// Profile routes are handled by the UserController in bootstrap.go

		// Files endpoint (global access to file details)
		authenticated.GET("/files/:fileId", classNoteController.GetFileDetails)

		// Faculty protected routes (now under authenticated group)
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

	// Health check endpoint (public)
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, dto.APIResponse{
			Data: gin.H{"status": "ok"},
		})
	})

	// Swagger routes are set up in bootstrap.go already
}
