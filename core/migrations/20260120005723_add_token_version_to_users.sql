-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "role" character varying NOT NULL DEFAULT 'platform_user', ADD COLUMN "is_system" boolean NOT NULL DEFAULT false, ADD COLUMN "token_version" bigint NOT NULL DEFAULT 0, ADD COLUMN "deleted_at" timestamptz NULL;
-- Create index "user_deleted_at" to table: "users"
CREATE INDEX "user_deleted_at" ON "users" ("deleted_at");
-- Create index "user_email_deleted_at" to table: "users"
CREATE INDEX "user_email_deleted_at" ON "users" ("email", "deleted_at");
-- Create index "user_email_nickname" to table: "users"
CREATE INDEX "user_email_nickname" ON "users" ("email", "nickname");
-- Create index "user_is_system" to table: "users"
CREATE INDEX "user_is_system" ON "users" ("is_system");
-- Create index "user_role" to table: "users"
CREATE INDEX "user_role" ON "users" ("role");
-- Create index "user_role_status" to table: "users"
CREATE INDEX "user_role_status" ON "users" ("role", "status");
-- Create index "user_token_version" to table: "users"
CREATE INDEX "user_token_version" ON "users" ("token_version");
