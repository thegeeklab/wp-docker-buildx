package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
	"github.com/thegeeklab/wp-plugin-go/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/execabs"
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
	p.Settings.Daemon.Registry = p.Settings.Login.Registry

	if p.Settings.Build.TagsAuto {
		// return true if tag event or default branch
		if UseDefaultTag(
			p.Settings.Build.Ref,
			p.Settings.Build.Branch,
		) {
			tag, err := DefaultTagSuffix(
				p.Settings.Build.Ref,
				p.Settings.Build.TagsSuffix,
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
	// start the Docker daemon server
	//nolint: nestif
	if !p.Settings.Daemon.Disabled {
		// If no custom DNS value set start internal DNS server
		if len(p.Settings.Daemon.DNS.Value()) == 0 {
			ip, err := getContainerIP()
			if err != nil {
				log.Warn().Msgf("error detecting IP address: %v", err)
			}

			if ip != "" {
				log.Debug().Msgf("discovered IP address: %v", ip)
				p.startCoredns()

				if err := p.Settings.Daemon.DNS.Set(ip); err != nil {
					return fmt.Errorf("error setting daemon dns: %w", err)
				}
			}
		}

		p.startDaemon()
	}

	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; i < 15; i++ {
		cmd := commandInfo()

		err := cmd.Run()
		if err == nil {
			break
		}

		time.Sleep(time.Second * 1)
	}

	// Create Auth Config File
	if p.Settings.Login.Config != "" {
		if err := os.MkdirAll(dockerHome, strictFilePerm); err != nil {
			return fmt.Errorf("failed to create docker home: %w", err)
		}

		path := filepath.Join(dockerHome, "config.json")

		err := os.WriteFile(path, []byte(p.Settings.Login.Config), strictFilePerm)
		if err != nil {
			return fmt.Errorf("error writing config.json: %w", err)
		}
	}

	// login to the Docker registry
	if p.Settings.Login.Password != "" {
		cmd := commandLogin(p.Settings.Login)

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error authenticating: %w", err)
		}
	}

	if p.Settings.Daemon.BuildkitConfig != "" {
		err := os.WriteFile(buildkitConfig, []byte(p.Settings.Daemon.BuildkitConfig), strictFilePerm)
		if err != nil {
			return fmt.Errorf("error writing buildkit.toml: %w", err)
		}
	}

	switch {
	case p.Settings.Login.Password != "":
		log.Info().Msgf("Detected registry credentials")
	case p.Settings.Login.Config != "":
		log.Info().Msgf("Detected registry credentials file")
	default:
		log.Info().Msgf("Registry credentials or Docker config not provided. Guest mode enabled.")
	}

	// add proxy build args
	addProxyBuildArgs(&p.Settings.Build)

	backoffOps := func() error {
		versionCmd := commandVersion() // docker version

		versionCmd.Stdout = os.Stdout
		versionCmd.Stderr = os.Stderr
		trace(versionCmd)

		return versionCmd.Run()
	}
	backoffLog := func(err error, delay time.Duration) {
		log.Error().Msgf("failed to run docker version command: %v: retry in %s", err, delay.Truncate(time.Second))
	}

	if err := backoff.RetryNotify(backoffOps, newBackoff(daemonBackoffMaxRetries), backoffLog); err != nil {
		return err
	}

	var batchCmd []*execabs.Cmd
	batchCmd = append(batchCmd, commandInfo()) // docker info
	batchCmd = append(batchCmd, commandBuilder(p.Settings.Daemon))
	batchCmd = append(batchCmd, commandBuildx())
	batchCmd = append(batchCmd, commandBuild(p.Settings.Build, p.Settings.Dryrun)) // docker build

	// execute all commands in batch mode.
	for _, cmd := range batchCmd {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil {
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
