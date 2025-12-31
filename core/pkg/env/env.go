package env

import (
"os"
"strconv"
"time"
)

// Get retrieves an environment variable or returns the default value if not set
func Get(key, defaultValue string) string {
if value := os.Getenv(key); value != "" {
return value
}
return defaultValue
}

// MustGet retrieves an environment variable or panics if not set
func MustGet(key string) string {
value := os.Getenv(key)
if value == "" {
panic("environment variable " + key + " is required but not set")
}
return value
}

// GetInt retrieves an integer environment variable or returns the default value
func GetInt(key string, defaultValue int) int {
value := os.Getenv(key)
if value == "" {
return defaultValue
}

intValue, err := strconv.Atoi(value)
if err != nil {
return defaultValue
}
return intValue
}

// GetBool retrieves a boolean environment variable or returns the default value
// Accepts: "true", "1", "yes", "on" (case-insensitive) as true
func GetBool(key string, defaultValue bool) bool {
value := os.Getenv(key)
if value == "" {
return defaultValue
}

boolValue, err := strconv.ParseBool(value)
if err != nil {
return defaultValue
}
return boolValue
}

// GetDuration retrieves a duration environment variable or returns the default value
// Accepts values like "5s", "10m", "1h"
func GetDuration(key string, defaultValue time.Duration) time.Duration {
value := os.Getenv(key)
if value == "" {
return defaultValue
}

duration, err := time.ParseDuration(value)
if err != nil {
return defaultValue
}
return duration
}
