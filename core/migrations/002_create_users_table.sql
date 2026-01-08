-- Migration: Create users table for authentication
-- Story: Story 18 - User Registration
-- Date: 2026-01-08
-- Purpose: Create users table with authentication, profile, and audit fields

-- Drop existing users table if it exists (clean slate)
DROP TABLE IF EXISTS users CASCADE;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    -- Primary Key
    id BIGSERIAL PRIMARY KEY,

    -- Authentication Fields
    username VARCHAR(64) UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,

    -- Profile Fields
    nickname VARCHAR(64),
    avatar VARCHAR(255),
    phone VARCHAR(20),
    gender SMALLINT DEFAULT 0,
    signature VARCHAR(255),

    -- Account Status
    status SMALLINT DEFAULT 1,

    -- Login History
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),

    -- Localization
    timezone VARCHAR(64) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'zh-CN',

    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT users_email_check CHECK (email <> ''),
    CONSTRAINT users_password_hash_check CHECK (password_hash <> ''),
    CONSTRAINT users_gender_check CHECK (gender IN (0, 1, 2)),
    CONSTRAINT users_status_check CHECK (status IN (0, 1))
);

-- Create indexes
CREATE INDEX IF NOT EXISTS users_status_idx ON users(status);
CREATE INDEX IF NOT EXISTS users_created_at_idx ON users(created_at);

-- Add comments
COMMENT ON TABLE users IS 'User accounts for authentication and authorization';
COMMENT ON COLUMN users.id IS 'User ID';
COMMENT ON COLUMN users.username IS 'Username for login (optional, alphanumeric + underscore)';
COMMENT ON COLUMN users.email IS 'Email address for login (required, unique)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt password hash (cost=12)';
COMMENT ON COLUMN users.nickname IS 'Display name / nickname';
COMMENT ON COLUMN users.avatar IS 'Avatar URL';
COMMENT ON COLUMN users.phone IS 'Phone number';
COMMENT ON COLUMN users.gender IS 'Gender: 0-Unknown, 1-Male, 2-Female';
COMMENT ON COLUMN users.signature IS 'User signature / bio';
COMMENT ON COLUMN users.status IS 'Account status: 0-Disabled, 1-Active';
COMMENT ON COLUMN users.last_login_at IS 'Last login timestamp';
COMMENT ON COLUMN users.last_login_ip IS 'Last login IP address (supports IPv6)';
COMMENT ON COLUMN users.timezone IS 'User timezone (e.g., Asia/Shanghai)';
COMMENT ON COLUMN users.language IS 'Preferred language (e.g., zh-CN, en-US)';
COMMENT ON COLUMN users.created_at IS 'Account creation timestamp';
COMMENT ON COLUMN users.updated_at IS 'Last update timestamp';

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Re-create foreign key for servers table
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'servers') THEN
        -- Add foreign key if servers table exists
        ALTER TABLE servers
        ADD COLUMN IF NOT EXISTS user_users BIGINT,
        ADD CONSTRAINT servers_users_fkey
        FOREIGN KEY (user_users) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Summary
SELECT 'Migration 002: Users table created successfully!' AS status;
