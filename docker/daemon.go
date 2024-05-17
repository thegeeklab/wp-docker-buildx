package docker

import (
	"os/exec"

	"github.com/thegeeklab/wp-plugin-go/v3/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/execabs"
)

const dockerdBin = "/usr/local/bin/dockerd"

// Daemon defines Docker daemon parameters.
type Daemon struct {
	Registry             string          // Docker registry
	Mirror               string          // Docker registry mirror
	Insecure             bool            // Docker daemon enable insecure registries
	StorageDriver        string          // Docker daemon storage driver
	StoragePath          string          // Docker daemon storage path
	Disabled             bool            // DOcker daemon is disabled (already running)
	Debug                bool            // Docker daemon started in debug mode
	Bip                  string          // Docker daemon network bridge IP address
	DNS                  cli.StringSlice // Docker daemon dns server
	DNSSearch            cli.StringSlice // Docker daemon dns search domain
	MTU                  string          // Docker daemon mtu setting
	IPv6                 bool            // Docker daemon IPv6 networking
	Experimental         bool            // Docker daemon enable experimental mode
	BuildkitConfigFile   string          // Docker buildkit config file
	MaxConcurrentUploads string          // Docker daemon max concurrent uploads
}

// helper function to create the docker daemon command.
func (d *Daemon) Start() *types.Cmd {
	args := []string{
		"--data-root", d.StoragePath,
		"--host=unix:///var/run/docker.sock",
	}

	if d.StorageDriver != "" {
		args = append(args, "-s", d.StorageDriver)
	}

	if d.Insecure && d.Registry != "" {
		args = append(args, "--insecure-registry", d.Registry)
	}

	if d.IPv6 {
		args = append(args, "--ipv6")
	}

	if d.Mirror != "" {
		args = append(args, "--registry-mirror", d.Mirror)
	}

	if d.Bip != "" {
		args = append(args, "--bip", d.Bip)
	}

	for _, dns := range d.DNS.Value() {
		args = append(args, "--dns", dns)
	}

	for _, dnsSearch := range d.DNSSearch.Value() {
		args = append(args, "--dns-search", dnsSearch)
	}

	if d.MTU != "" {
		args = append(args, "--mtu", d.MTU)
	}

	if d.Experimental {
		args = append(args, "--experimental")
	}

	if d.MaxConcurrentUploads != "" {
		args = append(args, "--max-concurrent-uploads", d.MaxConcurrentUploads)
	}

	return &types.Cmd{
		Cmd:     execabs.Command(dockerdBin, args...),
		Private: !d.Debug,
	}
}

func (d *Daemon) CreateBuilder() *types.Cmd {
	args := []string{
		"buildx",
		"create",
		"--use",
	}

	if d.BuildkitConfigFile != "" {
		args = append(args, "--config", d.BuildkitConfigFile)
	}

	return &types.Cmd{
		Cmd: execabs.Command(dockerBin, args...),
	}
}

func (d *Daemon) ListBuilder() *types.Cmd {
	return &types.Cmd{
		Cmd: execabs.Command(dockerBin, "buildx", "ls"),
	}
}

func (d *Daemon) StartCoreDNS() *types.Cmd {
	return &types.Cmd{
		Cmd:     exec.Command("coredns", "-conf", "/etc/coredns/Corefile"),
		Private: !d.Debug,
	}
}
