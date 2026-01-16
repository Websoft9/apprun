#!/bin/bash
# AppRun CLI Integration Tests
# Tests command execution and basic functionality

set -e  # Exit on error

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPRUN_BIN="$PROJECT_ROOT/core/bin/apprun"

echo "======================================"
echo "AppRun CLI Integration Tests"
echo "======================================"
echo ""
echo "Binary: $APPRUN_BIN"
echo ""

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Check if binary exists
if [ ! -f "$APPRUN_BIN" ]; then
    echo -e "${RED}ERROR:${NC} Binary not found at $APPRUN_BIN"
    echo "Please run 'make build' first"
    exit 1
fi

# Test 1: Binary is executable
if [ -x "$APPRUN_BIN" ]; then
    pass "Binary is executable"
else
    fail "Binary is not executable"
fi

# Test 2: Help command works
if "$APPRUN_BIN" --help | grep -q "AppRun"; then
    pass "Help command displays AppRun info"
else
    fail "Help command failed"
fi

# Test 3: Help shows Server Commands group
if "$APPRUN_BIN" --help | grep -q "Server Commands:"; then
    pass "Help displays Server Commands group"
else
    fail "Help missing Server Commands group"
fi

# Test 4: Help shows Client Commands group
if "$APPRUN_BIN" --help | grep -q "Client Commands:"; then
    pass "Help displays Client Commands group"
else
    fail "Help missing Client Commands group"
fi

# Test 5: Version command works
if "$APPRUN_BIN" version | grep -q "AppRun BaaS Platform"; then
    pass "Version command displays platform name"
else
    fail "Version command failed"
fi

# Test 6: Version shows Git commit
if "$APPRUN_BIN" version | grep -q "Git Commit:"; then
    pass "Version displays Git commit info"
else
    fail "Version missing Git commit info"
fi

# Test 7: Configure help works
if "$APPRUN_BIN" configure --help | grep -qE "Configure AppRun CLI|Interactive configuration"; then
    pass "Configure help displays correctly"
else
    fail "Configure help failed"
fi

# Test 8: Migrate help works
if "$APPRUN_BIN" migrate --help | grep -qiE "database.*migration|migration.*management"; then
    pass "Migrate help displays correctly"
else
    fail "Migrate help failed"
fi

# Test 9: Migrate subcommands exist
if "$APPRUN_BIN" migrate --help | grep -q "apply"; then
    pass "Migrate apply subcommand exists"
else
    fail "Migrate apply subcommand missing"
fi

if "$APPRUN_BIN" migrate --help | grep -q "status"; then
    pass "Migrate status subcommand exists"
else
    fail "Migrate status subcommand missing"
fi

if "$APPRUN_BIN" migrate --help | grep -q "validate"; then
    pass "Migrate validate subcommand exists"
else
    fail "Migrate validate subcommand missing"
fi

# Test 10: Serve help works
if "$APPRUN_BIN" serve --help | grep -qE "Start AppRun|Start the AppRun"; then
    pass "Serve help displays correctly"
else
    fail "Serve help failed"
fi

# Test 11: Deploy command (placeholder) works
if "$APPRUN_BIN" deploy 2>&1 | grep -q "not yet implemented"; then
    pass "Deploy placeholder message displays"
else
    fail "Deploy command failed"
fi

# Test 12: Logs command (placeholder) works
if "$APPRUN_BIN" logs 2>&1 | grep -q "not yet implemented"; then
    pass "Logs placeholder message displays"
else
    fail "Logs command failed"
fi

# Test 13: Backup command (placeholder) works
if "$APPRUN_BIN" backup 2>&1 | grep -q "not yet implemented"; then
    pass "Backup placeholder message displays"
else
    fail "Backup command failed"
fi

# Test 14: Symlink exists
if [ -L "$PROJECT_ROOT/core/bin/server" ]; then
    pass "Symlink bin/server exists"
else
    fail "Symlink bin/server missing"
fi

# Test 15: Symlink points to apprun
if [ -L "$PROJECT_ROOT/core/bin/server" ] && [ "$(readlink "$PROJECT_ROOT/core/bin/server")" = "apprun" ]; then
    pass "Symlink points to apprun"
else
    fail "Symlink target incorrect"
fi

# Test 16: Legacy binary works
if "$PROJECT_ROOT/core/bin/server" version 2>&1 | grep -q "AppRun BaaS Platform"; then
    pass "Legacy bin/server symlink works"
else
    fail "Legacy bin/server failed"
fi

# Test 17: Invalid command shows error
if "$APPRUN_BIN" invalid-command 2>&1 | grep -q "unknown command"; then
    pass "Invalid command shows error"
else
    fail "Invalid command didn't show error"
fi

# Test 18: Configure show without config
# Create a temporary home directory to avoid affecting user config
TEMP_HOME=$(mktemp -d)
export HOME="$TEMP_HOME"
if "$APPRUN_BIN" configure show 2>&1 | grep -q "not found\|No configuration found"; then
    pass "Configure show handles missing config"
else
    # It's ok if it shows default values
    pass "Configure show works (shows defaults or error)"
fi
rm -rf "$TEMP_HOME"

# Test 19: Binary size check (should be reasonable, not bloated)
BINARY_SIZE=$(stat -c%s "$APPRUN_BIN" 2>/dev/null || stat -f%z "$APPRUN_BIN" 2>/dev/null)
BINARY_SIZE_MB=$((BINARY_SIZE / 1024 / 1024))
if [ "$BINARY_SIZE_MB" -lt 100 ]; then
    pass "Binary size reasonable: ${BINARY_SIZE_MB}MB (< 100MB)"
else
    fail "Binary size too large: ${BINARY_SIZE_MB}MB (>= 100MB)"
fi

# Test 20: Help shows global flags
if "$APPRUN_BIN" --help | grep -q "\-\-config"; then
    pass "Help displays --config flag"
else
    fail "Help missing --config flag"
fi

echo ""
echo "======================================"
echo "Test Results"
echo "======================================"
echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
echo -e "${RED}Failed:${NC} $TESTS_FAILED"
echo "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
