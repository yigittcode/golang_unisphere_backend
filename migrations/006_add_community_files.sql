-- Add community_files table for properly tracking community files

-- Create the community_files association table
CREATE TABLE IF NOT EXISTS community_files (
    id BIGSERIAL PRIMARY KEY,
    community_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_community_files_community FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
    CONSTRAINT fk_community_files_file FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    CONSTRAINT unique_community_file UNIQUE(community_id, file_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_community_files_community_id ON community_files(community_id);
CREATE INDEX IF NOT EXISTS idx_community_files_file_id ON community_files(file_id);

-- Migrate existing files with COMMUNITY resource type to the community_files table
INSERT INTO community_files (community_id, file_id)
SELECT resource_id, id
FROM files
WHERE resource_type = 'COMMUNITY'
ON CONFLICT DO NOTHING;