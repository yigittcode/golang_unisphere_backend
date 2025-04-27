-- Migration to add cascade deletion for files
-- This will ensure that when a file is deleted, all references to it are removed as well

-- First, let's add ON DELETE CASCADE to users.profile_photo_file_id
DO $$
BEGIN
    -- Drop the existing foreign key constraint
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_user_profile_photo' AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users DROP CONSTRAINT fk_user_profile_photo;
    END IF;

    -- Add the new constraint with ON DELETE SET NULL
    ALTER TABLE users 
    ADD CONSTRAINT fk_user_profile_photo 
    FOREIGN KEY (profile_photo_file_id) 
    REFERENCES files(id) 
    ON DELETE SET NULL;
END$$;

-- For past_exam_files, ensure cascade delete works in both directions
-- When a past_exam is deleted, all its files should be deleted (already done with ON DELETE CASCADE)
-- When a file is deleted, the corresponding entry in past_exam_files should also be deleted
DO $$
BEGIN
    -- Drop the existing foreign key constraint for file_id
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_past_exam_files_file' AND conrelid = 'past_exam_files'::regclass
    ) THEN
        ALTER TABLE past_exam_files DROP CONSTRAINT fk_past_exam_files_file;
    END IF;

    -- Add the new constraint with ON DELETE CASCADE
    ALTER TABLE past_exam_files 
    ADD CONSTRAINT fk_past_exam_files_file 
    FOREIGN KEY (file_id) 
    REFERENCES files(id) 
    ON DELETE CASCADE;
END$$;

-- For class_note_files, ensure cascade delete works in both directions
-- When a class_note is deleted, all its files should be deleted (already done with ON DELETE CASCADE)
-- When a file is deleted, the corresponding entry in class_note_files should also be deleted
DO $$
BEGIN
    -- Drop the existing foreign key constraint for file_id
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_class_note_files_file' AND conrelid = 'class_note_files'::regclass
    ) THEN
        ALTER TABLE class_note_files DROP CONSTRAINT fk_class_note_files_file;
    END IF;

    -- Add the new constraint with ON DELETE CASCADE
    ALTER TABLE class_note_files 
    ADD CONSTRAINT fk_class_note_files_file 
    FOREIGN KEY (file_id) 
    REFERENCES files(id) 
    ON DELETE CASCADE;
END$$;

-- Create a function to delete physical files when a file record is deleted
CREATE OR REPLACE FUNCTION delete_physical_file() RETURNS TRIGGER AS $$
BEGIN
    -- This is just a placeholder. The actual file deletion is handled in the application code
    -- because the file storage location is configured in the application, not the database.
    -- This trigger is here for documentation purposes and potential future use.
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to run the function before deleting a file
DROP TRIGGER IF EXISTS before_delete_file ON files;
CREATE TRIGGER before_delete_file
BEFORE DELETE ON files
FOR EACH ROW
EXECUTE FUNCTION delete_physical_file();

-- Add comment to document the file deletion behavior
COMMENT ON TABLE files IS 'Stores file metadata. Physical files are deleted by the application when a record is deleted from this table.'; 