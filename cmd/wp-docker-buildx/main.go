package main

import (
	"fmt"

	"github.com/thegeeklab/wp-docker-buildx/plugin"

	wp "github.com/thegeeklab/wp-plugin-go/plugin"
)

//nolint:gochecknoglobals
var (
	BuildVersion = "devel"
	BuildDate    = "00000000"
)

func main() {
	settings := &plugin.Settings{}
	options := wp.Options{
		Name:            "wp-docker-buildx",
		Description:     "Build multiarch OCI images with buildx",
		Version:         BuildVersion,
		VersionMetadata: fmt.Sprintf("date=%s", BuildDate),
		Flags:           settingsFlags(settings, wp.FlagsPluginCategory),
	}

	plugin.New(options, settings).Run()
}
