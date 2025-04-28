-- UniSphere Veritabanı Tam Şema - Birleştirilmiş Migration

-- ENUM türlerini oluştur
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_type') THEN
        CREATE TYPE role_type AS ENUM ('STUDENT', 'INSTRUCTOR');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'term_type') THEN
        CREATE TYPE term_type AS ENUM ('FALL', 'SPRING');
    END IF;
END$$;

-- updated_at kolonunu güncelleyen fonksiyon
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- Ana Tablolar --

-- Fakülte tablosu
CREATE TABLE IF NOT EXISTS faculties (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(20) UNIQUE,
    description TEXT
);

-- Bölüm tablosu
CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    faculty_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(20) NOT NULL,
    CONSTRAINT fk_faculty
        FOREIGN KEY (faculty_id) REFERENCES faculties(id),
    CONSTRAINT unique_department_code UNIQUE (code)
);

-- Kullanıcılar tablosu
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
    department_id BIGINT NULL,
    profile_photo_file_id BIGINT NULL,
    CONSTRAINT fk_user_department FOREIGN KEY (department_id) REFERENCES departments(id)
);

-- Users için updated_at trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Öğrenci tablosu
CREATE TABLE IF NOT EXISTS students (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    identifier VARCHAR(20) NOT NULL UNIQUE,
    graduation_year INT,
    CONSTRAINT fk_student_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Öğretim üyesi tablosu
CREATE TABLE IF NOT EXISTS instructors (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    title VARCHAR(100) NOT NULL,
    CONSTRAINT fk_instructor_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Geçmiş sınavlar tablosu
CREATE TABLE IF NOT EXISTS past_exams (
    id BIGSERIAL PRIMARY KEY,
    year INT NOT NULL,
    term term_type NOT NULL,
    department_id BIGINT NOT NULL,
    course_code VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    instructor_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_past_exams_department
        FOREIGN KEY (department_id) REFERENCES departments(id),
    CONSTRAINT fk_past_exams_instructor
        FOREIGN KEY (instructor_id) REFERENCES instructors(id)
);

-- Past exams için updated_at trigger
DROP TRIGGER IF EXISTS update_past_exams_updated_at ON past_exams;
CREATE TRIGGER update_past_exams_updated_at
    BEFORE UPDATE ON past_exams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Ders notları tablosu
CREATE TABLE IF NOT EXISTS class_notes (
    id BIGSERIAL PRIMARY KEY,
    year INT NOT NULL,
    term term_type NOT NULL,
    department_id BIGINT NOT NULL,
    course_code VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_class_notes_department
        FOREIGN KEY (department_id) REFERENCES departments(id),
    CONSTRAINT fk_class_notes_user
        FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Class notes için updated_at trigger
DROP TRIGGER IF EXISTS update_class_notes_updated_at ON class_notes;
CREATE TRIGGER update_class_notes_updated_at
    BEFORE UPDATE ON class_notes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Refresh Token tablosu
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

-- Dersler tablosu
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

-- Ders sunumları tablosu
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

-- Dosya Tabloları --

-- Dosyalar için ana tablo - ÖNEMLİ: file_name yerine filename kullanıldı
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,  -- Kodu değiştirmemek için file_name kullanıldı (önceki filename)
    file_path TEXT NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(100) NOT NULL,  -- MIME type
    resource_type VARCHAR(50),        -- PAST_EXAM, CLASS_NOTE, USER gibi
    resource_id BIGINT,               -- İlgili kaynağın ID'si
    uploaded_by BIGINT NOT NULL,      -- Yükleyen kullanıcı ID'si
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_files_uploaded_by FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Files için updated_at trigger
DROP TRIGGER IF EXISTS update_files_updated_at ON files;
CREATE TRIGGER update_files_updated_at
    BEFORE UPDATE ON files
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Geçmiş sınav dosyaları bağlantı tablosu
CREATE TABLE IF NOT EXISTS past_exam_files (
    id BIGSERIAL PRIMARY KEY,
    past_exam_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_past_exam_files_past_exam FOREIGN KEY (past_exam_id) REFERENCES past_exams(id) ON DELETE CASCADE,
    CONSTRAINT fk_past_exam_files_file FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    CONSTRAINT unique_past_exam_file UNIQUE(past_exam_id, file_id)
);

-- Ders notu dosyaları bağlantı tablosu
CREATE TABLE IF NOT EXISTS class_note_files (
    id BIGSERIAL PRIMARY KEY,
    class_note_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_class_note_files_class_note FOREIGN KEY (class_note_id) REFERENCES class_notes(id) ON DELETE CASCADE,
    CONSTRAINT fk_class_note_files_file FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    CONSTRAINT unique_class_note_file UNIQUE(class_note_id, file_id)
);

-- Kullanıcı profil fotoğrafı için foreign key
ALTER TABLE users 
    ADD CONSTRAINT fk_user_profile_photo 
    FOREIGN KEY (profile_photo_file_id) 
    REFERENCES files(id) ON DELETE SET NULL;

-- Performans için İndeksler
CREATE INDEX IF NOT EXISTS idx_class_notes_user ON class_notes(user_id);
CREATE INDEX IF NOT EXISTS idx_past_exams_instructor ON past_exams(instructor_id);
CREATE INDEX IF NOT EXISTS idx_departments_faculty ON departments(faculty_id);
CREATE INDEX IF NOT EXISTS idx_class_notes_course ON class_notes(course_code);
CREATE INDEX IF NOT EXISTS idx_past_exams_course ON past_exams(course_code);
CREATE INDEX IF NOT EXISTS idx_users_department ON users(department_id);
CREATE INDEX IF NOT EXISTS idx_files_resource ON files(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_past_exam_files_exam_id ON past_exam_files(past_exam_id);
CREATE INDEX IF NOT EXISTS idx_class_note_files_note_id ON class_note_files(class_note_id);
CREATE INDEX IF NOT EXISTS idx_users_profile_photo_file_id ON users(profile_photo_file_id);