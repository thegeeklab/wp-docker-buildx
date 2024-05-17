package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"github.com/thegeeklab/wp-docker-buildx/docker"
	"github.com/thegeeklab/wp-plugin-go/v3/file"
	"github.com/thegeeklab/wp-plugin-go/v3/tag"
	"github.com/thegeeklab/wp-plugin-go/v3/types"
	"github.com/thegeeklab/wp-plugin-go/v3/util"
	"github.com/urfave/cli/v2"
)

var ErrTypeAssertionFailed = errors.New("type assertion failed")

const (
	strictFilePerm               = 0o600
	daemonBackoffMaxRetries      = 3
	daemonBackoffInitialInterval = 2 * time.Second
	daemonBackoffMultiplier      = 3.5
)

//nolint:revive
func (p *Plugin) run(ctx context.Context) error {
	if err := p.FlagsFromContext(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := p.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := p.Execute(); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

// Validate handles the settings validation of the plugin.
func (p *Plugin) Validate() error {
	p.Settings.Build.Branch = p.Metadata.Repository.Branch
	p.Settings.Build.Ref = p.Metadata.Curr.Ref
	p.Settings.Daemon.Registry = p.Settings.Registry.Address

	if p.Settings.Build.TagsAuto {
		// return true if tag event or default branch
		if tag.IsTaggable(
			p.Settings.Build.Ref,
			p.Settings.Build.Branch,
		) {
			tag, err := tag.SemverTagSuffix(
				p.Settings.Build.Ref,
				p.Settings.Build.TagsSuffix,
				true,
			)
			if err != nil {
				return fmt.Errorf("cannot generate tags from %s, invalid semantic version: %w", p.Settings.Build.Ref, err)
			}

			p.Settings.Build.Tags = *cli.NewStringSlice(tag...)
		} else {
			log.Info().Msgf("skip auto-tagging for %s, not on default branch or tag", p.Settings.Build.Ref)

			return nil
		}
	}

	return nil
}

// Execute provides the implementation of the plugin.
//
//nolint:gocognit
func (p *Plugin) Execute() error {
	var err error

	homeDir := util.GetUserHomeDir()
	batchCmd := make([]*types.Cmd, 0)

	// start the Docker daemon server
	//nolint: nestif
	if !p.Settings.Daemon.Disabled {
		// If no custom DNS value set start internal DNS server
		if len(p.Settings.Daemon.DNS.Value()) == 0 {
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

				if err := p.Settings.Daemon.DNS.Set(ip); err != nil {
					return fmt.Errorf("error setting daemon dns: %w", err)
				}
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
		if err := WriteDockerConf(homeDir, p.Settings.Registry.Config); err != nil {
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
		if p.Settings.Daemon.BuildkitConfigFile, err = file.WriteTmpFile("buildkit.toml", buildkitConf); err != nil {
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

	backoffOps := func() error {
		return docker.Version().Run()
	}
	backoffLog := func(err error, delay time.Duration) {
		log.Error().Msgf("failed to run docker version command: %v: retry in %s", err, delay.Truncate(time.Second))
	}

	if err := backoff.RetryNotify(backoffOps, newBackoff(daemonBackoffMaxRetries), backoffLog); err != nil {
		return err
	}

	batchCmd = append(batchCmd, docker.Info())
	batchCmd = append(batchCmd, p.Settings.Daemon.CreateBuilder())
	batchCmd = append(batchCmd, p.Settings.Daemon.ListBuilder())
	batchCmd = append(batchCmd, p.Settings.Build.Run())

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
	cacheFrom, ok := p.Context.Generic("cache-from").(*types.StringSliceFlag)
	if !ok {
		return fmt.Errorf("%w: failed to read cache-from input", ErrTypeAssertionFailed)
	}

	p.Settings.Build.CacheFrom = cacheFrom.Get()

	secrets, ok := p.Context.Generic("secrets").(*types.StringSliceFlag)
	if !ok {
		return fmt.Errorf("%w: failed to read secrets input", ErrTypeAssertionFailed)
	}

	p.Settings.Build.Secrets = secrets.Get()

	return nil
}

func newBackoff(maxRetries uint64) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = daemonBackoffInitialInterval
	b.Multiplier = daemonBackoffMultiplier

	return backoff.WithMaxRetries(b, maxRetries)
}
