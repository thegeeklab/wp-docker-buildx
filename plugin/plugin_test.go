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

			_ = got.App.Run(t.Context(), []string{"wp-docker-buildx"})
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

			_ = got.App.Run(t.Context(), []string{"wp-docker-buildx"})
			_ = got.FlagsFromContext()

			assert.ElementsMatch(t, tt.want, got.Environment.Value())
		})
	}
}

func TestCacheFromFlag(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want []string
	}{
		{
			name: "simple escape",
			envs: map[string]string{
				"PLUGIN_CACHE_FROM": `type=registry\,ref=example,foo=bar`,
			},
			want: []string{
				"type=registry,ref=example",
				"foo=bar",
			},
		},
		{
			name: "double escape",
			envs: map[string]string{
				"PLUGIN_CACHE_FROM": "type=registry\\,ref=example,foo=bar",
			},
			want: []string{
				"type=registry,ref=example",
				"foo=bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := New(func(_ context.Context) error { return nil })

			_ = got.App.Run(t.Context(), []string{"wp-docker-buildx"})
			_ = got.FlagsFromContext()

			assert.ElementsMatch(t, tt.want, got.Settings.Build.CacheFrom)
		})
	}
}

func TestBuildArgsFlag(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want map[string]string
	}{
		{
			name: "not nil",
			envs: map[string]string{},
			want: map[string]string{},
		},
		{
			name: "parse args",
			envs: map[string]string{
				"PLUGIN_BUILD_ARGS": `{"arg1": "value1", "arg2": "value2"}`,
			},
			want: map[string]string{
				"arg1": "value1",
				"arg2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}

			got := New(func(_ context.Context) error { return nil })

			_ = got.App.Run(t.Context(), []string{"wp-docker-buildx"})
			_ = got.FlagsFromContext()

			assert.Equal(t, tt.want, got.Settings.Build.Args)
		})
	}
}
