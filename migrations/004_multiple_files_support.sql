-- Migration: Multiple files support
-- Description: Adds support for multiple files per past exam and class note
-- Author: System
-- Date: 2024-06-19

BEGIN;

-- 0. Check if files table already exists and handle accordingly
DO $$
BEGIN
    -- Check if files table exists
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'files') THEN
        -- Check if resource_type column doesn't exist yet
        IF NOT EXISTS (SELECT FROM information_schema.columns 
                      WHERE table_schema = 'public' AND table_name = 'files' 
                      AND column_name = 'resource_type') THEN
            -- Add resource_type column
            ALTER TABLE files ADD COLUMN resource_type VARCHAR(50);
            -- Add resource_id column if it doesn't exist
            IF NOT EXISTS (SELECT FROM information_schema.columns 
                          WHERE table_schema = 'public' AND table_name = 'files' 
                          AND column_name = 'resource_id') THEN
                ALTER TABLE files ADD COLUMN resource_id BIGINT;
            END IF;
        END IF;
    ELSE
        -- 1. Creating main files table only if it doesn't exist
        CREATE TABLE files (
            id BIGSERIAL PRIMARY KEY,
            file_name VARCHAR(255) NOT NULL,
            file_path TEXT NOT NULL,
            file_url TEXT NOT NULL,
            file_size BIGINT NOT NULL,
            file_type VARCHAR(100) NOT NULL,  -- MIME type
            resource_type VARCHAR(50) NOT NULL,  -- PAST_EXAM, CLASS_NOTE
            resource_id BIGINT NOT NULL,
            uploaded_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END
$$;

-- 2. Creating past_exam_files junction table
CREATE TABLE IF NOT EXISTS past_exam_files (
    id BIGSERIAL PRIMARY KEY,
    past_exam_id BIGINT NOT NULL REFERENCES past_exams(id) ON DELETE CASCADE,
    file_id BIGINT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(past_exam_id, file_id)
);

-- 3. Creating class_note_files junction table
CREATE TABLE IF NOT EXISTS class_note_files (
    id BIGSERIAL PRIMARY KEY,
    class_note_id BIGINT NOT NULL REFERENCES class_notes(id) ON DELETE CASCADE,
    file_id BIGINT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(class_note_id, file_id)
);

-- 4. Create indexes to speed up common queries (IF NOT EXISTS is not supported for indexes in some Postgres versions)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_files_resource') THEN
        CREATE INDEX idx_files_resource ON files(resource_type, resource_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_past_exam_files_exam_id') THEN
        CREATE INDEX idx_past_exam_files_exam_id ON past_exam_files(past_exam_id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_class_note_files_note_id') THEN
        CREATE INDEX idx_class_note_files_note_id ON class_note_files(class_note_id);
    END IF;
END
$$;

-- 5. Add necessary function to update timestamps if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 6. Add trigger to update timestamps if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_files_updated_at') THEN
        CREATE TRIGGER update_files_updated_at
        BEFORE UPDATE ON files
        FOR EACH ROW
        EXECUTE FUNCTION update_updated_at_column();
    END IF;
END
$$;

-- 7. Migration de veriler: PastExam için dosya aktarımı
DO $$
DECLARE
    exam_record RECORD;
BEGIN
    -- Only run data migration if past_exam_files is empty
    IF (SELECT COUNT(*) FROM past_exam_files) = 0 THEN
        -- Past exam dosyalarını yeni sistem için kopyala
        FOR exam_record IN (SELECT id, file_url, instructor_id FROM past_exams WHERE file_url IS NOT NULL) LOOP
            -- Yeni dosya kaydı oluştur
            WITH new_file AS (
                INSERT INTO files (
                    file_name, 
                    file_path, 
                    file_url, 
                    file_size, 
                    file_type, 
                    resource_type, 
                    resource_id, 
                    uploaded_by
                ) VALUES (
                    'exam_file', -- Varsayılan dosya adı
                    exam_record.file_url, -- Dosya yolu
                    exam_record.file_url, -- Dosya URL'i
                    0, -- Varsayılan boyut
                    'application/pdf', -- Varsayılan tip
                    'PAST_EXAM',
                    exam_record.id,
                    exam_record.instructor_id -- Yükleyen ID'si (bu örnekte instructor ID kullanıldı)
                ) RETURNING id
            )
            -- Bağlantı tablosuna ekle
            INSERT INTO past_exam_files (past_exam_id, file_id)
            SELECT exam_record.id, id FROM new_file;
        END LOOP;
    END IF;
END;
$$;

-- 8. Migration de veriler: ClassNote için dosya aktarımı
DO $$
DECLARE
    note_record RECORD;
BEGIN
    -- Only run data migration if class_note_files is empty
    IF (SELECT COUNT(*) FROM class_note_files) = 0 THEN
        -- Check if image column exists in class_notes table
        IF EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_schema = 'public' AND table_name = 'class_notes' 
                  AND column_name = 'image') THEN
            -- Class note dosyalarını yeni sistem için kopyala
            FOR note_record IN (SELECT id, image, user_id FROM class_notes WHERE image IS NOT NULL) LOOP
                -- Yeni dosya kaydı oluştur
                WITH new_file AS (
                    INSERT INTO files (
                        file_name, 
                        file_path, 
                        file_url, 
                        file_size, 
                        file_type, 
                        resource_type, 
                        resource_id, 
                        uploaded_by
                    ) VALUES (
                        'note_image', -- Varsayılan dosya adı
                        note_record.image, -- Dosya yolu
                        note_record.image, -- Dosya URL'i
                        0, -- Varsayılan boyut
                        'image/jpeg', -- Varsayılan tip
                        'CLASS_NOTE',
                        note_record.id,
                        note_record.user_id -- Yükleyen ID'si 
                    ) RETURNING id
                )
                -- Bağlantı tablosuna ekle
                INSERT INTO class_note_files (class_note_id, file_id)
                SELECT note_record.id, id FROM new_file;
            END LOOP;
        END IF;
    END IF;
END;
$$;

ALTER TABLE past_exams DROP COLUMN file_url;
ALTER TABLE class_notes DROP COLUMN image;

COMMIT; 