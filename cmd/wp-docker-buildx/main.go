package main

import (
	"errors"
	"fmt"

	"github.com/thegeeklab/wp-docker-buildx/plugin"

	wp "github.com/thegeeklab/wp-plugin-go/plugin"
)

//nolint:gochecknoglobals
var (
	BuildVersion = "devel"
	BuildDate    = "00000000"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed")

func main() {
	settings := &plugin.Settings{}
	options := wp.Options{
		Name:            "wp-docker-buildx",
		Description:     "Build OCI container with DinD and buildx",
		Version:         BuildVersion,
		VersionMetadata: fmt.Sprintf("date=%s", BuildDate),
		Flags:           settingsFlags(settings, wp.FlagsPluginCategory),
	}

	plugin.New(options, settings).Run()
}
