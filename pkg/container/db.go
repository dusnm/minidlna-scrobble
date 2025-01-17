package container

import (
	"database/sql"

	"github.com/dusnm/minidlna-scrobble/pkg/repositories/metadata"
	_ "github.com/glebarez/go-sqlite"
)

func (c *Container) GetDB() *sql.DB {
	if c.db == nil {
		db, err := sql.Open("sqlite", c.Cfg.DBFile)
		if err != nil {
			c.Logger.
				Fatal().
				Err(err).
				Msg("error opening the database file")
		}

		if err := db.Ping(); err != nil {
			c.Logger.
				Fatal().
				Err(err).
				Msg("unable to communicate with the database")
		}

		c.db = db
	}

	return c.db
}

func (c *Container) GetMetadataRepository() *metadata.Repository {
	if c.metadataRepo == nil {
		metadataRepo, err := metadata.New(
			c.GetDB(),
			c.Logger.
				With().
				Str("repository", "metadata").
				Logger(),
		)
		if err != nil {
			c.Logger.Fatal().Err(err).Msg("unable to create an instance of the metadata repo")
		}

		c.metadataRepo = metadataRepo
	}

	return c.metadataRepo
}
