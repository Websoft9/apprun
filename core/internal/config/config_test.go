package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		kind     reflect.Kind
		expected interface{}
	}{
		{"bool true", "true", reflect.Bool, true},
		{"bool false", "false", reflect.Bool, false},
		{"int", "42", reflect.Int, 42},
		{"int64", "123", reflect.Int64, 123},
		{"string", "hello", reflect.String, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDefaultValue(tt.value, tt.kind)
			assert.Equal(t, tt.expected, result)
		})
	}
}
