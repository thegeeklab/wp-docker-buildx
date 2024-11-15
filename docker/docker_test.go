package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasProxyBuildArg(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]string
		key      string
		expected bool
	}{
		{
			name:     "lowercase key exists",
			args:     map[string]string{"http_proxy": "http://proxy.example.com"},
			key:      "http_proxy",
			expected: true,
		},
		{
			name:     "uppercase key exists",
			args:     map[string]string{"HTTP_PROXY": "http://proxy.example.com"},
			key:      "http_proxy",
			expected: true,
		},
		{
			name:     "key does not exist",
			args:     map[string]string{},
			key:      "http_proxy",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Build{Args: tt.args}
			result := b.hasProxyBuildArg(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProxyValue(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		key      string
		expected string
	}{
		{
			name: "lowercase env var exists",
			envVars: map[string]string{
				"http_proxy": "http://proxy.local",
			},
			key:      "http_proxy",
			expected: "http://proxy.local",
		},
		{
			name: "uppercase env var exists",
			envVars: map[string]string{
				"HTTP_PROXY": "http://proxy.upper.local",
			},
			key:      "http_proxy",
			expected: "http://proxy.upper.local",
		},
		{
			name: "both cases exist, lowercase preferred",
			envVars: map[string]string{
				"http_proxy": "http://proxy.lower.local",
				"HTTP_PROXY": "http://proxy.upper.local",
			},
			key:      "http_proxy",
			expected: "http://proxy.lower.local",
		},
		{
			name:     "no env vars exist",
			envVars:  map[string]string{},
			key:      "http_proxy",
			expected: "",
		},
		{
			name: "different proxy type",
			envVars: map[string]string{
				"https_proxy": "https://secure.proxy.local",
				"HTTPS_PROXY": "https://secure.proxy.upper.local",
			},
			key:      "https_proxy",
			expected: "https://secure.proxy.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			b := &Build{}
			result := b.getProxyValue(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddArgFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		key      string
		existing map[string]string
		expected map[string]string
	}{
		{
			name: "add new env var",
			envVars: map[string]string{
				"NEW_VAR": "test_value",
			},
			key: "NEW_VAR",
			expected: map[string]string{
				"NEW_VAR": "test_value",
			},
		},
		{
			name:     "empty env var",
			envVars:  map[string]string{},
			key:      "MISSING_VAR",
			expected: map[string]string{},
		},
		{
			name: "env var with empty value",
			envVars: map[string]string{
				"EMPTY_VAR": "",
			},
			key:      "EMPTY_VAR",
			expected: map[string]string{},
		},
		{
			name: "multiple env vars",
			envVars: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			key: "VAR1",
			expected: map[string]string{
				"VAR1": "value1",
			},
		},
		{
			name: "special characters in value",
			envVars: map[string]string{
				"SPECIAL_VAR": "!@#$%^&*()",
			},
			key: "SPECIAL_VAR",
			expected: map[string]string{
				"SPECIAL_VAR": "!@#$%^&*()",
			},
		},
		{
			name: "preserve existing args",
			envVars: map[string]string{
				"TEST_VAR": "new_value",
			},
			key: "TEST_VAR",
			existing: map[string]string{
				"TEST_VAR": "old_value",
			},
			expected: map[string]string{
				"TEST_VAR": "old_value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment before each test
			for k := range tt.envVars {
				t.Setenv(k, "")
			}

			// Set test environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			b := &Build{Args: make(map[string]string)}
			if tt.existing != nil {
				b.Args = tt.existing
			}

			b.addArgFromEnv(tt.key)
			assert.Equal(t, tt.expected, b.Args)
		})
	}
}
