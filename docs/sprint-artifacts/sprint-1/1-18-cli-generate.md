# Story 1.18: CLI Generate - Unified Code Generation

**Priority**: P1  
**Effort**: 2-3 days  
**Owner**: Platform Dev  
**Status**: 🔄 In Progress  
**Module**: Infrastructure  
**Epic**: [1-infrastructure-epic](../../epics/1-infrastructure-epic.md)

---

## User Story

As a **platform developer**, I want a unified `apprun generate` CLI command that provides consistent code generation capabilities (config examples, migrations, documentation, etc.) so that I can automate repetitive tasks and maintain consistency across the platform.

---

## Vision

建立统一的代码生成 CLI 工具集，遵循 Go 社区标准（如 `go generate`），为平台提供：
- 配置示例生成（config.example, .env.example）
- 数据库迁移文件生成
- API 文档生成
- 模板代码生成
- 其他自动化生成任务

---

## Acceptance Criteria

### Functional
- [x] Unified CLI command: `apprun generate <type> [flags]`
- [x] Config generation: `apprun generate config`
  - [x] Generate `config/config.example` from registered modules
  - [x] Generate `.env.example` with environment variables
  - [x] Support nested structs and metadata comments
  - [x] Include validation rules and default values
- [ ] Migration generation: `apprun generate migration <name>`
  - [ ] Create migration files with timestamp prefix
  - [ ] Support both SQL and Atlas formats
- [ ] Documentation generation: `apprun generate docs`
  - [ ] Update Swagger/OpenAPI specs
  - [ ] Generate API markdown documentation
- [ ] Extensible plugin system for custom generators
- [ ] Consistent help and error messages across all generators

### Technical
- [x] Generator interface in `modules/config/generator.go`
- [x] CLI command structure in `cmd/generate.go`
- [ ] Plugin registration system for generators
- [ ] Uses reflection for struct introspection
- [ ] Integrated into Makefile targets
- [ ] Unit tests for each generator type
- [ ] Comprehensive CLI help and examples

---

## Architecture

### Overall Structure

```
┌──────────────────────────────────────────────┐
│ User Interface Layer                          │
│ ├─ CLI: bin/apprun generate <type>           │
│ ├─ Makefile: make config-example, etc.       │
│ └─ Git hooks: pre-commit generation          │
└──────────────────────────────────────────────┘
                    ↓
┌──────────────────────────────────────────────┐
│ Command Layer: cmd/generate.go               │
│ ├─ Cobra subcommands:                        │
│ │  ├─ config   (config.example generation)   │
│ │  ├─ migration (DB migration files)         │
│ │  ├─ docs     (API documentation)           │
│ │  └─ [future: model, handler, test]        │
│ └─ Flags: --output, --format, --dry-run     │
└──────────────────────────────────────────────┘
                    ↓
┌──────────────────────────────────────────────┐
│ Generator Registry & Interface               │
│ ├─ GeneratorRegistry (plugin system)         │
│ ├─ Generator interface (common contract)     │
│ └─ Generator implementations:                │
│    ├─ ConfigGenerator                        │
│    ├─ MigrationGenerator                     │
│    ├─ DocsGenerator                          │
│    └─ [future: custom generators]            │
└──────────────────────────────────────────────┘
                    ↓
┌──────────────────────────────────────────────┐
│ Data Sources & Templates                     │
│ ├─ ConfigRegistry (module configs)           │
│ ├─ Ent Schema (database models)              │
│ ├─ Route definitions (API endpoints)         │
│ └─ Template files (.tmpl)                    │
└──────────────────────────────────────────────┘
```

### Generator Interface (Common Contract)

```go
// Generator defines the common interface for all code generators
type Generator interface {
    // Name returns the generator type (config, migration, docs, etc.)
    Name() string
    
    // Description returns a brief description of what this generator does
    Description() string
    
    // Generate performs the generation and returns the content
    Generate(opts GenerateOptions) ([]GeneratedFile, error)
    
    // Validate checks if generation is possible with given options
    Validate(opts GenerateOptions) error
}

// GenerateOptions provides common options for all generators
type GenerateOptions struct {
    OutputDir  string            // Output directory
    Format     string            // Output format (yaml, json, sql, etc.)
    DryRun     bool              // Don't write files, just show output
    Overwrite  bool              // Overwrite existing files
    Metadata   map[string]string // Generator-specific metadata
}

// GeneratedFile represents a generated file
type GeneratedFile struct {
    Path    string // Relative path from output directory
    Content []byte // File content
    Mode    os.FileMode // File permissions
}
```

### Config Generator (Implemented)

```go
// In modules/config/generator.go
type ConfigGenerator struct {
    registry *ConfigRegistry
}

func (g *ConfigGenerator) Generate(opts GenerateOptions) ([]GeneratedFile, error) {
    var files []GeneratedFile
    
    // Generate config.example
    configContent, err := g.GenerateConfigExample()
    if err != nil {
        return nil, err
    }
    files = append(files, GeneratedFile{
        Path:    "config/config.example",
        Content: []byte(configContent),
        Mode:    0644,
    })
    
    // Generate .env.example
    envContent := g.GenerateEnvExample()
    files = append(files, GeneratedFile{
        Path:    ".env.example",
        Content: []byte(envContent),
        Mode:    0644,
    })
    
    return files, nil
}
```

---

## Implemented Features

### ✅ Config Generation (Story 1.18.1)

**Command:**
```bash
# Generate both files
./bin/apprun generate config

# Generate config.example only
./bin/apprun generate config --config-only

# Generate .env.example only
./bin/apprun generate config --env-only

# Custom output directory
./bin/apprun generate config --output /path/to/output

# Help
./bin/apprun generate config --help
```

**Features:**
1. **Dual Tag Support**: Reads both `yaml` and `mapstructure` tags
2. **Nested Structures**: Recursively generates nested struct fields with metadata
3. **Metadata Comments**: Includes `default`, `validate`, and `db` tags as inline comments
4. **Environment Variables**: Automatic UPPERCASE_UNDERSCORE naming convention
5. **Module Filtering**: Separates `envonly` modules from regular config modules

**Output Files:**
- `core/config/config.example` - YAML format with inline comments
- `core/.env.example` - Environment variable format with sections

**Module Registration Pattern:**
```go
// Automatic Discovery (Business Modules)
type Config struct {
    Logger logger.Config `register:"auto"` // Auto-registered
    I18n   i18n.Config   `register:"auto"`
    Auth   authmod.Config `register:"auto"`
}

// Manual Registration (Infrastructure Modules)
registry.Register("database", &database.Config{})
registry.Register("cache", &cache.Config{})
registry.Register("server", &server.Config{})
```

**Makefile Integration:**
```makefile
config-example:
	@echo "🔧 Generating config.example and .env.example..."
	@if [ ! -f core/bin/apprun ]; then \
		echo "⚠️  bin/apprun not found, building..." && $(MAKE) build-fast; \
	fi
	@cd core && ./bin/apprun generate config
	@echo "💡 Review and customize for your environment"
```

---

## Planned Features

### 📋 Migration Generation (Story 1.18.2)

**Command:**
```bash
# Create new migration
./bin/apprun generate migration add_user_preferences

# Create migration from model changes (Atlas)
./bin/apprun generate migration --auto-diff

# Preview migration without creating files
./bin/apprun generate migration --dry-run add_email_verification
```

**Output:**
```
migrations/
├── 20260119120000_add_user_preferences.up.sql
└── 20260119120000_add_user_preferences.down.sql
```

**Features:**
- Timestamp-based naming convention
- Template-based migration files
- Support for both SQL and Atlas formats
- Automatic diff generation from Ent schema changes

---

### 📋 Documentation Generation (Story 1.18.3)

**Command:**
```bash
# Generate all documentation
./bin/apprun generate docs

# Generate specific doc types
./bin/apprun generate docs --type swagger
./bin/apprun generate docs --type markdown

# Watch mode (regenerate on changes)
./bin/apprun generate docs --watch
```

**Output:**
```
docs/
├── api.md              # Human-readable API documentation
├── swagger.json        # OpenAPI 3.0 specification
└── swagger.yaml        # OpenAPI 3.0 (YAML format)
```

**Features:**
- Swagger/OpenAPI spec generation from route definitions
- Markdown API documentation with examples
- Changelog generation from git history
- Architecture diagram generation

---

### 📋 Model/Handler Generation (Future)

**Command:**
```bash
# Generate CRUD handlers from Ent model
./bin/apprun generate handler --model User

# Generate test scaffolding
./bin/apprun generate test --handler UserHandler

# Generate from template
./bin/apprun generate from-template --template custom.tmpl
```

---

## Design Principles

1. **Convention Over Configuration**
   - Sensible defaults for output paths
   - Follow Go community standards
   - Consistent naming conventions

2. **Composability**
   - Generators work independently
   - Can be chained via Makefile
   - Plugin-based architecture for extensibility

3. **Idempotency**
   - Same input → same output
   - Safe to run multiple times
   - Clear overwrite warnings

4. **Transparency**
   - Verbose mode shows what's being generated
   - Dry-run mode for preview
   - Clear error messages with suggestions

5. **Integration**
   - Git hooks for automatic generation
   - CI/CD pipeline integration
   - Make targets for common workflows

---

## Files Structure

```
core/
├── cmd/
│   └── generate.go              # Main CLI command with subcommands
├── modules/
│   ├── config/
│   │   └── generator.go         # ✅ Config generator (implemented)
│   ├── migration/
│   │   └── generator.go         # 📋 Migration generator (planned)
│   └── docs/
│       └── generator.go         # 📋 Docs generator (planned)
├── pkg/
│   └── generator/
│       ├── interface.go         # 📋 Common generator interface
│       ├── registry.go          # 📋 Plugin registration system
│       └── templates/           # 📋 Generation templates
│           ├── migration.sql.tmpl
│           ├── handler.go.tmpl
│           └── test.go.tmpl
└── config/
    └── config.example           # ✅ Generated output
```

---

## Implementation Phases

### ✅ Phase 1: Config Generation (Completed)
- [x] Config generator interface and implementation
- [x] CLI command: `apprun generate config`
- [x] Makefile integration: `make config-example`
- [x] Reflection-based struct introspection
- [x] YAML and ENV format support

### 📋 Phase 2: Foundation (Planned)
- [ ] Common generator interface
- [ ] Generator registry system
- [ ] Template engine setup
- [ ] Dry-run and verbose modes
- [ ] Unit test framework for generators

### 📋 Phase 3: Migration Generation (Planned)
- [ ] Migration generator implementation
- [ ] SQL template creation
- [ ] Atlas diff integration
- [ ] Timestamp naming convention

### 📋 Phase 4: Documentation Generation (Planned)
- [ ] Swagger/OpenAPI generator
- [ ] Markdown API doc generator
- [ ] Route introspection
- [ ] Example code generation

### 📋 Phase 5: Advanced Features (Future)
- [ ] Custom template support
- [ ] Model/handler code generation
- [ ] Test scaffolding
- [ ] Watch mode for auto-regeneration

---

## Usage Examples

### Example 1: Development Workflow

```bash
# After modifying config structs
make config-example

# After adding Ent models
make generate-ent
./bin/apprun generate migration add_new_model

# Before commit
make generate-all  # Runs all generators
git add .
git commit -m "Add new feature"
```

### Example 2: CI/CD Integration

```yaml
# .github/workflows/ci.yml
- name: Verify Generated Files
  run: |
    make generate-all
    git diff --exit-code || (echo "Generated files are out of sync" && exit 1)
```

### Example 3: Custom Generator Plugin

```go
// custom/generator/email_templates.go
type EmailTemplateGenerator struct{}

func (g *EmailTemplateGenerator) Name() string {
    return "email-templates"
}

func (g *EmailTemplateGenerator) Generate(opts generator.GenerateOptions) ([]generator.GeneratedFile, error) {
    // Custom generation logic
    return files, nil
}

// Register in cmd/generate.go
generator.Register("email-templates", &EmailTemplateGenerator{})
```

---

## Success Metrics

- [x] Config generation works without compilation
- [x] Consistent output across runs (idempotent)
- [x] Makefile integration seamless
- [ ] All core generators implemented (config, migration, docs)
- [ ] Generator plugin system functional
- [ ] 80%+ test coverage for generators
- [ ] < 1s execution time for typical generations
- [ ] Zero manual file updates for generated content

---

## Migration Notes

### From Standalone Script to Unified CLI

**Before (3-4-config-example-generator):**
```bash
cd core && go run ./scripts/generate-config-example.go
```
- Separate script per generator type
- Required compilation on every run
- Not reusable or composable

**After (1-18-cli-generate):**
```bash
./bin/apprun generate config
./bin/apprun generate migration <name>
./bin/apprun generate docs
```
- Unified CLI with subcommands
- Uses existing compiled binary
- Consistent interface across all generators
- Extensible plugin architecture

---

## Related

### Implementation
- [generator.go](../../core/modules/config/generator.go) - Config generator
- [generate.go](../../core/cmd/generate.go) - CLI command
- [Story 3.3: Module Config Registry](./3-3-module-config-registry.md) - Config registration pattern

### Dependencies
- **Story 1.6**: Unified CLI Architecture - CLI framework foundation
- **Story 1.5**: Database Migration - Migration file patterns
- **Story 4.x**: API Documentation - Swagger generation needs

### Related Epics
- [1-infrastructure-epic](../../epics/1-infrastructure-epic.md) - Platform infrastructure
- [3-config-epic](../../epics/3-config-epic.md) - Config generator originally from here

---

## Notes

**Design Philosophy:**
- 遵循 Go 社区标准（`go generate`, `go run`, etc.）
- 优先使用编译后的二进制（避免每次 `go run` 编译）
- 生成器应该是幂等的（同样输入产生同样输出）
- 支持 dry-run 模式便于预览
- 清晰的错误消息和建议

**Future Vision:**
- 支持自定义模板系统
- 集成 AI 辅助生成（GitHub Copilot API）
- Watch 模式自动重新生成
- 可视化生成器配置界面

**Current Status:** Phase 1 完成（Config Generation），Phase 2-5 规划中
