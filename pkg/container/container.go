package container

import (
	"github.com/dusnm/minidlna-scrobble/pkg/config"
	"github.com/dusnm/minidlna-scrobble/pkg/services/auth"
	"github.com/dusnm/minidlna-scrobble/pkg/services/metadata"
	"github.com/dusnm/minidlna-scrobble/pkg/services/scrobble"
	"github.com/dusnm/minidlna-scrobble/pkg/services/sessioncache"
	"github.com/dusnm/minidlna-scrobble/pkg/services/watcher"
)

type (
	Container struct {
		Cfg                 *config.Config
		authService         *auth.Service
		sessionCacheService *sessioncache.Service
		watcherService      *watcher.Service
		metadataService     *metadata.Service
		scrobbleService     *scrobble.Service
	}
)

func New() (*Container, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	return &Container{
		Cfg: cfg,
	}, nil
}

func (c *Container) Close() error {
	if c.watcherService != nil {
		return c.watcherService.Close()
	}

	return nil
}
