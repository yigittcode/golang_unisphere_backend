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
	instructorController *controllers.InstructorController,
	pastExamController *controllers.PastExamController,
	classNoteController *controllers.ClassNoteController,
	authMiddleware *middleware.AuthMiddleware,
) {
	// API version group
	v1 := router.Group("/api/v1")

	// --- Public Auth routes ---
	auth := v1.Group("/auth")
	{
		auth.POST("/register-student", authController.RegisterStudent)
		auth.POST("/register-instructor", authController.RegisterInstructor)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)
		// Profile route moved to authenticated group
	}

	// --- Authenticated Routes Group ---
	authenticated := v1.Group("")               // Create a new group for all authenticated routes
	authenticated.Use(authMiddleware.JWTAuth()) // Apply JWT Auth middleware to this group
	{
		// Auth Profile route (moved here)
		authenticated.GET("/auth/profile", authController.GetCurrentUser)
		authenticated.PUT("/auth/profile", authController.UpdateProfile)
		authenticated.POST("/auth/profile/photo", authController.UpdateProfilePhoto)
		authenticated.DELETE("/auth/profile/photo", authController.DeleteProfilePhoto)

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

		// Faculty-departments route (now under authenticated group)
		authenticated.GET("/faculty-departments/:facultyId", departmentController.GetDepartmentsByFacultyID)

		// Department routes (now under authenticated group)
		departments := authenticated.Group("/departments")
		{
			departments.GET("", departmentController.GetAllDepartments)
			departments.GET("/:id", departmentController.GetDepartmentByID)

			departmentsInstructorProtected := departments.Group("")
			departmentsInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
			{
				departmentsInstructorProtected.POST("", departmentController.CreateDepartment)
				departmentsInstructorProtected.PUT("/:id", departmentController.UpdateDepartment)
				departmentsInstructorProtected.DELETE("/:id", departmentController.DeleteDepartment)
			}
		}

		// Department-instructors route (now under authenticated group)
		authenticated.GET("/department-instructors/:departmentId", instructorController.GetInstructorsByDepartment)

		// Instructor routes (now under authenticated group)
		instructors := authenticated.Group("/instructors")
		{
			instructors.GET("/:id", instructorController.GetInstructorByID)

			instructorsProtected := instructors.Group("")
			instructorsProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
			{
				instructorsProtected.GET("/profile", instructorController.GetInstructorProfile)
				instructorsProtected.PUT("/title", instructorController.UpdateTitle)
			}
		}

		// Past Exam routes (now under authenticated group)
		pastExams := authenticated.Group("/past-exams")
		{
			pastExams.GET("", pastExamController.GetAllPastExams)
			pastExams.GET("/:id", pastExamController.GetPastExamByID)

			// Role-protected routes within pastExams
			pastExamsInstructorProtected := pastExams.Group("")
			pastExamsInstructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
			{
				createExamReq := &dto.CreatePastExamRequest{}
				pastExamsInstructorProtected.POST("", middleware.ValidateRequest(createExamReq), pastExamController.CreatePastExam)

				updateExamReq := &dto.UpdatePastExamRequest{}
				pastExamsInstructorProtected.PUT("/:id", middleware.ValidateRequest(updateExamReq), pastExamController.UpdatePastExam)

				pastExamsInstructorProtected.DELETE("/:id", pastExamController.DeletePastExam)
			}
		}

		// Class Note routes (now under authenticated group)
		classNotes := authenticated.Group("/class-notes")
		{
			classNotes.GET("", classNoteController.GetAllNotes)
			classNotes.GET("/:noteId", classNoteController.GetNoteByID)

			// Form validasyonu için ValidateFormDataRequest kullanıyoruz (multipart/form-data)
			createNoteReq := &dto.CreateClassNoteRequest{}
			classNotes.POST("", middleware.ValidateFormDataRequest(createNoteReq), classNoteController.CreateNote)

			// JSON validasyonu için ValidateJsonRequest kullanıyoruz (application/json)
			updateNoteReq := &dto.UpdateClassNoteRequest{}
			classNotes.PUT("/:noteId", middleware.ValidateJsonRequest(updateNoteReq), classNoteController.UpdateNote)

			classNotes.DELETE("/:noteId", classNoteController.DeleteNote)

			// Add files to existing note
			classNotes.POST("/:noteId/files", classNoteController.AddFilesToNote)
		}
	}
}
