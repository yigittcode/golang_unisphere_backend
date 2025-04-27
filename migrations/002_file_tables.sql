-- Migration to add separate tables for file management
-- This addresses the issue where file_url is stored directly in past_exams and class_notes tables

-- Create a new table for storing files
CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,  -- Original filename (for reference)
    file_path VARCHAR(255) NOT NULL, -- Path relative to storage directory
    file_size BIGINT NOT NULL,       -- Size in bytes
    mime_type VARCHAR(100),          -- MIME type of the file (supports PDFs, images, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger for updated_at column
DROP TRIGGER IF EXISTS update_files_updated_at ON files;
CREATE TRIGGER update_files_updated_at
    BEFORE UPDATE ON files
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create table for past exam files (allows multiple files per exam)
CREATE TABLE IF NOT EXISTS past_exam_files (
    id BIGSERIAL PRIMARY KEY,
    past_exam_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE, -- Indicates if this is the primary file
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_past_exam_files_past_exam
        FOREIGN KEY (past_exam_id) REFERENCES past_exams(id) ON DELETE CASCADE,
    CONSTRAINT fk_past_exam_files_file
        FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Create an index for better query performance
CREATE INDEX IF NOT EXISTS idx_past_exam_files_past_exam_id ON past_exam_files(past_exam_id);

-- Create table for class note files (allows multiple files per note - both images and PDFs)
CREATE TABLE IF NOT EXISTS class_note_files (
    id BIGSERIAL PRIMARY KEY,
    class_note_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE, -- Indicates if this is the primary file
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_class_note_files_class_note
        FOREIGN KEY (class_note_id) REFERENCES class_notes(id) ON DELETE CASCADE,
    CONSTRAINT fk_class_note_files_file
        FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Create an index for better query performance
CREATE INDEX IF NOT EXISTS idx_class_note_files_class_note_id ON class_note_files(class_note_id);

-- Add profile_photo_file_id column to users table if it doesn't exist yet
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'profile_photo_file_id'
    ) THEN
        ALTER TABLE users ADD COLUMN profile_photo_file_id BIGINT NULL;
    END IF;
END$$;

-- Add foreign key constraint if it doesn't exist yet
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_user_profile_photo' AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users ADD CONSTRAINT fk_user_profile_photo FOREIGN KEY (profile_photo_file_id) REFERENCES files(id);
    END IF;
END$$;

-- Create an index for better query performance
CREATE INDEX IF NOT EXISTS idx_users_profile_photo_file_id ON users(profile_photo_file_id);

-- Eventually, we'll need to migrate data from the existing columns to these tables.
-- This can be done after implementing the functionality in the application code.

-- For now, keep the existing columns but consider them deprecated:
-- - past_exams.file_url
-- - class_notes.image
-- - users.profile_photo_url 