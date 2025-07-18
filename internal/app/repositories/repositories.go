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
	UserRepository                 *UserRepository
	FacultyRepository              *FacultyRepository
	DepartmentRepository           *DepartmentRepository
	TokenRepository                *TokenRepository
	VerificationTokenRepository    *VerificationTokenRepository
	PasswordResetTokenRepository   *PasswordResetTokenRepository
	PastExamRepository             *PastExamRepository
	ClassNoteRepository            *ClassNoteRepository
	FileRepository                 *FileRepository
	CommunityRepository            *CommunityRepository
	CommunityParticipantRepository *CommunityParticipantRepository
	ChatRepository                 *ChatRepository
}

// NewRepositories initializes all repositories
func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		UserRepository:                 NewUserRepository(db),
		FacultyRepository:              NewFacultyRepository(db),
		DepartmentRepository:           NewDepartmentRepository(db),
		TokenRepository:                NewTokenRepository(db),
		VerificationTokenRepository:    NewVerificationTokenRepository(db),
		PasswordResetTokenRepository:   NewPasswordResetTokenRepository(db),
		PastExamRepository:             NewPastExamRepository(db),
		ClassNoteRepository:            NewClassNoteRepository(db),
		FileRepository:                 NewFileRepository(db),
		CommunityRepository:            NewCommunityRepository(db),
		CommunityParticipantRepository: NewCommunityParticipantRepository(db),
		ChatRepository:                 NewChatRepository(db),
	}
}
