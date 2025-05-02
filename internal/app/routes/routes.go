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
	authMiddleware *middleware.AuthMiddleware,
) {
	// API version group
	v1 := router.Group("/api/v1")

	// --- Public Auth routes ---
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)
		// Profile route moved to authenticated group
	}

	// --- Authenticated Routes Group ---
	authenticated := v1.Group("")               // Create a new group for all authenticated routes
	authenticated.Use(authMiddleware.JWTAuth()) // Apply JWT Auth middleware to this group
	{
		// Profile routes are handled by the UserController in bootstrap.go

		// Files endpoint (global access to file details)
		authenticated.GET("/files/:fileId", classNoteController.GetFileDetails)

		// Faculty routes (now under authenticated group)
		faculties := authenticated.Group("/faculties")
		{
			// All faculty routes now require authentication
			faculties.GET("", facultyController.GetAllFaculties)
			faculties.GET("/:id", facultyController.GetFacultyByID)

			// Role-protected routes within faculties
			facultiesInstructorProtected := faculties.Group("")
			// No need for JWTAuth() again, inherited from parent 'authenticated' group
			facultiesInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
			{
				facultiesInstructorProtected.POST("", facultyController.CreateFaculty)
				facultiesInstructorProtected.PUT("/:id", facultyController.UpdateFaculty)
				facultiesInstructorProtected.DELETE("/:id", facultyController.DeleteFaculty)
			}
		}

		// Department routes
		departments := authenticated.Group("/departments")
		{
			departments.GET("", departmentController.GetAllDepartments)
			departments.GET("/:id", departmentController.GetDepartmentByID)
			// departments.GET("/:id/instructors", departmentController.GetInstructorsByDepartment)

			// Role-protected routes within departments
			departmentsInstructorProtected := departments.Group("")
			departmentsInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
			{
				departmentsInstructorProtected.POST("", departmentController.CreateDepartment)
				departmentsInstructorProtected.PUT("/:id", departmentController.UpdateDepartment)
				departmentsInstructorProtected.DELETE("/:id", departmentController.DeleteDepartment)
			}
		}

		// Past Exam routes - Endpoints for accessing and managing past examination materials
		pastExams := authenticated.Group("/past-exams")
		{
			// Public routes accessible to all authenticated users (students and instructors)
			pastExams.GET("", pastExamController.GetAllPastExams)        // List all past exams with optional filtering
			pastExams.GET("/:id", pastExamController.GetPastExamByID)    // Retrieve a specific past exam by ID

			// Instructor-only routes - Protected by role-based middleware
			// These routes are restricted to users with the Instructor role
			pastExamsInstructorProtected := pastExams.Group("")
			pastExamsInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
			{
				// CRUD operations for past exam resources
				pastExamsInstructorProtected.POST("", pastExamController.CreatePastExam)                       // Create a new past exam
				pastExamsInstructorProtected.PUT("/:id", pastExamController.UpdatePastExam)                    // Update an existing past exam
				pastExamsInstructorProtected.DELETE("/:id", pastExamController.DeletePastExam)                 // Delete a past exam
				
				// File management for past exams
				pastExamsInstructorProtected.POST("/:id/files", pastExamController.AddFileToPastExam)          // Upload and attach files to a past exam
				pastExamsInstructorProtected.DELETE("/:id/files/:fileId", pastExamController.DeleteFileFromPastExam) // Remove a file from a past exam
			}
		}

		// Class Notes routes
		classNotes := authenticated.Group("/class-notes")
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
	}

	// Health check endpoint (public)
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, dto.APIResponse{
			Data: gin.H{"status": "ok"},
		})
	})

	// Swagger routes are set up in bootstrap.go already
}