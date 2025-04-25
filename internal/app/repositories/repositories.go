package repositories

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Veri erişim katmanı burada olacak
// Örnek:
// type UserRepository struct {
//     // DB bağlantısı
// }
//
// func NewUserRepository() *UserRepository {
//     return &UserRepository{}
// }
//
// func (r *UserRepository) FindAll() ([]models.User, error) {
//     // Veritabanından veri çekme işlemleri
// }

// Repositories holds all the repository instances
type Repositories struct {
	UserRepository       *UserRepository
	FacultyRepository    *FacultyRepository
	DepartmentRepository *DepartmentRepository
	TokenRepository      *TokenRepository
	PastExamRepository   *PastExamRepository
	ClassNoteRepository  *ClassNoteRepository
}

// NewRepositories initializes all repositories
func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepository:       NewUserRepository(db),
		FacultyRepository:    NewFacultyRepository(db),
		DepartmentRepository: NewDepartmentRepository(db),
		TokenRepository:      NewTokenRepository(db),
		PastExamRepository:   NewPastExamRepository(db),
		ClassNoteRepository:  NewClassNoteRepository(db),
	}
}
