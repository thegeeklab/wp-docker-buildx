package plugin

import (
	"fmt"

	"github.com/thegeeklab/wp-docker-buildx/docker"
	plugin_cli "github.com/thegeeklab/wp-plugin-go/v6/cli"
	plugin_base "github.com/thegeeklab/wp-plugin-go/v6/plugin"
	"github.com/urfave/cli/v3"
)

//go:generate go run ../internal/docs/main.go -output=../docs/data/data-raw.yaml

// Plugin implements provide the plugin.
type Plugin struct {
	*plugin_base.Plugin
	Repository *plugin_base.Repository
	Commit     *plugin_base.Commit
	Settings   *Settings
}

// Settings for the Plugin.
type Settings struct {
	BuildkitConfig string

	Daemon   docker.Daemon
	Registry docker.Registry
	Build    docker.Build
}

func New(e plugin_base.ExecuteFunc, build ...string) *Plugin {
	p := &Plugin{
		Settings: &Settings{},
	}

	options := plugin_base.Options{
		Name:                "wp-docker-buildx",
		Description:         "Build multiarch OCI images with buildx",
		Flags:               Flags(p.Settings, plugin_base.FlagsPluginCategory),
		Execute:             p.run,
		HideWoodpeckerFlags: true,
	}

	if len(build) > 0 {
		options.Version = build[0]
	}

	if len(build) > 1 {
		options.VersionMetadata = fmt.Sprintf("date=%s", build[1])
	}

	if e != nil {
		options.Execute = e
	}

	p.Plugin = plugin_base.New(options)

	return p
}

// Flags returns a slice of CLI flags for the plugin.
//
//nolint:maintidx
func Flags(settings *Settings, category string) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "dry-run",
			Sources:     cli.EnvVars("PLUGIN_DRY_RUN"),
			Usage:       "disable docker push",
			Destination: &settings.Build.Dryrun,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "daemon.mirror",
			Sources:     cli.EnvVars("PLUGIN_MIRROR", "DOCKER_PLUGIN_MIRROR"),
			Usage:       "registry mirror to pull images",
			Destination: &settings.Daemon.Mirror,
			DefaultText: "$DOCKER_PLUGIN_MIRROR",
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "daemon.storage-driver",
			Sources:     cli.EnvVars("PLUGIN_STORAGE_DRIVER"),
			Usage:       "docker daemon storage driver",
			Destination: &settings.Daemon.StorageDriver,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "daemon.storage-path",
			Sources:     cli.EnvVars("PLUGIN_STORAGE_PATH"),
			Usage:       "docker daemon storage path",
			Value:       "/var/lib/docker",
			Destination: &settings.Daemon.StoragePath,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "daemon.bip",
			Sources:     cli.EnvVars("PLUGIN_BIP"),
			Usage:       "allow the docker daemon to bride IP address",
			Destination: &settings.Daemon.Bip,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "daemon.mtu",
			Sources:     cli.EnvVars("PLUGIN_MTU"),
			Usage:       "docker daemon custom MTU setting",
			Destination: &settings.Daemon.MTU,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name:        "daemon.dns",
			Sources:     cli.EnvVars("PLUGIN_CUSTOM_DNS"),
			Usage:       "custom docker daemon DNS server",
			Destination: &settings.Daemon.DNS,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name:        "daemon.dns-search",
			Sources:     cli.EnvVars("PLUGIN_CUSTOM_DNS_SEARCH"),
			Usage:       "custom docker daemon DNS search domain",
			Destination: &settings.Daemon.DNSSearch,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "daemon.insecure",
			Sources:     cli.EnvVars("PLUGIN_INSECURE"),
			Usage:       "allow the docker daemon to use insecure registries",
			Value:       false,
			Destination: &settings.Daemon.Insecure,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "daemon.ipv6",
			Sources:     cli.EnvVars("PLUGIN_IPV6"),
			Usage:       "enable docker daemon IPv6 support",
			Value:       false,
			Destination: &settings.Daemon.IPv6,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "daemon.experimental",
			Sources:     cli.EnvVars("PLUGIN_EXPERIMENTAL"),
			Usage:       "enable docker daemon experimental mode",
			Value:       false,
			Destination: &settings.Daemon.Experimental,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "daemon.debug",
			Sources:     cli.EnvVars("PLUGIN_DEBUG"),
			Usage:       "enable verbose debug mode for the docker daemon",
			Value:       false,
			Destination: &settings.Daemon.Debug,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "daemon.off",
			Sources:     cli.EnvVars("PLUGIN_DAEMON_OFF"),
			Usage:       "disable the startup of the docker daemon",
			Value:       false,
			Destination: &settings.Daemon.Disabled,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "daemon.buildkit-config",
			Sources:     cli.EnvVars("PLUGIN_BUILDKIT_CONFIG"),
			Usage:       "content of the docker buildkit toml config",
			Destination: &settings.BuildkitConfig,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "daemon.max-concurrent-uploads",
			Sources:     cli.EnvVars("PLUGIN_MAX_CONCURRENT_UPLOADS"),
			Usage:       "max concurrent uploads for each push",
			Destination: &settings.Daemon.MaxConcurrentUploads,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "containerfile",
			Sources:     cli.EnvVars("PLUGIN_CONTAINERFILE"),
			Usage:       "containerfile to use for the image build",
			Value:       "Containerfile",
			Destination: &settings.Build.Containerfile,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "context",
			Sources:     cli.EnvVars("PLUGIN_CONTEXT"),
			Usage:       "path of the build context",
			Value:       ".",
			Destination: &settings.Build.Context,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name:        "named-context",
			Sources:     cli.EnvVars("PLUGIN_NAMED_CONTEXT"),
			Usage:       "additional named build context",
			Destination: &settings.Build.NamedContext,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name: "tags",
			Sources: cli.ValueSourceChain{
				Chain: []cli.ValueSource{
					cli.EnvVar("PLUGIN_TAGS"),
					cli.EnvVar("PLUGIN_TAG"),
					cli.File(".tags"),
				},
			},
			Usage:       "repository tags to use for the image",
			Destination: &settings.Build.Tags,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "tags.auto",
			Sources:     cli.EnvVars("PLUGIN_AUTO_TAG", "PLUGIN_DEFAULT_TAGS"),
			Usage:       "generate tag names automatically based on git branch and git tag",
			Value:       false,
			Destination: &settings.Build.TagsAuto,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "tags.suffix",
			Sources:     cli.EnvVars("PLUGIN_AUTO_TAG_SUFFIX", "PLUGIN_DEFAULT_SUFFIX"),
			Usage:       "generate tag names with the given suffix",
			Destination: &settings.Build.TagsSuffix,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name: "extra.tags",
			Sources: cli.ValueSourceChain{
				Chain: []cli.ValueSource{
					cli.EnvVar("PLUGIN_EXTRA_TAGS"),
					cli.File(".extratags"),
				},
			},
			Usage:       "additional tags to use for the image including registry",
			Destination: &settings.Build.ExtraTags,
			Category:    category,
		},
		&plugin_cli.StringMapFlag{
			Name:        "args",
			Sources:     cli.EnvVars("PLUGIN_BUILD_ARGS"),
			Usage:       "custom build arguments for the build",
			Destination: &settings.Build.Args,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name:        "args-from-env",
			Sources:     cli.EnvVars("PLUGIN_BUILD_ARGS_FROM_ENV"),
			Usage:       "forward environment variables as custom arguments to the build",
			Destination: &settings.Build.ArgsEnv,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "quiet",
			Sources:     cli.EnvVars("PLUGIN_QUIET"),
			Usage:       "enable suppression of the build output",
			Value:       false,
			Destination: &settings.Build.Quiet,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "output",
			Sources:     cli.EnvVars("PLUGIN_OUTPUT"),
			Usage:       "export action for the build result",
			Destination: &settings.Build.Output,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "target",
			Sources:     cli.EnvVars("PLUGIN_TARGET"),
			Usage:       "build target to use",
			Destination: &settings.Build.Target,
			Category:    category,
		},
		&plugin_cli.StringSliceFlag{
			Name:        "cache-from",
			Sources:     cli.EnvVars("PLUGIN_CACHE_FROM"),
			Usage:       "images to consider as cache sources",
			Destination: &settings.Build.CacheFrom,
			Config: plugin_cli.StringSliceConfig{
				Delimiter:    ",",
				EscapeString: "\\",
			},
			Category: category,
		},
		&cli.StringFlag{
			Name:        "cache-to",
			Sources:     cli.EnvVars("PLUGIN_CACHE_TO"),
			Usage:       "cache destination for the build cache",
			Destination: &settings.Build.CacheTo,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "pull-image",
			Sources:     cli.EnvVars("PLUGIN_PULL_IMAGE"),
			Usage:       "enforce to pull base image at build time",
			Value:       true,
			Destination: &settings.Build.Pull,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "compress",
			Sources:     cli.EnvVars("PLUGIN_COMPRESS"),
			Usage:       "enable compression of the build context using gzip",
			Value:       false,
			Destination: &settings.Build.Compress,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "repo",
			Sources:     cli.EnvVars("PLUGIN_REPO"),
			Usage:       "repository name for the image",
			Destination: &settings.Build.Repo,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "docker.registry",
			Sources:     cli.EnvVars("PLUGIN_REGISTRY", "DOCKER_REGISTRY"),
			Usage:       "docker registry to authenticate with",
			Value:       "https://index.docker.io/v1/",
			Destination: &settings.Registry.Address,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "docker.username",
			Sources:     cli.EnvVars("PLUGIN_USERNAME", "DOCKER_USERNAME"),
			Usage:       "username for registry authentication",
			Destination: &settings.Registry.Username,
			DefaultText: "$DOCKER_USERNAME",
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "docker.password",
			Sources:     cli.EnvVars("PLUGIN_PASSWORD", "DOCKER_PASSWORD"),
			Usage:       "password for registry authentication",
			Destination: &settings.Registry.Password,
			DefaultText: "$DOCKER_PASSWORD",
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "docker.email",
			Sources:     cli.EnvVars("PLUGIN_EMAIL", "DOCKER_EMAIL"),
			Usage:       "email address for registry authentication",
			Destination: &settings.Registry.Email,
			DefaultText: "$DOCKER_EMAIL",
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "registry.config",
			Sources:     cli.EnvVars("PLUGIN_REGISTRY_CONFIG", "DOCKER_REGISTRY_CONFIG"),
			Usage:       "content of the registry credentials store file",
			Destination: &settings.Registry.Config,
			DefaultText: "$DOCKER_REGISTRY_CONFIG",
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "no-cache",
			Sources:     cli.EnvVars("PLUGIN_NO_CACHE"),
			Usage:       "disable the usage of cached intermediate containers",
			Value:       false,
			Destination: &settings.Build.NoCache,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name:        "add-host",
			Sources:     cli.EnvVars("PLUGIN_ADD_HOST"),
			Usage:       "additional `host:ip` mapping",
			Destination: &settings.Build.AddHost,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name:        "platforms",
			Sources:     cli.EnvVars("PLUGIN_PLATFORMS"),
			Usage:       "target platform for build",
			Destination: &settings.Build.Platforms,
			Category:    category,
		},
		&cli.StringSliceFlag{
			Name:        "labels",
			Sources:     cli.EnvVars("PLUGIN_LABELS"),
			Usage:       "labels to add to image",
			Destination: &settings.Build.Labels,
			Category:    category,
		},
		&cli.BoolFlag{
			Name:        "labels.auto",
			Sources:     cli.EnvVars("PLUGIN_AUTO_LABEL", "PLUGIN_DEFAULT_LABELS"),
			Usage:       "generates labels automatically based on git repository information",
			Value:       false,
			Destination: &settings.Build.LabelsAuto,
			Category:    category,
		},

		&cli.StringFlag{
			Name:        "provenance",
			Sources:     cli.EnvVars("PLUGIN_PROVENANCE"),
			Usage:       "generates provenance attestation for the build",
			Destination: &settings.Build.Provenance,
			Category:    category,
		},
		&cli.StringFlag{
			Name:        "sbom",
			Sources:     cli.EnvVars("PLUGIN_SBOM"),
			Usage:       "generates SBOM attestation for the build",
			Destination: &settings.Build.SBOM,
			Category:    category,
		},
		&plugin_cli.StringSliceFlag{
			Name:        "secrets",
			Sources:     cli.EnvVars("PLUGIN_SECRETS"),
			Usage:       "exposes secrets to the build",
			Destination: &settings.Build.Secrets,
			Config: plugin_cli.StringSliceConfig{
				Delimiter:    ",",
				EscapeString: "\\",
			},
			Category: category,
		},
	}
}
