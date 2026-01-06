# Linter Usage Guide

## 📋 Overview

The project uses `golangci-lint` v2.7.2+ for code quality checks, integrated with both local development and CI/CD.

## 🚀 Quick Start

### Installation

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b $(go env GOPATH)/bin

# Verify installation
golangci-lint version
```

### Basic Usage

```bash
# Run linter (same as CI)
make lint

# Run linter with auto-fix
make lint-fix

# View help
make help | grep lint
```

## 🔍 Enabled Linters

### Default Linters (Always Enabled)
- **errcheck**: Checks for unchecked errors
- **govet**: Go vet static analysis  
- **ineffassign**: Detects ineffectual assignments
- **staticcheck**: Comprehensive static checks
- **unused**: Finds unused code

### Additional Linters
- **goconst**: Finds repeated strings that should be constants
- **misspell**: Detects spelling errors
- **dupl**: Finds duplicate code
- **gosec**: Security checks
- **unconvert**: Unnecessary type conversions
- **prealloc**: Slice preallocation optimization
- **gocognit**: Cognitive complexity (threshold: 20)
- **gocyclo**: Cyclomatic complexity (threshold: 15)
- **gocritic**: Comprehensive diagnostics
- **errorlint**: Error wrapping checks (Go 1.13+)
- **bodyclose**: HTTP response body close checks

## 📊 Current Issues Summary

### Found Issues (as of last run)
- **errcheck**: 31 violations (unchecked errors)
- **errorlint**: 4 issues (error wrapping)
- **gocognit**: 3 issues (high complexity functions)
- **goconst**: 2 issues (repeated strings)

### Priority Fixing Order

1. **High Priority** - errcheck violations
   - Unchecked errors in `defer` statements
   - Test setup/teardown errors
   
2. **Medium Priority** - errorlint issues
   - Use `errors.Is()` instead of `!=` comparison
   - Use `errors.As()` for type assertions

3. **Low Priority** - goconst issues
   - Convert repeated strings to constants
   - Mostly in test files

4. **Refactoring** - gocognit issues
   - Simplify complex functions
   - Break down large test functions

## 🛠️ Common Fixes

### Unchecked Errors in defer

**Before**:
```go
defer client.Close()
```

**After**:
```go
defer func() {
    if err := client.Close(); err != nil {
        log.Error("Failed to close client", "error", err)
    }
}()
```

### Error Comparison

**Before**:
```go
if err != ErrNotFound {
    return err
}
```

**After**:
```go
if !errors.Is(err, ErrNotFound) {
    return err
}
```

### Error Type Assertion

**Before**:
```go
appErr, ok := err.(*AppError)
```

**After**:
```go
var appErr *AppError
ok := errors.As(err, &appErr)
```

### Repeated Strings

**Before**:
```go
if lang != "en-US" {
    // ...
}
if lang != "en-US" {
    // ...
}
```

**After**:
```go
const DefaultLang = "en-US"

if lang != DefaultLang {
    // ...
}
```

## 🔧 Configuration

### File: `.golangci.yml`

Key settings:
- **Version**: v2 format
- **Timeout**: 5 minutes
- **Excluded directories**: `ent/`, `docs/`, `tests/e2e/`
- **Test files**: Relaxed rules for `*_test.go`

### Customization

To adjust complexity thresholds:

```yaml
linters-settings:
  gocognit:
    min-complexity: 30  # Increase threshold
  
  gocyclo:
    min-complexity: 20  # Increase threshold
```

To disable specific checks:

```yaml
issues:
  exclude-rules:
    - path: your/path
      linters:
        - errcheck
      text: "specific error pattern"
```

## 🚦 CI Integration

### GitHub Actions

Lint runs automatically on:
- Push to `main`/`develop`
- Pull requests

### CI Workflow

```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v4
  with:
    version: latest
    working-directory: core
    args: --timeout=5m --config=../.golangci.yml
```

## 📝 Best Practices

1. **Run before commit**:
   ```bash
   make lint
   ```

2. **Fix auto-fixable issues**:
   ```bash
   make lint-fix
   ```

3. **Check specific files**:
   ```bash
   cd core && golangci-lint run ./path/to/file.go
   ```

4. **View available linters**:
   ```bash
   golangci-lint help linters
   ```

5. **Update golangci-lint**:
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

## 🐛 Troubleshooting

### "golangci-lint not found"

```bash
# Add to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or install
make lint  # Will show installation instructions
```

### "Config version error"

```bash
# Check version
golangci-lint version

# Ensure .golangci.yml has:
version: '2'
```

### "Timeout exceeded"

```bash
# Increase timeout in .golangci.yml
run:
  timeout: 10m
```

## 📚 References

- [golangci-lint Documentation](https://golangci-lint.run/)
- [Linters List](https://golangci-lint.run/usage/linters/)
- [Configuration](https://golangci-lint.run/usage/configuration/)
- [GitHub Actions Integration](https://github.com/golangci/golangci-lint-action)
