package version_test

import (
	"strings"
	"testing"

	"apprun/pkg/version"
)

func TestBuildVersion(t *testing.T) {
	result := version.BuildVersion()

	// Should contain version and commit in format: "version (commit)"
	if !strings.Contains(result, "(") || !strings.Contains(result, ")") {
		t.Errorf("BuildVersion() = %v, should contain commit in parentheses", result)
	}

	// Should start with version
	if !strings.HasPrefix(result, version.Version) {
		t.Errorf("BuildVersion() = %v, should start with version %v", result, version.Version)
	}

	// If commit is long enough, should be truncated to 7 chars
	if len(version.GitCommit) >= 7 {
		expectedCommit := version.GitCommit[:7]
		if !strings.Contains(result, expectedCommit) {
			t.Errorf("BuildVersion() = %v, should contain truncated commit %v", result, expectedCommit)
		}
	}
}

func TestInfo(t *testing.T) {
	result := version.Info()

	// Check that result contains expected sections
	requiredStrings := []string{
		"AppRun BaaS Platform",
		"Version:",
		"Git Commit:",
		"Build Time:",
		"Go Version:",
		"Platform:",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(result, required) {
			t.Errorf("Info() missing required string: %s", required)
		}
	}
}

func TestShort(t *testing.T) {
	result := version.Short()

	// Short() should return just the version string
	if result == "" {
		t.Error("Short() returned empty string")
	}

	// Should not contain parentheses (no commit hash)
	if strings.Contains(result, "(") || strings.Contains(result, ")") {
		t.Errorf("Short() should not contain commit hash, got: %s", result)
	}
}

func TestVersionConstants(t *testing.T) {
	// Ensure constants are set (even if to default values)
	if version.Version == "" {
		t.Error("Version constant is empty")
	}

	if version.GitCommit == "" {
		t.Error("GitCommit constant is empty")
	}

	if version.BuildTime == "" {
		t.Error("BuildTime constant is empty")
	}
}
