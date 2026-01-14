-- Add RBAC tables: projects, project_members, casbin_rules
-- Generated manually for Story 5-5

-- Create projects table
CREATE TABLE IF NOT EXISTS "projects" (
    "id" bigserial NOT NULL PRIMARY KEY,
    "uuid" varchar(36) NOT NULL UNIQUE,
    "name" varchar(100) NOT NULL,
    "description" varchar(500),
    "owner_id" bigint NOT NULL,
    "status" smallint NOT NULL DEFAULT 1,
    "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "fk_projects_owner" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_projects_uuid" ON "projects" ("uuid");
CREATE INDEX IF NOT EXISTS "idx_projects_owner_id" ON "projects" ("owner_id");
CREATE INDEX IF NOT EXISTS "idx_projects_status" ON "projects" ("status");
CREATE INDEX IF NOT EXISTS "idx_projects_created_at" ON "projects" ("created_at");

-- Create project_members table
CREATE TABLE IF NOT EXISTS "project_members" (
    "id" bigserial NOT NULL PRIMARY KEY,
    "project_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "role" varchar(20) NOT NULL,
    "joined_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "fk_project_members_project" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON DELETE CASCADE,
    CONSTRAINT "fk_project_members_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_project_members_unique" ON "project_members" ("project_id", "user_id");
CREATE INDEX IF NOT EXISTS "idx_project_members_user_id" ON "project_members" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_project_members_role" ON "project_members" ("role");

-- Create casbin_rules table (for database-backed policy storage, optional)
CREATE TABLE IF NOT EXISTS "casbin_rules" (
    "id" bigserial NOT NULL PRIMARY KEY,
    "ptype" varchar(100) NOT NULL,
    "v0" varchar(100),
    "v1" varchar(100),
    "v2" varchar(100),
    "v3" varchar(100),
    "v4" varchar(100),
    "v5" varchar(100)
);

CREATE INDEX IF NOT EXISTS "idx_casbin_rules_ptype" ON "casbin_rules" ("ptype");
CREATE INDEX IF NOT EXISTS "idx_casbin_rules_v0" ON "casbin_rules" ("v0");
CREATE INDEX IF NOT EXISTS "idx_casbin_rules_v1" ON "casbin_rules" ("v1");

-- Add comment for traceability
COMMENT ON TABLE "projects" IS 'Story 5-5: RBAC project isolation domains';
COMMENT ON TABLE "project_members" IS 'Story 5-5: Project member roles for RBAC';
COMMENT ON TABLE "casbin_rules" IS 'Story 5-5: Optional database-backed Casbin policy storage';
