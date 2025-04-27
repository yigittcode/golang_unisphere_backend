-- Database structure for Unisphere API

-- Create ENUM types if they don't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_type') THEN
        CREATE TYPE role_type AS ENUM ('STUDENT', 'INSTRUCTOR');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'term_type') THEN
        CREATE TYPE term_type AS ENUM ('FALL', 'SPRING');
    END IF;
END$$;

-- Users table (common properties for both students and instructors)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    role_type role_type NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP NULL,
    department_id BIGINT NULL
);

-- Create a trigger to update the updated_at field
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- Drop trigger if exists and create it again
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Faculty table
CREATE TABLE IF NOT EXISTS faculties (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(20) UNIQUE,
    description TEXT
);

-- Department table
CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    faculty_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(20) NOT NULL,
    CONSTRAINT fk_faculty
        FOREIGN KEY (faculty_id) REFERENCES faculties(id),
    CONSTRAINT unique_department_code UNIQUE (code)
);

-- Add foreign key constraint for user-department relationship (only if it doesn't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_user_department' AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users ADD CONSTRAINT fk_user_department FOREIGN KEY (department_id) REFERENCES departments(id);
    END IF;
END$$;

-- Student table (extends User)
CREATE TABLE IF NOT EXISTS students (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    identifier VARCHAR(20) NOT NULL UNIQUE,
    department_id BIGINT NOT NULL,
    graduation_year INT,
    CONSTRAINT fk_student_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_student_department
        FOREIGN KEY (department_id) REFERENCES departments(id)
);

-- Instructor table (extends User)
CREATE TABLE IF NOT EXISTS instructors (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    department_id BIGINT NOT NULL,
    title VARCHAR(100) NOT NULL,
    CONSTRAINT fk_instructor_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_instructor_department
        FOREIGN KEY (department_id) REFERENCES departments(id)
);

-- Past Exams table
CREATE TABLE IF NOT EXISTS past_exams (
    id BIGSERIAL PRIMARY KEY,
    year INT NOT NULL,
    term term_type NOT NULL,
    department_id BIGINT NOT NULL,
    course_code VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    file_url VARCHAR(255),
    instructor_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_past_exams_department
        FOREIGN KEY (department_id) REFERENCES departments(id),
    CONSTRAINT fk_past_exams_instructor
        FOREIGN KEY (instructor_id) REFERENCES instructors(id)
);

-- Drop trigger if exists and create it again
DROP TRIGGER IF EXISTS update_past_exams_updated_at ON past_exams;
CREATE TRIGGER update_past_exams_updated_at
    BEFORE UPDATE ON past_exams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Class Notes table
CREATE TABLE IF NOT EXISTS class_notes (
    id BIGSERIAL PRIMARY KEY,
    year INT NOT NULL,
    term term_type NOT NULL,
    department_id BIGINT NOT NULL,
    course_code VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    image VARCHAR(255),
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_class_notes_department
        FOREIGN KEY (department_id) REFERENCES departments(id),
    CONSTRAINT fk_class_notes_user
        FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Drop trigger if exists and create it again
DROP TRIGGER IF EXISTS update_class_notes_updated_at ON class_notes;
CREATE TRIGGER update_class_notes_updated_at
    BEFORE UPDATE ON class_notes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Refresh Token table for authentication
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    expiry_date TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_refresh_tokens_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Courses table
CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    credits INT NOT NULL,
    CONSTRAINT fk_courses_department
        FOREIGN KEY (department_id) REFERENCES departments(id),
    CONSTRAINT unique_course_code UNIQUE (code)
);

-- Course Offerings table
CREATE TABLE IF NOT EXISTS course_offerings (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL,
    instructor_id BIGINT NOT NULL,
    year INT NOT NULL,
    term term_type NOT NULL,
    CONSTRAINT fk_course_offerings_course
        FOREIGN KEY (course_id) REFERENCES courses(id),
    CONSTRAINT fk_course_offerings_instructor
        FOREIGN KEY (instructor_id) REFERENCES instructors(id),
    CONSTRAINT unique_course_offering UNIQUE (course_id, instructor_id, year, term)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_class_notes_user ON class_notes(user_id);
CREATE INDEX IF NOT EXISTS idx_past_exams_instructor ON past_exams(instructor_id);
CREATE INDEX IF NOT EXISTS idx_students_department ON students(department_id);
CREATE INDEX IF NOT EXISTS idx_instructors_department ON instructors(department_id);
CREATE INDEX IF NOT EXISTS idx_departments_faculty ON departments(faculty_id);
CREATE INDEX IF NOT EXISTS idx_class_notes_course ON class_notes(course_code);
CREATE INDEX IF NOT EXISTS idx_past_exams_course ON past_exams(course_code);
CREATE INDEX IF NOT EXISTS idx_users_department ON users(department_id);

-- Örnek sorgular (yorum yapılmış):
/*
-- Fakülte ID'sine göre ders notlarını getirme sorgusu (JOIN kullanarak)
SELECT cn.*, d.name as department_name, f.name as faculty_name
FROM class_notes cn
JOIN departments d ON cn.department_id = d.id
JOIN faculties f ON d.faculty_id = f.id
WHERE f.id = $1
ORDER BY cn.created_at DESC
LIMIT $2 OFFSET $3;

-- Fakülte ID'sine göre geçmiş sınavları getirme sorgusu (JOIN kullanarak)
SELECT pe.*, d.name as department_name, f.name as faculty_name
FROM past_exams pe
JOIN departments d ON pe.department_id = d.id
JOIN faculties f ON d.faculty_id = f.id
WHERE f.id = $1
ORDER BY pe.created_at DESC
LIMIT $2 OFFSET $3;
*/ 