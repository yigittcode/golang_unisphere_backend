-- Remove community_files table that's not needed

-- Drop the community_files table
DROP TABLE IF EXISTS community_files;

-- Note: We're keeping the existing file records in the files table
-- as they have the proper resource_type and resource_id set already