package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
