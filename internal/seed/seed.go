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
func CreateDefaultData(ctx context.Context, dbPool *pgxpool.Pool, lgr zerolog.Logger) error {
	facultyRepo := appRepos.NewFacultyRepository(dbPool)
	departmentRepo := appRepos.NewDepartmentRepository(dbPool)
	userRepo := appRepos.NewUserRepository(dbPool)

	lgr.Info().Msg("Checking/Creating default data (Faculties/Departments)...")
	var finalErr error // To collect potential errors without stopping the process

	// --- Helper function to create department ---
	createDept := func(facultyID int64, name, code string) {
		dept := &appModels.Department{FacultyID: facultyID, Name: name, Code: code}
		err := departmentRepo.Create(ctx, dept)
		if err != nil && !errors.Is(err, apperrors.ErrDepartmentAlreadyExists) {
			lgr.Error().Err(err).Msgf("Error creating department: %s", name)
			finalErr = errors.Join(finalErr, err)
		}
	}

	// --- Helper function to create faculty and get its ID ---
	createFaculty := func(name, code string) int64 {
		faculty := &appModels.Faculty{Name: name, Code: code}
		id, err := facultyRepo.CreateFaculty(ctx, faculty)
		if err != nil && !errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
			lgr.Error().Err(err).Msgf("Error creating faculty: %s", name)
			finalErr = errors.Join(finalErr, err)
			return 0
		} else if errors.Is(err, apperrors.ErrFacultyAlreadyExists) {
			// Find existing ID if it already exists
			faculties, errGet := facultyRepo.GetAllFaculties(ctx)
			if errGet == nil {
				for _, f := range faculties {
					if f.Code == code {
						return f.ID
					}
				}
			} else {
				lgr.Error().Err(errGet).Msgf("Error getting existing faculties to find %s ID", code)
				finalErr = errors.Join(finalErr, errGet)
			}
		}
		return id
	}

	// --- Güzel Sanatlar ve Tasarım Fakültesi --- //
	gstfID := createFaculty("Güzel Sanatlar ve Tasarım Fakültesi", "GSTF")
	if gstfID > 0 {
		createDept(gstfID, "Film Tasarım ve Yönetimi (İngilizce)", "FTY")
		createDept(gstfID, "Yeni Medya ve İletişim (İngilizce)", "YMI")
		createDept(gstfID, "İç Mimarlık ve Çevre Tasarımı (İngilizce)", "ICM")
	}

	// --- Hukuk Fakültesi --- //
	createFaculty("Hukuk Fakültesi", "HUKUK") // No departments for this one as per request

	// --- Mühendislik ve Mimarlık Fakültesi --- //
	mmfID := createFaculty("Mühendislik ve Mimarlık Fakültesi", "MMF")
	if mmfID > 0 {
		createDept(mmfID, "Bilgisayar Mühendisliği (İngilizce)", "CENG")
		createDept(mmfID, "Bilgisayar Mühendisliği (Türkçe)", "CENG_TR")
		createDept(mmfID, "Bilişim Sistemleri Mühendisliği (İngilizce)", "ISE")
		createDept(mmfID, "Endüstri Mühendisliği (İngilizce)", "IE")
		createDept(mmfID, "Elektrik Elektronik Mühendisliği (İngilizce)", "EEE")
		createDept(mmfID, "Yazılım Mühendisliği (İngilizce)", "SE")
	}

	// --- İnsan ve Toplum Bilimleri Fakültesi --- //
	itbfID := createFaculty("İnsan ve Toplum Bilimleri Fakültesi", "ITBF")
	if itbfID > 0 {
		createDept(itbfID, "Psikoloji", "PSI")
		createDept(itbfID, "İngilizce Mütercim ve Tercümanlık", "IMT")
		createDept(itbfID, "İşletme", "ISL")
		createDept(itbfID, "Yönetim Bilişim Sistemleri", "YBS")
		createDept(itbfID, "Siyaset Bilimi ve Kamu Yönetimi Bölümü", "SBKY")
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
			if mmfID > 0 {
				departments, err := departmentRepo.GetByFacultyID(ctx, mmfID)
				if err == nil && len(departments) > 0 {
					for _, dept := range departments {
						if dept.Code == "CENG" {
							departmentID = dept.ID
							break
						}
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
