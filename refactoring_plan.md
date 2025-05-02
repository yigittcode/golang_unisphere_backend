# Refactoring Plan for Unisphere Backend

## Duplicate Profile Endpoints
The application has duplicate profile endpoints between auth_controller.go and user_controller.go:

1. `AuthController` implements:
   - GET `/profile` → GetProfile
   - PUT `/profile` → UpdateProfile
   - POST `/profile/photo` → UpdateProfilePhoto
   - DELETE `/profile/photo` → DeleteProfilePhoto

2. `UserController` implements:
   - GET `/users/profile` → GetUserProfile
   - PUT `/users/profile` → UpdateUserProfile
   - POST `/users/profile/photo` → UpdateProfilePhoto

### Solution:
- In `routes.go`, we've already removed the `/profile` endpoints from auth_controller
- Keep `/users/profile` endpoints implemented by user_controller
- Consider adding the missing DeleteProfilePhoto endpoint to user_controller

## Redundant Student and Instructor Models

The application has redundant models and functionality spread across:

1. Redundant Models:
   - `models/student.go` - Will be removed
   - `models/instructor.go` - Will be removed 

2. Redundant Service:
   - `services/instructor_service.go` - Will be refactored to use User model directly 

3. User Repository methods:
   - Several methods in user_repository.go need to be modified to use combined data structure

### Solution:

1. Integrate Instructor/Student data into User model:
   ```go
   type User struct {
     // Existing fields...
     
     // Instructor fields
     Title string `json:"title,omitempty" db:"title"` // For instructors
     
     // Student fields
     Identifier string `json:"identifier,omitempty" db:"identifier"` // For students
     GraduationYear int `json:"graduationYear,omitempty" db:"graduation_year"` // For students
   }
   ```

2. Update database schema:
   - Migrate instructor and student table data to user table
   - Add title, identifier, and graduation_year fields to users table
   - Create migration to handle this change

3. Update UserService and UserController:
   - Add methods for instructor title update 
   - Add methods for student information
   - Move functionality from InstructorService

4. Update Repositories:
   - Modify UserRepository to handle all user operations
   - Remove references to Student and Instructor models

## Implementation Steps

1. ✅ Remove duplicate profile endpoints from routes.go
2. ✅ Modify User model to include fields from Student and Instructor
3. ✅ Create a database migration to add the new fields
4. ✅ Refactor UserRepository methods to work with the updated structure
5. ✅ Migrate InstructorService functionality to UserService
6. ✅ Update UserController with methods for role-specific operations
7. ✅ Add new endpoints for role-specific operations
8. Pending: Delete redundant files after migration is tested and verified:
   - `/internal/app/models/student.go`
   - `/internal/app/models/instructor.go`
   - `/internal/app/services/instructor_service.go`