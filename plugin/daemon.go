package plugin

import (
	"io"
	"os"

	"github.com/thegeeklab/wp-plugin-go/trace"
)

const (
	dockerBin      = "/usr/local/bin/docker"
	dockerdBin     = "/usr/local/bin/dockerd"
	dockerHome     = "/root/.docker/"
	buildkitConfig = "/tmp/buildkit.toml"
)

func (p Plugin) startDaemon() {
	cmd := commandDaemon(p.Settings.Daemon)
	if p.Settings.Daemon.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	go func() {
		trace.Cmd(cmd)
		_ = cmd.Run()
	}()
}
