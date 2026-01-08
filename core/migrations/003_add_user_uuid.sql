-- Add UUID field to users table
-- Created: 2026-01-08
-- Story: Story 18 - User Registration Enhancement (UUID Support)

-- Add uuid column with default UUID generation
ALTER TABLE users 
ADD COLUMN uuid UUID NOT NULL DEFAULT gen_random_uuid();

-- Create unique index on uuid for fast lookups
CREATE UNIQUE INDEX users_uuid_key ON users(uuid);

-- Generate UUIDs for existing users (if any)
-- This ensures all existing records have valid UUIDs
UPDATE users SET uuid = gen_random_uuid() WHERE uuid IS NULL;

-- Make uuid immutable by removing default (new inserts must provide UUID)
-- Note: Ent will handle UUID generation in application code
ALTER TABLE users ALTER COLUMN uuid DROP DEFAULT;

-- Add comment for documentation
COMMENT ON COLUMN users.uuid IS 'Public UUID for external API reference (immutable, globally unique)';
