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
	authMiddleware *middleware.AuthMiddleware,
) {
	// API version group
	v1 := router.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		// Public accessible routes
		auth.POST("/register-student", authController.RegisterStudent)
		auth.POST("/register-instructor", authController.RegisterInstructor)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.RefreshToken)

		// Routes requiring authentication
		authProtected := auth.Group("")
		authProtected.Use(authMiddleware.JWTAuth())
		{
			authProtected.GET("/profile", authController.GetProfile)
		}
	}

	// Faculty routes
	faculties := v1.Group("/faculties")
	{
		// Public routes for viewing faculties
		faculties.GET("", facultyController.GetAllFaculties)
		faculties.GET("/:id", facultyController.GetFacultyByID)

		// Only instructors can create, update, and delete faculties
		facultiesProtected := faculties.Group("")
		facultiesProtected.Use(authMiddleware.JWTAuth())
		facultiesProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
		{
			facultiesProtected.POST("", facultyController.CreateFaculty)
			facultiesProtected.PUT("/:id", facultyController.UpdateFaculty)
			facultiesProtected.DELETE("/:id", facultyController.DeleteFaculty)
		}
	}

	// Faculty-departments route (separate from main faculty routes to avoid conflicts)
	v1.GET("/faculty-departments/:facultyId", departmentController.GetDepartmentsByFacultyID)

	// Department routes
	departments := v1.Group("/departments")
	{
		// Public routes for viewing departments
		departments.GET("", departmentController.GetAllDepartments)
		departments.GET("/:id", departmentController.GetDepartmentByID)

		// Only instructors can create, update, and delete departments
		departmentsProtected := departments.Group("")
		departmentsProtected.Use(authMiddleware.JWTAuth())
		departmentsProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
		{
			departmentsProtected.POST("", departmentController.CreateDepartment)
			departmentsProtected.PUT("/:id", departmentController.UpdateDepartment)
			departmentsProtected.DELETE("/:id", departmentController.DeleteDepartment)
		}
	}

	// Department-instructors route (separate from main department routes to avoid conflicts)
	v1.GET("/department-instructors/:departmentId", instructorController.GetInstructorsByDepartment)

	// Instructor routes
	instructors := v1.Group("/instructors")
	{
		instructors.GET("/:id", instructorController.GetInstructorByID)

		// Protected instructor routes
		instructorsProtected := instructors.Group("")
		instructorsProtected.Use(authMiddleware.JWTAuth())
		instructorsProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
		{
			instructorsProtected.GET("/profile", instructorController.GetInstructorProfile)
			instructorsProtected.PUT("/title", instructorController.UpdateTitle)
		}
	}

	// Past Exam routes
	pastExams := v1.Group("/pastexams")
	{
		// Public routes for viewing past exams
		pastExams.GET("", pastExamController.GetAllPastExams)
		pastExams.GET("/:id", pastExamController.GetPastExamByID)

		// Protected routes for creating, updating, deleting past exams
		pastExamsProtected := pastExams.Group("")
		pastExamsProtected.Use(authMiddleware.JWTAuth())
		{
			// Only instructors can create, update, and delete past exams
			instructorProtected := pastExamsProtected.Group("")
			instructorProtected.Use(authMiddleware.RoleRequired(string(models.RoleInstructor)))
			{
				// Create with validation
				createExamReq := &dto.CreatePastExamRequest{}
				instructorProtected.POST("", middleware.ValidateRequest(createExamReq), pastExamController.CreatePastExam)

				// Update with validation
				updateExamReq := &dto.UpdatePastExamRequest{}
				instructorProtected.PUT("/:id", middleware.ValidateRequest(updateExamReq), pastExamController.UpdatePastExam)

				instructorProtected.DELETE("/:id", pastExamController.DeletePastExam)
			}
		}
	}
}
