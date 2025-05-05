package seed

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	appModels "github.com/yigit/unisphere/internal/app/models"
	appRepos "github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"golang.org/x/crypto/bcrypt"
)

// CreateDefaultData creates default faculties and departments if they don't exist.
// Moved from bootstrap package.
func CreateDefaultData(ctx context.Context, dbPool *pgxpool.Pool, lgr zerolog.Logger) error {
	facultyRepo := appRepos.NewFacultyRepository(dbPool)
	departmentRepo := appRepos.NewDepartmentRepository(dbPool)
	userRepo := appRepos.NewUserRepository(dbPool)

	lgr.Info().Msg("Checking/Creating default data (Faculties/Departments)...")
	var finalErr error // To collect potential errors without stopping the process

	// --- Engineering Faculty & Departments --- //
	engineeringFaculty := &appModels.Faculty{Name: "Engineering Faculty", Code: "ENG"}
	engineeringID, err := facultyRepo.CreateFaculty(ctx, engineeringFaculty)
	if err != nil && !errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
		lgr.Error().Err(err).Msg("Error creating engineering faculty")
		finalErr = errors.Join(finalErr, err)
	} else if errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
		// Find existing ID if needed
		faculties, errGet := facultyRepo.GetAllFaculties(ctx)
		if errGet == nil {
			for _, f := range faculties {
				if f.Code == "ENG" {
					engineeringID = f.ID
					break
				}
			}
		} else {
			lgr.Error().Err(errGet).Msg("Error getting existing faculties to find ENG ID")
			finalErr = errors.Join(finalErr, errGet)
		}
	}

	if engineeringID > 0 {
		// Create Computer Engineering
		compEngDept := &appModels.Department{FacultyID: engineeringID, Name: "Computer Engineering", Code: "CENG"}
		err = departmentRepo.Create(ctx, compEngDept)
		if err != nil && !errors.Is(err, apperrors.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating computer engineering department")
			finalErr = errors.Join(finalErr, err)
		}
		// Create Electrical Engineering
		eeeDept := &appModels.Department{FacultyID: engineeringID, Name: "Electrical Engineering", Code: "EEE"}
		err = departmentRepo.Create(ctx, eeeDept)
		if err != nil && !errors.Is(err, apperrors.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating electrical engineering department")
			finalErr = errors.Join(finalErr, err)
		}
	}

	// --- Science Faculty & Departments --- //
	scienceFaculty := &appModels.Faculty{Name: "Science Faculty", Code: "SCI"}
	scienceID, err := facultyRepo.CreateFaculty(ctx, scienceFaculty)
	if err != nil && !errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
		lgr.Error().Err(err).Msg("Error creating science faculty")
		finalErr = errors.Join(finalErr, err)
	} else if errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
		// Find existing ID if needed
		faculties, errGet := facultyRepo.GetAllFaculties(ctx)
		if errGet == nil {
			for _, f := range faculties {
				if f.Code == "SCI" {
					scienceID = f.ID
					break
				}
			}
		} else {
			lgr.Error().Err(errGet).Msg("Error getting existing faculties to find SCI ID")
			finalErr = errors.Join(finalErr, errGet)
		}
	}

	if scienceID > 0 {
		// Create Mathematics
		mathDept := &appModels.Department{FacultyID: scienceID, Name: "Mathematics", Code: "MATH"}
		err = departmentRepo.Create(ctx, mathDept)
		if err != nil && !errors.Is(err, apperrors.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating mathematics department")
			finalErr = errors.Join(finalErr, err)
		}
		// Create Physics
		physDept := &appModels.Department{FacultyID: scienceID, Name: "Physics", Code: "PHYS"}
		err = departmentRepo.Create(ctx, physDept)
		if err != nil && !errors.Is(err, apperrors.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating physics department")
			finalErr = errors.Join(finalErr, err)
		}
	}

	// --- Create Default Admin User --- //
	// Check if admin user already exists
	exists, err := userRepo.EmailExists(ctx, "admin@unisphere.edu.tr")
	if err != nil {
		lgr.Error().Err(err).Msg("Error checking if admin user exists")
		finalErr = errors.Join(finalErr, err)
	} else if !exists {
		// Create admin user if not exists
		lgr.Info().Msg("Creating default admin user...")

		// Hash password for admin
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
		if err != nil {
			lgr.Error().Err(err).Msg("Error hashing admin password")
			finalErr = errors.Join(finalErr, err)
		} else {
			// Get a department ID for the admin (Computer Engineering if available)
			var departmentID int64
			departments, err := departmentRepo.GetByFacultyID(ctx, engineeringID)
			if err == nil && len(departments) > 0 {
				for _, dept := range departments {
					if dept.Code == "CENG" {
						departmentID = dept.ID
						break
					}
				}
			}

			if departmentID == 0 {
				// If CENG not found, get any department
				allDepts, err := departmentRepo.GetAll(ctx)
				if err == nil && len(allDepts) > 0 {
					departmentID = allDepts[0].ID
				}
			}

			// Create admin user
			if departmentID > 0 {
				admin := &appModels.User{
					Email:         "admin@unisphere.edu.tr",
					Password:      string(hashedPassword),
					FirstName:     "System",
					LastName:      "Administrator",
					RoleType:      appModels.RoleAdmin,
					IsActive:      true,
					EmailVerified: true,
					DepartmentID:  &departmentID,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}

				adminID, err := userRepo.CreateUser(ctx, admin)
				if err != nil {
					lgr.Error().Err(err).Msg("Error creating admin user")
					finalErr = errors.Join(finalErr, err)
				} else {
					lgr.Info().Int64("adminID", adminID).Msg("Default admin user created successfully")
				}
			} else {
				lgr.Error().Msg("No department found for admin user")
				finalErr = errors.Join(finalErr, errors.New("no department found for admin user"))
			}
		}
	} else {
		lgr.Info().Msg("Admin user already exists, skipping creation")
	}

	lgr.Info().Msg("Default data check/creation finished.")
	return finalErr // Return collected errors, if any
}
