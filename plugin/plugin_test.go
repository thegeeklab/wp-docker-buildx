package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretsFlag(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want []string
	}{
		{
			name: "parse secrets list with escape",
			envs: map[string]string{
				"PLUGIN_SECRETS": "id=raw_file_secret\\,src=file.txt,id=SECRET_TOKEN",
			},
			want: []string{
				"id=raw_file_secret,src=file.txt",
				"id=SECRET_TOKEN",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := New(func(_ context.Context) error { return nil })

			_ = got.App.Run([]string{"wp-docker-buildx"})
			_ = got.FlagsFromContext()

			assert.EqualValues(t, tt.want, got.Settings.Build.Secrets)
		})
	}
}

func TestEnvironmentFlag(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want []string
	}{
		{
			name: "simple environment",
			envs: map[string]string{
				"PLUGIN_ENVIRONMENT": `{"env1": "value1", "env2": "value2"}`,
			},
			want: []string{
				"env1=value1",
				"env2=value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := New(func(_ context.Context) error { return nil })

			_ = got.App.Run([]string{"wp-docker-buildx"})
			_ = got.FlagsFromContext()

			assert.EqualValues(t, tt.want, got.Environment.Value())
		})
	}
}
