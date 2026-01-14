-- Migration: Add project_id to configitem for multi-tenant isolation
-- Story: Story 5.5 - RBAC Permissions
-- Date: 2026-01-14
-- Purpose: Add project_id field and indexes to support project-level data isolation

-- Add project_id column to configitem table
ALTER TABLE configitems 
ADD COLUMN IF NOT EXISTS project_id BIGINT;

-- Add comment
COMMENT ON COLUMN configitems.project_id IS '所属项目ID，用于多租户隔离（NULL表示全局配置）';

-- Create index for project_id
CREATE INDEX IF NOT EXISTS configitems_project_id_idx ON configitems(project_id);

-- Create composite index for project-level config queries
CREATE INDEX IF NOT EXISTS configitems_project_id_key_idx ON configitems(project_id, key);

-- Update atlas.sum metadata
-- This migration adds project_id support to enable RBAC-based project isolation
