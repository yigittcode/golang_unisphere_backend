-- Add email verification to users table
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- Create email verification tokens table
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    expiry_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_email_verification_tokens_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index for faster token lookup
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token ON email_verification_tokens(token);

-- Email settings table for configuring SMTP
CREATE TABLE IF NOT EXISTS email_settings (
    id BIGSERIAL PRIMARY KEY,
    smtp_host VARCHAR(255) NOT NULL,
    smtp_port INT NOT NULL,
    smtp_username VARCHAR(255) NOT NULL,
    smtp_password VARCHAR(255) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    from_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- updated_at trigger for email_settings
DROP TRIGGER IF EXISTS update_email_settings_updated_at ON email_settings;
CREATE TRIGGER update_email_settings_updated_at
    BEFORE UPDATE ON email_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();