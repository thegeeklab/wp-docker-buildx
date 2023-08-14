package plugin

import (
	wp "github.com/thegeeklab/wp-plugin-go/plugin"
	"github.com/urfave/cli/v2"
)

// Plugin implements provide the plugin implementation.
type Plugin struct {
	*wp.Plugin
	Settings *Settings
}

// Settings for the Plugin.
type Settings struct {
	Daemon Daemon
	Login  Login
	Build  Build
	Dryrun bool
}

// Daemon defines Docker daemon parameters.
type Daemon struct {
	Registry       string          // Docker registry
	Mirror         string          // Docker registry mirror
	Insecure       bool            // Docker daemon enable insecure registries
	StorageDriver  string          // Docker daemon storage driver
	StoragePath    string          // Docker daemon storage path
	Disabled       bool            // DOcker daemon is disabled (already running)
	Debug          bool            // Docker daemon started in debug mode
	Bip            string          // Docker daemon network bridge IP address
	DNS            cli.StringSlice // Docker daemon dns server
	DNSSearch      cli.StringSlice // Docker daemon dns search domain
	MTU            string          // Docker daemon mtu setting
	IPv6           bool            // Docker daemon IPv6 networking
	Experimental   bool            // Docker daemon enable experimental mode
	BuildkitConfig string          // Docker buildkit config
}

// Login defines Docker login parameters.
type Login struct {
	Registry string // Docker registry address
	Username string // Docker registry username
	Password string // Docker registry password
	Email    string // Docker registry email
	Config   string // Docker Auth Config
}

// Build defines Docker build parameters.
type Build struct {
	Ref           string          // Git commit ref
	Branch        string          // Git repository branch
	Containerfile string          // Docker build Containerfile
	Context       string          // Docker build context
	TagsAuto      bool            // Docker build auto tag
	TagsSuffix    string          // Docker build tags with suffix
	Tags          cli.StringSlice // Docker build tags
	ExtraTags     cli.StringSlice // Docker build tags including registry
	Platforms     cli.StringSlice // Docker build target platforms
	Args          cli.StringSlice // Docker build args
	ArgsEnv       cli.StringSlice // Docker build args from env
	Target        string          // Docker build target
	Pull          bool            // Docker build pull
	CacheFrom     []string        // Docker build cache-from
	CacheTo       string          // Docker build cache-to
	Compress      bool            // Docker build compress
	Repo          string          // Docker build repository
	NoCache       bool            // Docker build no-cache
	AddHost       cli.StringSlice // Docker build add-host
	Quiet         bool            // Docker build quiet
	Output        string          // Docker build output folder
	NamedContext  cli.StringSlice // Docker build named context
	Labels        cli.StringSlice // Docker build labels
	Provenance    string          // Docker build provenance attestation
	SBOM          string          // Docker build sbom attestation
	Secrets       []string        // Docker build secrets
}

func New(options wp.Options, settings *Settings) *Plugin {
	p := &Plugin{}

	options.Execute = p.run

	p.Plugin = wp.New(options)
	p.Settings = settings

	return p
}
