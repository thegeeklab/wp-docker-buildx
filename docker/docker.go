package docker

import (
	"fmt"
	"maps"
	"os"
	"strings"

	plugin_exec "github.com/thegeeklab/wp-plugin-go/v4/exec"
	"github.com/urfave/cli/v2"
)

const dockerBin = "/usr/local/bin/docker"

// Login defines Docker login parameters.
type Registry struct {
	Address  string // Docker registry address
	Username string // Docker registry username
	Password string // Docker registry password
	Email    string // Docker registry email
	Config   string // Docker Auth Config
}

// Build defines Docker build parameters.
type Build struct {
	Ref           string            // Git commit ref
	Branch        string            // Git repository branch
	Containerfile string            // Docker build Containerfile
	Context       string            // Docker build context
	TagsAuto      bool              // Docker build auto tag
	TagsSuffix    string            // Docker build tags with suffix
	Tags          cli.StringSlice   // Docker build tags
	ExtraTags     cli.StringSlice   // Docker build tags including registry
	Platforms     cli.StringSlice   // Docker build target platforms
	Args          map[string]string // Docker build args
	ArgsEnv       cli.StringSlice   // Docker build args from env
	Target        string            // Docker build target
	Pull          bool              // Docker build pull
	CacheFrom     []string          // Docker build cache-from
	CacheTo       string            // Docker build cache-to
	Compress      bool              // Docker build compress
	Repo          string            // Docker build repository
	NoCache       bool              // Docker build no-cache
	AddHost       cli.StringSlice   // Docker build add-host
	Quiet         bool              // Docker build quiet
	Output        string            // Docker build output folder
	NamedContext  cli.StringSlice   // Docker build named context
	Labels        cli.StringSlice   // Docker build labels
	LabelsAuto    bool              // Docker build labels auto
	Provenance    string            // Docker build provenance attestation
	SBOM          string            // Docker build sbom attestation
	Secrets       []string          // Docker build secrets
	Dryrun        bool              // Docker build dryrun
	Time          string            // Docker build time
}

// helper function to create the docker login command.
func (r *Registry) Login() *plugin_exec.Cmd {
	args := []string{
		"login",
		"-u", r.Username,
		"-p", r.Password,
	}

	if r.Email != "" {
		args = append(args, "-e", r.Email)
	}

	args = append(args, r.Address)

	cmd := plugin_exec.Command(dockerBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// helper function to create the docker info command.
func Version() *plugin_exec.Cmd {
	cmd := plugin_exec.Command(dockerBin, "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// helper function to create the docker info command.
func Info() *plugin_exec.Cmd {
	cmd := plugin_exec.Command(dockerBin, "info")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// helper function to create the docker build command.
func (b *Build) Run(env []string) *plugin_exec.Cmd {
	args := []string{
		"buildx",
		"build",
		"--rm=true",
		"-f", b.Containerfile,
	}

	defaultBuildArgs := map[string]string{
		"DOCKER_IMAGE_CREATED": b.Time,
	}

	maps.Copy(b.Args, defaultBuildArgs)

	args = append(args, b.Context)
	if !b.Dryrun && b.Output == "" && len(b.Tags.Value()) > 0 {
		args = append(args, "--push")
	}

	if b.Compress {
		args = append(args, "--compress")
	}

	if b.Pull {
		args = append(args, "--pull=true")
	}

	if b.NoCache {
		args = append(args, "--no-cache")
	}

	for _, arg := range b.CacheFrom {
		args = append(args, "--cache-from", arg)
	}

	if b.CacheTo != "" {
		args = append(args, "--cache-to", b.CacheTo)
	}

	for _, arg := range b.ArgsEnv.Value() {
		b.addArgFromEnv(arg)
	}

	for key, value := range b.Args {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	for _, host := range b.AddHost.Value() {
		args = append(args, "--add-host", host)
	}

	if b.Target != "" {
		args = append(args, "--target", b.Target)
	}

	if b.Quiet {
		args = append(args, "--quiet")
	}

	if b.Output != "" {
		args = append(args, "--output", b.Output)
	}

	for _, arg := range b.NamedContext.Value() {
		args = append(args, "--build-context", arg)
	}

	if len(b.Platforms.Value()) > 0 {
		args = append(args, "--platform", strings.Join(b.Platforms.Value(), ","))
	}

	for _, arg := range b.Tags.Value() {
		args = append(args, "-t", fmt.Sprintf("%s:%s", b.Repo, arg))
	}

	for _, arg := range b.ExtraTags.Value() {
		args = append(args, "-t", arg)
	}

	for _, arg := range b.Labels.Value() {
		args = append(args, "--label", arg)
	}

	if b.Provenance != "" {
		args = append(args, "--provenance", b.Provenance)
	}

	if b.SBOM != "" {
		args = append(args, "--sbom", b.SBOM)
	}

	for _, secret := range b.Secrets {
		args = append(args, "--secret", secret)
	}

	cmd := plugin_exec.Command(dockerBin, args...)

	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// helper function to add proxy values from the environment.
func (b *Build) AddProxyBuildArgs() {
	b.addProxyValue("http_proxy")
	b.addProxyValue("https_proxy")
	b.addProxyValue("no_proxy")
}

// helper function to add the upper and lower case version of a proxy value.
func (b *Build) addProxyValue(key string) {
	value := b.getProxyValue(key)

	if value != "" && !b.hasProxyBuildArg(key) {
		b.Args[key] = value
		b.Args[strings.ToUpper(key)] = value
	}
}

func (b *Build) addArgFromEnv(key string) {
	if value, ok := b.Args[key]; ok && value != "" {
		return
	}

	if value, ok := os.LookupEnv(key); ok && value != "" {
		b.Args[key] = value
	}
}

// helper function to get a proxy value from the environment.
//
// assumes that the upper and lower case versions of are the same.
func (b *Build) getProxyValue(key string) string {
	value := os.Getenv(key)

	if value != "" {
		return value
	}

	return os.Getenv(strings.ToUpper(key))
}

// helper function that looks to see if a proxy value was set in the build args.
func (b *Build) hasProxyBuildArg(key string) bool {
	if _, ok := b.Args[key]; ok {
		return true
	}

	if _, ok := b.Args[strings.ToUpper(key)]; ok {
		return true
	}

	return false
}
