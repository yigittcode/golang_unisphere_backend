-- Create password reset tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expiry_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used BOOLEAN NOT NULL DEFAULT FALSE
);

-- Add index for faster token lookup
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);

-- Create functions for token management
CREATE OR REPLACE FUNCTION delete_expired_password_reset_tokens() RETURNS void AS $$
BEGIN
    DELETE FROM password_reset_tokens WHERE expiry_date < CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql;

-- Comment
COMMENT ON TABLE password_reset_tokens IS 'Stores password reset tokens for forgot password functionality'; 