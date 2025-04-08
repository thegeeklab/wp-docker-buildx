package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/rs/zerolog/log"
	"github.com/thegeeklab/wp-docker-buildx/docker"
	plugin_exec "github.com/thegeeklab/wp-plugin-go/v4/exec"
	plugin_file "github.com/thegeeklab/wp-plugin-go/v4/file"
	plugin_tag "github.com/thegeeklab/wp-plugin-go/v4/tag"
	plugin_util "github.com/thegeeklab/wp-plugin-go/v4/util"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed")

const (
	strictFilePerm               = 0o600
	daemonBackoffMaxRetries      = 3
	daemonBackoffInitialInterval = 2 * time.Second
	daemonBackoffMultiplier      = 3.5
)

func (p *Plugin) run(ctx context.Context) error {
	if err := p.FlagsFromContext(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := p.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := p.Execute(ctx); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

// Validate handles the settings validation of the plugin.
func (p *Plugin) Validate() error {
	var err error

	p.Settings.Build.Time = time.Now().Format(time.RFC3339)
	p.Settings.Build.Branch = p.Metadata.Repository.Branch
	p.Settings.Build.Ref = p.Metadata.Curr.Ref
	p.Settings.Daemon.Registry = p.Settings.Registry.Address

	if p.Settings.Build.TagsAuto {
		// return true if tag event or default branch
		if plugin_tag.IsTaggable(
			p.Settings.Build.Ref,
			p.Settings.Build.Branch,
		) {
			p.Settings.Build.Tags, err = plugin_tag.SemverTagSuffix(
				p.Settings.Build.Ref,
				p.Settings.Build.TagsSuffix,
				true,
			)
			if err != nil {
				return fmt.Errorf("cannot generate tags from %s, invalid semantic version: %w", p.Settings.Build.Ref, err)
			}
		} else {
			log.Info().Msgf("skip auto-tagging for %s, not on default branch or tag", p.Settings.Build.Ref)

			return nil
		}
	}

	if p.Settings.Build.LabelsAuto {
		p.Settings.Build.Labels = p.GenerateLabels()
	}

	return nil
}

// Execute provides the implementation of the plugin.
//
//nolint:gocognit
func (p *Plugin) Execute(ctx context.Context) error {
	var err error

	homeDir := plugin_util.GetUserHomeDir()
	batchCmd := make([]*plugin_exec.Cmd, 0)

	// start the Docker daemon server
	//nolint: nestif
	if !p.Settings.Daemon.Disabled {
		// If no custom DNS value set start internal DNS server
		if len(p.Settings.Daemon.DNS) == 0 {
			ip, err := GetContainerIP()
			if err != nil {
				log.Warn().Msgf("error detecting IP address: %v", err)
			}

			if ip != "" {
				log.Debug().Msgf("discovered IP address: %v", ip)

				cmd := p.Settings.Daemon.StartCoreDNS()
				go func() {
					_ = cmd.Run()
				}()

				p.Settings.Daemon.DNS = append(p.Settings.Daemon.DNS, ip)
			}
		}

		cmd := p.Settings.Daemon.Start()
		go func() {
			_ = cmd.Run()
		}()
	}

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; i < 15; i++ {
		cmd := docker.Info()

		err := cmd.Run()
		if err == nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	if p.Settings.Registry.Config != "" {
		path := filepath.Join(homeDir, ".docker", "config.json")
		if err := os.MkdirAll(filepath.Dir(path), strictFilePerm); err != nil {
			return err
		}

		if err := WriteDockerConf(path, p.Settings.Registry.Config); err != nil {
			return fmt.Errorf("error writing docker config: %w", err)
		}
	}

	if p.Settings.Registry.Password != "" {
		if err := p.Settings.Registry.Login().Run(); err != nil {
			return fmt.Errorf("error authenticating: %w", err)
		}
	}

	buildkitConf := p.Settings.BuildkitConfig
	if buildkitConf != "" {
		if p.Settings.Daemon.BuildkitConfigFile, err = plugin_file.WriteTmpFile("buildkit.toml", buildkitConf); err != nil {
			return fmt.Errorf("error writing buildkit config: %w", err)
		}

		defer os.Remove(p.Settings.Daemon.BuildkitConfigFile)
	}

	switch {
	case p.Settings.Registry.Password != "":
		log.Info().Msgf("Detected registry credentials")
	case p.Settings.Registry.Config != "":
		log.Info().Msgf("Detected registry credentials file")
	default:
		log.Info().Msgf("Registry credentials or Docker config not provided. Guest mode enabled.")
	}

	p.Settings.Build.AddProxyBuildArgs()

	bf := backoff.NewExponentialBackOff()
	bf.InitialInterval = daemonBackoffInitialInterval
	bf.Multiplier = daemonBackoffMultiplier

	bfo := func() (any, error) {
		return nil, docker.Version().Run()
	}

	bfn := func(err error, delay time.Duration) {
		log.Error().Msgf("failed to run docker version command: %v: retry in %s", err, delay.Truncate(time.Second))
	}

	_, err = backoff.Retry(ctx, bfo,
		backoff.WithBackOff(bf),
		backoff.WithMaxTries(daemonBackoffMaxRetries),
		backoff.WithNotify(bfn))
	if err != nil {
		return err
	}

	batchCmd = append(batchCmd, docker.Info())
	batchCmd = append(batchCmd, p.Settings.Daemon.CreateBuilder())
	batchCmd = append(batchCmd, p.Settings.Daemon.ListBuilder())
	batchCmd = append(batchCmd, p.Settings.Build.Run(p.Environment.Value()))

	for _, cmd := range batchCmd {
		if cmd == nil {
			continue
		}

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) FlagsFromContext() error {
	var ok bool

	p.Settings.Build.CacheFrom, ok = p.App.Value("cache-from").([]string)
	if !ok {
		return fmt.Errorf("%w: failed to read cache-from input", ErrTypeAssertionFailed)
	}

	p.Settings.Build.Secrets, ok = p.App.Value("secrets").([]string)
	if !ok {
		return fmt.Errorf("%w: failed to read secrets input", ErrTypeAssertionFailed)
	}

	p.Settings.Build.Args, ok = p.App.Value("args").(map[string]string)
	if !ok {
		return fmt.Errorf("%w: failed to read args input", ErrTypeAssertionFailed)
	}

	if p.Settings.Build.Args == nil {
		p.Settings.Build.Args = make(map[string]string)
	}

	return nil
}
