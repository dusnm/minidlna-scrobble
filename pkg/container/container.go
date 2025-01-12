package container

import (
	"database/sql"
	"errors"

	"github.com/dusnm/minidlna-scrobble/pkg/config"
	"github.com/dusnm/minidlna-scrobble/pkg/repositories/metadata"
	"github.com/dusnm/minidlna-scrobble/pkg/services/auth"
	"github.com/dusnm/minidlna-scrobble/pkg/services/job"
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
		db                  *sql.DB
		authService         *auth.Service
		sessionCacheService *sessioncache.Service
		watcherService      *watcher.Service
		scrobbleService     *scrobble.Service
		jobService          *job.Service
		metadataRepo        *metadata.Repository
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
	var err error
	if c.watcherService != nil {
		err = errors.Join(err, c.watcherService.Close())
	}

	if c.metadataRepo != nil {
		err = errors.Join(err, c.metadataRepo.Close())
	}

	return err
}
