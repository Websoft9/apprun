// Package version provides version information for the AppRun platform.
// Version data is injected at build time via -ldflags.
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version (e.g., "v1.2.3")
	// Set via: -ldflags "-X apprun/pkg/version.Version=v1.2.3"
	Version = "dev"

	// GitCommit is the git commit hash
	// Set via: -ldflags "-X apprun/pkg/version.GitCommit=$(git rev-parse HEAD)"
	GitCommit = "unknown"

	// BuildTime is the build timestamp in RFC3339 format
	// Set via: -ldflags "-X apprun/pkg/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	BuildTime = "unknown"
)

// BuildVersion returns a formatted version string including commit hash
// Example: "v1.2.3 (a1b2c3d)"
func BuildVersion() string {
	if len(GitCommit) >= 7 {
		return fmt.Sprintf("%s (%s)", Version, GitCommit[:7])
	}
	return fmt.Sprintf("%s (%s)", Version, GitCommit)
}

// Info returns detailed version information
func Info() string {
	return fmt.Sprintf(`AppRun BaaS Platform
Version:    %s
Git Commit: %s
Build Time: %s
Go Version: %s
Platform:   %s/%s`,
		Version,
		GitCommit,
		BuildTime,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

// Short returns just the version number
func Short() string {
	return Version
}
