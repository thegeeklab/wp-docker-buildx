package plugin

import (
	"context"
	"reflect"
	"testing"
)

func TestFlags(t *testing.T) {
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
		for key, value := range tt.envs {
			t.Setenv(key, value)
		}

		got := New(func(_ context.Context) error { return nil })

		_ = got.App.Run([]string{"wp-docker-buildx"})
		_ = got.FlagsFromContext()

		if !reflect.DeepEqual(got.Settings.Build.Secrets, tt.want) {
			t.Errorf("%q. Build.Secrets = %v, want %v", tt.name, got.Settings.Build.Secrets, tt.want)
		}
	}
}
