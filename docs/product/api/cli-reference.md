# AppRun CLI Reference

Complete command reference for the AppRun CLI tool.

---

## Global Flags

Available for all commands:

```bash
-c, --config string   Config file path (default: ./config/default.yaml)
-v, --verbose         Verbose output
-h, --help           Help for any command
    --version        Version information
```

---

## Commands Overview

### Server Commands (Local Operations)

Commands that operate on the local system, no authentication required:

- [`apprun configure`](#apprun-configure) - Configure CLI settings
- [`apprun serve`](#apprun-serve) - Start HTTP server
- [`apprun migrate`](#apprun-migrate) - Database migration management
- [`apprun generate`](#apprun-generate) - Generate code and configuration artifacts
- [`apprun version`](#apprun-version) - Display version information

### Client Commands (Remote Operations)

Commands for remote API operations (authentication required):

- [`apprun deploy`](#apprun-deploy) - Deploy to remote environment *(placeholder)*
- [`apprun logs`](#apprun-logs) - View remote logs *(placeholder)*
- [`apprun backup`](#apprun-backup) - Manage remote backups *(placeholder)*

> **Note**: Client command authentication and full implementation are provided in Story 1.6.2 (CLI-API Adapter).

---

## Command Reference

### apprun configure

Configure AppRun CLI user settings.

**Usage:**
```bash
apprun configure              # Interactive configuration wizard
apprun configure show         # Display current configuration
```

**Description:**  
Interactive wizard that configures user-level settings stored in `~/.apprun/config.yaml`. This includes:
- `endpoint`: API endpoint for remote operations
- `api_key`: Authentication key for client commands
- `config_path`: Application config file path for server commands

**Configuration File Location:**  
`~/.apprun/config.yaml` (permissions: 0600)

**Examples:**
```bash
# Run interactive configuration
apprun configure

# View current configuration
apprun configure show
```

**Subcommands:**
- `show` - Display current configuration (API key is masked)

---

### apprun serve

Start the AppRun HTTP server.

**Usage:**
```bash
apprun serve [flags]
```

**Description:**  
Starts the AppRun BaaS platform HTTP server with REST API and GraphQL endpoints. Initializes:
- Database connections
- Authentication services
- Configuration management
- Cache clients (Redis)
- RBAC enforcement
- HTTP routes and middleware

**Flags:**
```bash
-c, --config string   Config file path (default: ./config/default.yaml)
```

**Configuration:**  
Loads application configuration from:
1. `--config` parameter (if provided)
2. `CONFIG_DIR` environment variable
3. Default: `./config/default.yaml`

**Environment Variables:**  
Required environment variables (if config file not found):
- `DATABASE_HOST` - Database host
- `DATABASE_USER` - Database user
- `DATABASE_PASSWORD` - Database password
- `DATABASE_DB_NAME` - Database name
- `AUTH_JWT_SECRET` - JWT secret (32+ characters)

**Examples:**
```bash
# Start with default config
apprun serve

# Start with custom config
apprun serve --config ./config/production.yaml

# Start with environment variables
export DATABASE_HOST=localhost
export DATABASE_USER=postgres
export DATABASE_PASSWORD=secret
export DATABASE_DB_NAME=apprun
export AUTH_JWT_SECRET=your-32-char-jwt-secret-key
apprun serve
```

**Exit Codes:**
- `0` - Server stopped gracefully
- `1` - Configuration error or startup failure

---

### apprun migrate

Database migration management using Atlas SDK.

**Usage:**
```bash
apprun migrate <subcommand> [flags]
```

**Description:**  
Manage database schema migrations. Migrations are tracked in the `atlas_schema_revisions` table. All migration files are embedded in the binary for portability.

**Subcommands:**
- [`apply`](#apprun-migrate-apply) - Apply pending migrations
- [`status`](#apprun-migrate-status) - Check migration status
- [`validate`](#apprun-migrate-validate) - Validate migration files
- [`inspect`](#apprun-migrate-inspect) - Inspect database schema drift
- [`repair`](#apprun-migrate-repair) - Repair database schema drift

**Environment Variables:**  
Required (if config file not found):
- `DATABASE_HOST`
- `DATABASE_USER`
- `DATABASE_PASSWORD`
- `DATABASE_DB_NAME`

---

#### apprun migrate apply

Apply pending database migrations.

**Usage:**
```bash
apprun migrate apply [flags]
```

**Description:**  
Applies all pending migrations in order and updates the `atlas_schema_revisions` table.

**Flags:**
```bash
--dry-run   Preview SQL without executing (future feature)
```

**Process:**
1. Connect to database
2. Check for pending migrations
3. Apply migrations sequentially
4. Update revision table

**Examples:**
```bash
# Apply all pending migrations
apprun migrate apply

# Preview migrations (not yet implemented)
apprun migrate apply --dry-run
```

**Exit Codes:**
- `0` - Migrations applied successfully
- `1` - Migration failed or database connection error

---

#### apprun migrate status

Check current migration status.

**Usage:**
```bash
apprun migrate status
```

**Description:**  
Displays the current state of database migrations including:
- Current version
- Applied migrations
- Pending migrations
- Overall migration state

**Examples:**
```bash
apprun migrate status
```

**Output:**
```
📊 Migration Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Database is up to date

Current Version: 20240115_001
Applied Migrations: 5
Pending Migrations: 0

Applied:
  ✓ 20240101_initial
  ✓ 20240102_users
  ✓ 20240110_projects
  ✓ 20240112_rbac
  ✓ 20240115_storage
```

---

#### apprun migrate validate

Validate migration files integrity.

**Usage:**
```bash
apprun migrate validate
```

**Description:**  
Validates that migration files are correctly formatted and sequential. Checks:
- Migration files exist
- Filenames follow naming convention
- Version numbers are sequential
- No duplicate versions
- SQL syntax is valid

**Examples:**
```bash
apprun migrate validate
```

**Exit Codes:**
- `0` - All validations passed
- `1` - Validation errors found

---

#### apprun migrate inspect

Inspect database schema for drift detection.

**Usage:**
```bash
apprun migrate inspect
```

**Description:**  
Compares the actual database schema against the Ent schema definition to detect schema drift. This command checks for:
- Missing tables or columns
- Unexpected tables or columns
- Index inconsistencies
- Foreign key violations

Unlike `status`, which checks migration history, `inspect` examines the actual database structure.

**Examples:**
```bash
# Check for schema drift
apprun migrate inspect

# Normal output (no drift)
✅ Database schema is in sync with Ent schema

# Drift detected output
⚠️  Schema drift detected!

Differences between database and Ent schema:
-- Add column "email" to table "users"
ALTER TABLE `users` ADD COLUMN `email` varchar(255) NOT NULL;
```

**Exit Codes:**
- `0` - Schema is in sync
- `1` - Schema drift detected

---

#### apprun migrate repair

Repair database schema drift using declarative migrations.

**Usage:**
```bash
apprun migrate repair [flags]
```

**Description:**  
Automatically repairs schema drift by generating and applying SQL to align the database with the Ent schema. This uses declarative migrations that directly compare database state with schema definitions.

**Security Features:**
- Default dry-run mode (preview only)
- Production environment disabled
- Requires explicit `--execute` flag for modifications
- Follows diff policy (no destructive operations)

**Flags:**
```bash
--execute   Execute the repair SQL (required for actual changes)
```

**Examples:**
```bash
# Preview repair SQL (dry-run)
apprun migrate repair

# Execute repair
apprun migrate repair --execute

# Production environment (blocked)
APP_ENV=production apprun migrate repair
# Error: 'migrate repair' is not allowed in production
```

**Process:**
1. Inspect schema drift
2. Generate repair SQL
3. Display changes (dry-run) or apply (execute)
4. Verify repair success

**Exit Codes:**
- `0` - Repair completed successfully
- `1` - Repair failed or schema drift detected

---

## Migration Modes and Best Practices

AppRun supports two migration modes: **Versioned Migrations** (recommended for production) and **Declarative Migrations** (for development fixes).

### Migration Commands Overview

| Command | Mode | Purpose | Generates Files | Production Safe |
|---------|------|---------|----------------|-----------------|
| `diff` | Versioned | Generate migration files | ✅ .sql | ✅ |
| `apply` | Versioned | Apply migration files | ❌ | ✅ |
| `sync` | Versioned | Auto generate + apply | ✅ .sql | ⚠️ Use cautiously |
| `status` | Versioned | Check migration history | ❌ | ✅ |
| `rollback` | Versioned | Rollback migrations | ❌ | ✅ |
| `reset` | Versioned | Clear database | ❌ | ❌ |
| `inspect` | Declarative | Check schema drift | ❌ | ✅ (read-only) |
| `repair` | Declarative | Fix schema drift | ❌ | ❌ |

### Recommended Workflows

#### Development Environment
```bash
# Normal development
./core/bin/apprun migrate sync

# Database corruption fix
./core/bin/apprun migrate inspect
./core/bin/apprun migrate repair --execute

# Regular checks
./core/bin/apprun migrate inspect
```

#### Production Environment
```bash
# Only use versioned migrations
./core/bin/apprun migrate apply

# Read-only checks
./core/bin/apprun migrate inspect
```

### Migration Maintenance

#### File Naming Conventions
Use descriptive names for migrations:
```bash
# ✅ Good
./core/bin/apprun migrate diff add_user_email_column
./core/bin/apprun migrate diff create_audit_log_table

# ❌ Bad
./core/bin/apprun migrate diff update
./core/bin/apprun migrate diff fix
```

#### Handling Large Migration Counts
When `core/migrations/` grows large (>100 files):

**Archive old migrations:**
```bash
mkdir -p core/migrations/archive/2025-q1/
mv core/migrations/202501*.sql core/migrations/archive/2025-q1/
cd core && atlas migrate hash --dir file://migrations/
```

**Squash development migrations:**
```bash
# Only for unreleased migrations
./core/bin/apprun migrate reset
rm core/migrations/*feature*.sql
./core/bin/apprun migrate diff consolidated_feature
./core/bin/apprun migrate apply
```

### Troubleshooting

#### "no schema changes detected"
**Cause:** Atlas compares schema against migration directory after reset.
**Solution:** Use `sync` which handles empty databases correctly.

#### "checksum mismatch"
**Cause:** Migration files modified after application.
**Solutions:**
- Restore original files: `git checkout core/migrations/*.sql`
- Use repair: `./core/bin/apprun migrate repair --execute`

#### Schema drift detected
**Cause:** Manual database changes or failed migrations.
**Solutions:**
- Quick fix: `./core/bin/apprun migrate repair --execute`
- Production: `./core/bin/apprun migrate diff fix_drift`

### Security Guidelines

#### ✅ Safe Operations
```bash
# Always safe
./core/bin/apprun migrate status
./core/bin/apprun migrate inspect
./core/bin/apprun migrate validate
```

#### ⚠️ Use with Caution
```bash
# Review before production
./core/bin/apprun migrate apply
./core/bin/apprun migrate rollback
```

#### ❌ Prohibited in Production
```bash
APP_ENV=production ./core/bin/apprun migrate reset     # ❌
APP_ENV=production ./core/bin/apprun migrate repair    # ❌
APP_ENV=production ./core/bin/apprun migrate sync      # ❌
```

### Makefile Shortcuts

```bash
# Versioned migrations
make db-sync              # Auto sync (development)
make db-diff NAME=xxx     # Generate migration
make db-migrate           # Apply migrations
make db-status            # Check status
make db-rollback          # Rollback

# Declarative migrations
make db-inspect           # Check schema drift
make db-repair            # Preview repair
make db-repair-execute    # Execute repair
```

---

### apprun generate

Generate code, configuration, and documentation artifacts.

**Usage:**
```bash
apprun generate [subcommand] [flags]
```

**Description:**  
Unified command for generating various project artifacts including configuration files, ORM models, and API documentation. Eliminates the need for manual script execution and provides a consistent interface for all generation tasks.

**Available Subcommands:**

#### apprun generate config

Generate configuration example files.

**Usage:**
```bash
apprun generate config [flags]
```

**Description:**  
Generates `config.example` (YAML) and `.env.example` files by introspecting module configurations. Automatically discovers all registered modules and extracts:
- Default values from struct tags
- Validation rules
- Database persistence flags
- Environment-only operational switches

**Flags:**
```bash
-o, --output string   Output directory (default: ".")
    --config-only     Generate config.example only
    --env-only        Generate .env.example only
```

**Examples:**
```bash
# Generate both files
apprun generate config

# Generate config.example only
apprun generate config --config-only

# Custom output directory
apprun generate config --output /path/to/output
```

**Output Files:**
- `config/config.example` - YAML configuration template (105 lines, 6 modules)
- `.env.example` - Environment variables template (112 lines, 7 modules)

**Equivalent Makefile:**
```bash
make config-example
```

---

#### apprun generate model

Generate Ent ORM code from schema definitions.

**Usage:**
```bash
apprun generate model
```

**Description:**  
Generates type-safe Go code from Ent schema definitions located in `ent/schema/`. Creates:
- Entity structs and interfaces
- Database query builders
- Migration files (with Atlas integration)
- CRUD operations with compile-time safety

**Features Enabled:**
- **VersionedMigration**: Atlas integration for schema versioning
- **Privacy**: Access control and authorization hooks
- **Upsert**: Insert-or-update operations

**Examples:**
```bash
# Generate Ent ORM code
apprun generate model
```

**Output:**
```
🔄 Generating Ent ORM code...
✅ Ent code generated successfully
💡 Generated files in ent/ directory
```

**Equivalent Makefile:**
```bash
make generate
```

**See Also:**
- Ent Documentation: https://entgo.io/docs/code-gen
- Schema Guide: `ent/schema/` directory

---

#### apprun generate openapi

Generate OpenAPI/Swagger API documentation.

**Usage:**
```bash
apprun generate openapi [flags]
```

**Description:**  
Generates OpenAPI 2.0 (Swagger) documentation from Go code annotations in handler files. Scans your codebase for Swagger annotations and produces interactive API documentation.

**Flags:**
```bash
-o, --output string   Output directory (default: "docs")
-m, --main string     Main file for API annotations (default: "internal/bootstrap/server.go")
```

**Generated Files:**
- `docs/swagger.json` - OpenAPI specification (JSON format, ~69KB)
- `docs/swagger.yaml` - OpenAPI specification (YAML format, ~34KB)
- `docs/docs.go` - Go code for serving documentation (~70KB)

**Examples:**
```bash
# Generate OpenAPI documentation
apprun generate openapi

# Custom output directory
apprun generate openapi --output api-docs

# Custom main file
apprun generate openapi --main cmd/api/main.go
```

**Output:**
```
📚 Generating OpenAPI/Swagger documentation...
✅ OpenAPI documentation generated in docs/
📄 Files created:
   - docs/swagger.json
   - docs/swagger.yaml
   - docs/docs.go
💡 Access at: http://localhost:8080/api/docs/
```

**Annotation Example:**
```go
// @Summary      Get user by ID
// @Description  Retrieve user details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      404  {object}  response.Response
// @Router       /users/{id} [get]
func GetUser(c *gin.Context) {
    // handler implementation
}
```

**Equivalent Makefile:**
```bash
make swagger
# or
make docs-api
```

**See Also:**
- Swag Documentation: https://github.com/swaggo/swag
- Annotation Reference: https://github.com/swaggo/swag#declarative-comments-format

---

**Generate Command Summary:**

| Subcommand | Purpose | Output Files | Equivalent Make |
|------------|---------|--------------|-----------------|
| `config` | Configuration examples | config.example, .env.example | `make config-example` |
| `model` | Ent ORM code | ent/*.go | `make generate` |
| `openapi` | API documentation | docs/swagger.* | `make swagger` |

**Benefits:**
- ✅ Unified CLI interface for all generation tasks
- ✅ No need to remember different tools (go generate, swag, scripts)
- ✅ Consistent flags and output format
- ✅ Help documentation built-in (`--help`)
- ✅ Works without separate tool installation (uses compiled binary)

---

### apprun version

Display version information.

**Usage:**
```bash
apprun version
```

**Description:**  
Shows detailed version information including:
- Version number
- Git commit hash
- Build timestamp
- Go version
- Platform/architecture

**Examples:**
```bash
apprun version
```

**Output:**
```
AppRun BaaS Platform
Version:    v1.0.0
Git Commit: a1b2c3d
Build Time: 2026-01-15T10:30:00Z
Go Version: go1.23.5
Platform:   linux/amd64
```

---

### apprun deploy

*(Placeholder - Full implementation in Story 1.6.2)*

Deploy application to remote AppRun environment.

**Usage:**
```bash
apprun deploy [flags]
```

**Status:**  
Not yet implemented. Currently displays a placeholder message.

**Planned Features:**
- Deploy application to remote AppRun instance
- Authentication via API key from `~/.apprun/config.yaml`
- Support for multiple environments (dev, staging, prod)
- Deployment confirmation and rollback

**Examples:**
```bash
apprun deploy
apprun deploy --environment prod
```

---

### apprun logs

*(Placeholder - Full implementation in Story 1.6.2)*

View logs from remote AppRun application.

**Usage:**
```bash
apprun logs [flags]
```

**Status:**  
Not yet implemented. Currently displays a placeholder message.

**Planned Features:**
- View application logs from remote instance
- Real-time log streaming (`--follow`)
- Filter logs by level, module, or pattern
- Authentication via API key

**Examples:**
```bash
apprun logs
apprun logs --follow
apprun logs --lines 100
```

---

### apprun backup

*(Placeholder - Full implementation in Story 1.6.2)*

Manage backups for remote AppRun application.

**Usage:**
```bash
apprun backup [subcommand]
```

**Status:**  
Not yet implemented. Currently displays a placeholder message.

**Planned Subcommands:**
- `create` - Create a new backup
- `list` - List available backups
- `restore <backup-id>` - Restore from backup
- `delete <backup-id>` - Delete a backup

**Examples:**
```bash
apprun backup create
apprun backup list
apprun backup restore backup-20260115-1030
```

---

## Configuration

### User Configuration File

**Location:** `~/.apprun/config.yaml`  
**Permissions:** `0600` (read/write for owner only)

**Format:**
```yaml
# Client CLI configuration (for remote operations)
endpoint: https://api.apprun.com  # API endpoint
api_key: your-api-key-here        # Authentication key

# Server CLI configuration (for local operations)
config_path: ./config/default.yaml  # Application config file path
```

**Creating Configuration:**
```bash
apprun configure
```

**Priority:**
1. `--config` command-line parameter
2. `~/.apprun/config.yaml` user config
3. Default values

---

### Application Configuration File

**Default Location:** `./config/default.yaml`  
**Override:** `--config` flag or `CONFIG_DIR` environment variable

Used by server commands (`serve`, `migrate`) for application settings including database, authentication, storage, etc.

See `config/config.example` for full configuration options.

---

## Backward Compatibility

### Legacy Binary

The legacy `bin/server` binary is maintained as a symbolic link:

```bash
bin/server -> apprun
```

**Usage:**
```bash
# Both commands are equivalent
./bin/server            # Legacy (symlink to apprun)
./bin/apprun serve      # New unified CLI
```

### Makefile Targets

```bash
make app-start          # Uses bin/apprun serve
make build              # Creates bin/apprun + symlink
```

---

## Exit Codes

Standard exit codes across all commands:

- `0` - Success
- `1` - General error (config, database, network, etc.)
- `2` - Command usage error (invalid flags, missing arguments)
- `130` - Interrupted by user (Ctrl+C)

---

## Troubleshooting

### Config File Not Found

**Error:**
```
❌ Configuration Error: Config file not found and essential environment variables missing
```

**Solutions:**

1. **Run from project root:**
   ```bash
   cd /path/to/apprun/core
   ./bin/apprun serve
   ```

2. **Specify config directory:**
   ```bash
   ./bin/apprun serve --config /path/to/config
   ```

3. **Set environment variables:**
   ```bash
   export DATABASE_HOST=localhost
   export DATABASE_USER=postgres
   export DATABASE_PASSWORD=secret
   export DATABASE_DB_NAME=apprun
   export AUTH_JWT_SECRET=your-32-char-jwt-secret-key
   ./bin/apprun serve
   ```

### Authentication Required

**Error:**
```
⚠️  Client command not yet implemented
Configure your API credentials: apprun configure
```

**Solution:**  
Client commands require authentication configuration. Full implementation available in Story 1.6.2.

### Database Connection Failed

**Error:**
```
failed to create migrator: connection refused
```

**Solution:**
1. Verify database is running
2. Check connection settings in config file
3. Verify environment variables are set correctly

---

## Shell Completion

*(Future feature - Phase 2)*

Generate shell completion scripts:

```bash
# Bash
apprun completion bash > /etc/bash_completion.d/apprun

# Zsh
apprun completion zsh > "${fpath[1]}/_apprun"

# Fish
apprun completion fish > ~/.config/fish/completions/apprun.fish
```

---

## Related Documentation

- [README.md](../README.md) - Project overview and quick start
- [Architecture](./architecture/) - System design documentation
- [Atlas Migration Modes](../architecture/ATLAS-MIGRATION-MODES.md) - Migration system technical details
- [Migration Inspect & Repair Guide](../architecture/MIGRATE-INSPECT-REPAIR.md) - Declarative migration usage
- [Migration Quick Reference](../architecture/MIGRATION-QUICK-REF.md) - Command reference and scenarios
- [Migration Maintenance Guide](../architecture/MIGRATIONS-MAINTENANCE.md) - Directory maintenance strategies
- [Story 1.6](./sprint-artifacts/sprint-0/1-6-unified-cli-architecture.md) - CLI architecture story
- [Story 1.6.2](./sprint-artifacts/sprint-2/1-6-2-cli-api-adapter.md) - CLI-API adapter (client commands)

---

**Last Updated:** 2026-01-16  
**Version:** Story 1.6 (Unified CLI Foundation with Migration Enhancements)
