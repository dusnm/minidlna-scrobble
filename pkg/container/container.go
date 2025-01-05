package container

import (
	"github.com/dusnm/minidlna-scrobble/pkg/config"
	"github.com/dusnm/minidlna-scrobble/pkg/services/auth"
	"github.com/dusnm/minidlna-scrobble/pkg/services/metadata"
	"github.com/dusnm/minidlna-scrobble/pkg/services/scrobble"
	"github.com/dusnm/minidlna-scrobble/pkg/services/sessioncache"
	"github.com/dusnm/minidlna-scrobble/pkg/services/watcher"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	Container struct {
		Cfg                 *config.Config
		Logger              zerolog.Logger
		authService         *auth.Service
		sessionCacheService *sessioncache.Service
		watcherService      *watcher.Service
		metadataService     *metadata.Service
		scrobbleService     *scrobble.Service
	}
)

func New(logLevel zerolog.Level) (*Container, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	return &Container{
		Cfg: cfg,
		Logger: log.
			Logger.
			Level(logLevel).
			With().
			Str("app", "minidlna-scrobble").
			Logger(),
	}, nil
}

func (c *Container) Close() error {
	if c.watcherService != nil {
		return c.watcherService.Close()
	}

	return nil
}
