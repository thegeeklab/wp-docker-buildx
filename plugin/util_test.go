package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thegeeklab/wp-docker-buildx/docker"
	plugin_base "github.com/thegeeklab/wp-plugin-go/v4/plugin"
	"github.com/urfave/cli/v3"
)

func TestWriteDockerConf(t *testing.T) {
	tests := []struct {
		name    string
		conf    string
		wantErr bool
	}{
		{
			name:    "valid json config",
			conf:    `{"auths":{"registry.example.com":{"auth":"dXNlcjpwYXNz"}}}`,
			wantErr: false,
		},
		{
			name:    "invalid json config",
			conf:    `{"auths":invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "config.json")

			err := WriteDockerConf(tmpFile, tt.conf)
			if tt.wantErr {
				assert.ErrorAs(t, err, &errInvalidDockerConfig)

				return
			}

			assert.NoError(t, err)

			content, err := os.ReadFile(tmpFile)
			assert.NoError(t, err, "Failed to read config file")

			var got, want interface{}
			err = json.Unmarshal(content, &got)
			assert.NoError(t, err, "Failed to parse written config")

			err = json.Unmarshal([]byte(tt.conf), &want)
			assert.NoError(t, err, "Failed to parse test config")

			assert.Equal(t, want, got, "Written config does not match expected")
		})
	}
}

func TestGenerateLabels(t *testing.T) {
	tests := []struct {
		name       string
		plugin     *Plugin
		wantLabels []string
	}{
		{
			name: "all fields populated",
			plugin: &Plugin{
				Settings: &Settings{
					Build: docker.Build{
						Time: "2023-01-01T00:00:00Z",
						Tags: *cli.NewStringSlice(
							"v1.0.0",
							"latest",
						),
					},
				},
				Repository: &plugin_base.Repository{
					URL: "https://github.com/example/repo",
				},
				Commit: &plugin_base.Commit{
					SHA: "abc123",
				},
			},
			wantLabels: []string{
				"org.opencontainers.image.created=2023-01-01T00:00:00Z",
				"org.opencontainers.image.source=https://github.com/example/repo",
				"org.opencontainers.image.url=https://github.com/example/repo",
				"org.opencontainers.image.revision=abc123",
				"org.opencontainers.image.version=latest",
			},
		},
		{
			name: "empty repository and commit",
			plugin: &Plugin{
				Settings: &Settings{
					Build: docker.Build{
						Time: "2023-01-01T00:00:00Z",
						Tags: *cli.NewStringSlice(
							"v1.0.0",
						),
					},
				},
			},
			wantLabels: []string{
				"org.opencontainers.image.created=2023-01-01T00:00:00Z",
				"org.opencontainers.image.version=v1.0.0",
			},
		},
		{
			name: "no tags",
			plugin: &Plugin{
				Settings: &Settings{
					Build: docker.Build{
						Time: "2023-01-01T00:00:00Z",
					},
				},
				Repository: &plugin_base.Repository{
					URL: "https://github.com/example/repo",
				},
				Commit: &plugin_base.Commit{
					SHA: "abc123",
				},
			},
			wantLabels: []string{
				"org.opencontainers.image.created=2023-01-01T00:00:00Z",
				"org.opencontainers.image.source=https://github.com/example/repo",
				"org.opencontainers.image.url=https://github.com/example/repo",
				"org.opencontainers.image.revision=abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plugin.GenerateLabels()
			assert.ElementsMatch(t, tt.wantLabels, got)
		})
	}
}
