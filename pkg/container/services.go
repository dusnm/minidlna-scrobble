package container

import (
	"github.com/dusnm/minidlna-scrobble/pkg/services/auth"
	"github.com/dusnm/minidlna-scrobble/pkg/services/job"
	"github.com/dusnm/minidlna-scrobble/pkg/services/scrobble"
	"github.com/dusnm/minidlna-scrobble/pkg/services/sessioncache"
	"github.com/dusnm/minidlna-scrobble/pkg/services/watcher"
)

func (c *Container) GetAuthService() *auth.Service {
	if c.authService == nil {
		c.authService = auth.New(c.Cfg.Credentials)
	}

	return c.authService
}

func (c *Container) GetSessionCacheService() *sessioncache.Service {
	if c.sessionCacheService == nil {
		service, err := sessioncache.New()
		if err != nil {
			c.Logger.
				Fatal().
				Err(err).
				Msg("unable to create an instance of session cache")
		}

		c.sessionCacheService = service
	}

	return c.sessionCacheService
}

func (c *Container) GetWatcherService() *watcher.Service {
	if c.watcherService == nil {
		watcherService, err := watcher.New(
			c.Cfg,
			c.GetMetadataRepository(),
			c.GetScrobbleService(),
			c.GetJobService(),
			c.Logger.
				With().
				Str("service", "watcher").
				Logger(),
		)
		if err != nil {
			c.Logger.
				Fatal().
				Err(err).
				Msg("unable to create an instance of watcher")
		}

		c.watcherService = watcherService
	}

	return c.watcherService
}

func (c *Container) GetScrobbleService() *scrobble.Service {
	if c.scrobbleService == nil {
		c.scrobbleService = scrobble.New(
			c.Cfg.Credentials,
			c.GetSessionCacheService(),
		)
	}

	return c.scrobbleService
}

func (c *Container) GetJobService() *job.Service {
	if c.jobService == nil {
		c.jobService = job.New(
			c.GetScrobbleService(),
			c.Logger.
				With().
				Str("service", "job").
				Logger(),
		)
	}

	return c.jobService
}
