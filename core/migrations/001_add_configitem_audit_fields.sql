-- Migration: Add audit fields to configitems table
-- Story: Story 04 - Ent Schema 配置管理
-- Date: 2026-01-04

BEGIN;

-- Step 1: Add status column (nullable first)
ALTER TABLE configitems 
ADD COLUMN IF NOT EXISTS status VARCHAR(20);

-- Step 2: Set default value for existing rows
UPDATE configitems 
SET status = 'active' 
WHERE status IS NULL;

-- Step 3: Make status NOT NULL with default
ALTER TABLE configitems 
ALTER COLUMN status SET NOT NULL,
ALTER COLUMN status SET DEFAULT 'active';

-- Step 4: Add created_at column (nullable first)
ALTER TABLE configitems 
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP;

-- Step 5: Set default value for existing rows (use current timestamp)
UPDATE configitems 
SET created_at = NOW() 
WHERE created_at IS NULL;

-- Step 6: Make created_at NOT NULL with default
ALTER TABLE configitems 
ALTER COLUMN created_at SET NOT NULL,
ALTER COLUMN created_at SET DEFAULT NOW();

-- Step 7: Add updated_at column (nullable first)
ALTER TABLE configitems 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP;

-- Step 8: Set default value for existing rows
UPDATE configitems 
SET updated_at = NOW() 
WHERE updated_at IS NULL;

-- Step 9: Make updated_at NOT NULL with default
ALTER TABLE configitems 
ALTER COLUMN updated_at SET NOT NULL,
ALTER COLUMN updated_at SET DEFAULT NOW();

-- Step 10: Add status index
CREATE INDEX IF NOT EXISTS configitems_status 
ON configitems(status);

COMMIT;

-- Verify migration
SELECT 
    COUNT(*) as total_rows,
    COUNT(status) as has_status,
    COUNT(created_at) as has_created_at,
    COUNT(updated_at) as has_updated_at
FROM configitems;
