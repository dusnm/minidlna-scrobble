package metadata

import (
	"context"
	"database/sql"
	"time"

	"github.com/dusnm/minidlna-scrobble/pkg/helpers"
	"github.com/dusnm/minidlna-scrobble/pkg/models"
	"github.com/rs/zerolog"
)

type (
	Repository struct {
		db                *sql.DB
		selectDetailsStmt *sql.Stmt
		logger            zerolog.Logger
	}
)

const (
	selectDetailsQuery = "SELECT ARTIST, ALBUM, TITLE, DURATION, TRACK FROM DETAILS WHERE ID = ?"
)

func New(
	db *sql.DB,
	logger zerolog.Logger,
) (*Repository, error) {
	stmt, err := db.Prepare(selectDetailsQuery)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:                db,
		selectDetailsStmt: stmt,
		logger:            logger,
	}, nil
}

func (r *Repository) Close() error {
	r.logger.Info().Msg("closing")

	return r.selectDetailsStmt.Close()
}

func (r *Repository) GetByID(ctx context.Context, ID int) (models.Track, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	row := r.selectDetailsStmt.QueryRowContext(ctx, ID)

	var (
		artist   string
		album    string
		title    string
		duration string
		track    int
	)

	err := row.Scan(
		&artist,
		&album,
		&title,
		&duration,
		&track,
	)

	if err != nil {
		return models.Track{}, err
	}

	d, err := helpers.ParseDBDuration(duration)
	if err != nil {
		return models.Track{}, err
	}

	return models.Track{
		Artist:    helpers.ReplaceSpecialChars(artist),
		Name:      helpers.ReplaceSpecialChars(title),
		Timestamp: time.Now(),
		Album:     helpers.ReplaceSpecialChars(album),
		Duration:  d,
		Number:    track,
	}, nil
}
