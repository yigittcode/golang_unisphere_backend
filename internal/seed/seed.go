package seed

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	appModels "github.com/yigit/unisphere/internal/app/models"
	appRepos "github.com/yigit/unisphere/internal/app/repositories"
)

// CreateDefaultData creates default faculties and departments if they don't exist.
// Moved from bootstrap package.
func CreateDefaultData(ctx context.Context, dbPool *pgxpool.Pool, lgr zerolog.Logger) error {
	facultyRepo := appRepos.NewFacultyRepository(dbPool)
	departmentRepo := appRepos.NewDepartmentRepository(dbPool)

	lgr.Info().Msg("Checking/Creating default data (Faculties/Departments)...")
	var finalErr error // To collect potential errors without stopping the process

	// --- Engineering Faculty & Departments --- //
	engineeringFaculty := &appModels.Faculty{Name: "Engineering Faculty", Code: "ENG"}
	engineeringID, err := facultyRepo.CreateFaculty(ctx, engineeringFaculty)
	if err != nil && !errors.Is(err, appRepos.ErrFacultyAlreadyExists) {
		lgr.Error().Err(err).Msg("Error creating engineering faculty")
		finalErr = errors.Join(finalErr, err)
	} else if errors.Is(err, appRepos.ErrFacultyAlreadyExists) {
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
		if err != nil && !errors.Is(err, appRepos.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating computer engineering department")
			finalErr = errors.Join(finalErr, err)
		}
		// Create Electrical Engineering
		eeeDept := &appModels.Department{FacultyID: engineeringID, Name: "Electrical Engineering", Code: "EEE"}
		err = departmentRepo.Create(ctx, eeeDept)
		if err != nil && !errors.Is(err, appRepos.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating electrical engineering department")
			finalErr = errors.Join(finalErr, err)
		}
	}

	// --- Science Faculty & Departments --- //
	scienceFaculty := &appModels.Faculty{Name: "Science Faculty", Code: "SCI"}
	scienceID, err := facultyRepo.CreateFaculty(ctx, scienceFaculty)
	if err != nil && !errors.Is(err, appRepos.ErrFacultyAlreadyExists) {
		lgr.Error().Err(err).Msg("Error creating science faculty")
		finalErr = errors.Join(finalErr, err)
	} else if errors.Is(err, appRepos.ErrFacultyAlreadyExists) {
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
		if err != nil && !errors.Is(err, appRepos.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating mathematics department")
			finalErr = errors.Join(finalErr, err)
		}
		// Create Physics
		physDept := &appModels.Department{FacultyID: scienceID, Name: "Physics", Code: "PHYS"}
		err = departmentRepo.Create(ctx, physDept)
		if err != nil && !errors.Is(err, appRepos.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msg("Error creating physics department")
			finalErr = errors.Join(finalErr, err)
		}
	}

	lgr.Info().Msg("Default data check/creation finished.")
	return finalErr // Return collected errors, if any
}
