-- Add Community Feature Tables

-- Communities table
CREATE TABLE IF NOT EXISTS communities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    abbreviation VARCHAR(50) NOT NULL UNIQUE,
    lead_id BIGINT NOT NULL,
    profile_photo_file_id BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_communities_lead FOREIGN KEY (lead_id) REFERENCES users(id),
    CONSTRAINT fk_communities_profile_photo FOREIGN KEY (profile_photo_file_id) REFERENCES files(id) ON DELETE SET NULL
);

-- updated_at trigger for Communities
DROP TRIGGER IF EXISTS update_communities_updated_at ON communities;
CREATE TRIGGER update_communities_updated_at
    BEFORE UPDATE ON communities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Community participation table
CREATE TABLE IF NOT EXISTS community_participants (
    id BIGSERIAL PRIMARY KEY,
    community_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_community_participants_community FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
    CONSTRAINT fk_community_participants_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_community_participant UNIQUE(community_id, user_id)
);

-- Index for community participants
CREATE INDEX IF NOT EXISTS idx_community_participants_community_id ON community_participants(community_id);
CREATE INDEX IF NOT EXISTS idx_community_participants_user_id ON community_participants(user_id);