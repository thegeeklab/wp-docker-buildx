package docker

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/thegeeklab/wp-plugin-go/v3/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/execabs"
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
	Dryrun        bool            // Docker build dryrun
}

// helper function to create the docker login command.
func (r *Registry) Login() *types.Cmd {
	args := []string{
		"login",
		"-u", r.Username,
		"-p", r.Password,
	}

	if r.Email != "" {
		args = append(args, "-e", r.Email)
	}

	args = append(args, r.Address)

	return &types.Cmd{
		Cmd: execabs.Command(dockerBin, args...),
	}
}

// helper function to create the docker info command.
func Version() *types.Cmd {
	return &types.Cmd{
		Cmd: execabs.Command(dockerBin, "version"),
	}
}

// helper function to create the docker info command.
func Info() *types.Cmd {
	return &types.Cmd{
		Cmd: execabs.Command(dockerBin, "info"),
	}
}

// helper function to create the docker build command.
func (b *Build) Run() *types.Cmd {
	args := []string{
		"buildx",
		"build",
		"--rm=true",
		"-f", b.Containerfile,
	}

	defaultBuildArgs := []string{
		fmt.Sprintf("DOCKER_IMAGE_CREATED=%s", time.Now().Format(time.RFC3339)),
	}

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
		b.addProxyValue(arg)
	}

	for _, arg := range append(defaultBuildArgs, b.Args.Value()...) {
		args = append(args, "--build-arg", arg)
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

	return &types.Cmd{
		Cmd: execabs.Command(dockerBin, args...),
	}
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

	if len(value) > 0 && !b.hasProxyBuildArg(key) {
		b.Args = *cli.NewStringSlice(append(b.Args.Value(), fmt.Sprintf("%s=%s", key, value))...)
		b.Args = *cli.NewStringSlice(append(b.Args.Value(), fmt.Sprintf("%s=%s", strings.ToUpper(key), value))...)
	}
}

// helper function to get a proxy value from the environment.
//
// assumes that the upper and lower case versions of are the same.
func (b *Build) getProxyValue(key string) string {
	value := os.Getenv(key)

	if len(value) > 0 {
		return value
	}

	return os.Getenv(strings.ToUpper(key))
}

// helper function that looks to see if a proxy value was set in the build args.
func (b *Build) hasProxyBuildArg(key string) bool {
	keyUpper := strings.ToUpper(key)

	for _, s := range b.Args.Value() {
		if strings.HasPrefix(s, key) || strings.HasPrefix(s, keyUpper) {
			return true
		}
	}

	return false
}
