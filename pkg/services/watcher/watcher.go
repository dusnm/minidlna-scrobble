package watcher

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dusnm/minidlna-scrobble/pkg/config"
	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/helpers"
	"github.com/dusnm/minidlna-scrobble/pkg/logparser"
	"github.com/dusnm/minidlna-scrobble/pkg/models"
	"github.com/dusnm/minidlna-scrobble/pkg/repositories/metadata"
	"github.com/dusnm/minidlna-scrobble/pkg/services/scrobble"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

type (
	Service struct {
		cfg             *config.Config
		mu              *sync.Mutex
		logger          zerolog.Logger
		metadata        *metadata.Repository
		scrobbleService *scrobble.Service
		jobs            map[string]context.CancelFunc
		watcher         *fsnotify.Watcher
	}
)

func New(
	cfg *config.Config,
	metadataRepo *metadata.Repository,
	scrobbleService *scrobble.Service,
	logger zerolog.Logger,
) (*Service, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return &Service{}, nil
	}

	return &Service{
		cfg:             cfg,
		mu:              &sync.Mutex{},
		logger:          logger,
		metadata:        metadataRepo,
		scrobbleService: scrobbleService,
		jobs:            make(map[string]context.CancelFunc, 0),
		watcher:         w,
	}, nil
}

func (s *Service) Close() error {
	s.logger.Info().Msg("closing")

	return s.watcher.Close()
}

func (s *Service) Watch(ctx context.Context) error {
	go func() {
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					return
				}

				if event.Name != s.cfg.LogFile {
					s.logger.
						Debug().
						Str("event", event.String()).
						Msg("not interested in this event")

					continue
				}

				if !event.Has(fsnotify.Write) {
					s.logger.
						Debug().
						Str("event", event.String()).
						Msg("not a write event")

					continue
				}

				line, err := s.lastLine()
				if err != nil {
					s.logger.Error().Err(err).Msg("")
					continue
				}

				if !strings.Contains(line, constants.MagicLogValue) {
					s.logger.
						Debug().
						Str("line", line).
						Msg("not interested in this log line")

					continue
				}

				// Cancel any previously enqueued jobs
				// if they didn't complete by now, they don't count
				s.cancelJobs()

				parsed, err := logparser.ParseLine(line)
				if err != nil {
					s.logger.Error().Err(err).Msg("")
					continue
				}

				id, err := strconv.Atoi(parsed.MessageID)
				if err != nil {
					s.logger.Error().Err(err).Msg("")
					continue
				}

				md, err := s.metadata.GetByID(ctx, id)
				if err != nil {
					s.logger.Error().Err(err).Msg("")
					continue
				}

				npResp, err := s.scrobbleService.SendNowPlaying(ctx, md)
				if err != nil {
					s.logger.Error().Err(err).Msg("")
					continue
				}

				if npResp.NowPlaying.IgnoredMessage.Code != "0" {
					s.logger.
						Info().
						Str("artist", npResp.NowPlaying.Artist.Text).
						Str("track", npResp.NowPlaying.Track.Text).
						Msg("ignoring track")

					continue
				}

				if err = s.enqueueScrobble(ctx, md); err != nil {
					s.logger.Error().Err(err).Msg("")
				}
			case err, ok := <-s.watcher.Errors:
				if !ok {
					return
				}

				s.logger.Error().Err(err).Msg("")
			}
		}
	}()

	if err := s.watcher.Add(filepath.Dir(s.cfg.LogFile)); err != nil {
		return err
	}

	return nil
}

func (s *Service) lastLine() (string, error) {
	f, err := os.OpenFile(s.cfg.LogFile, os.O_RDONLY, 0o644)
	if err != nil {
		return "", nil
	}

	defer f.Close()

	line := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return line, nil
}

func (s *Service) cancelJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, cancel := range s.jobs {
		s.logger.
			Debug().
			Str("id", id).
			Msg("cancelling job")

		cancel()
	}

	s.jobs = make(map[string]context.CancelFunc, 0)
}

func (s *Service) enqueueScrobble(ctx context.Context, md models.Track) error {
	ctx, cancel := context.WithCancel(ctx)
	if md.Duration <= time.Second*30 {
		// Not worth scrobbling
		s.logger.
			Info().
			Str("artist", md.Artist).
			Str("track", md.Name).
			Msg("track too short to scrobble")

		cancel()
		return nil
	}

	d := md.Duration / 2
	if d >= time.Minute*4 {
		d = time.Minute * 4
	}

	jobID, err := helpers.RandomID()
	if err != nil {
		cancel()
		return err
	}

	s.mu.Lock()
	s.jobs[jobID] = cancel
	s.mu.Unlock()

	go func(
		ctx context.Context,
		md models.Track,
		jobID string,
		offset time.Duration,
	) {
		ticker := time.NewTicker(offset)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				scrobbles, err := s.scrobbleService.Scrobble(ctx, md)
				if err != nil {
					// Well, we tried
					s.logger.Error().Err(err).Msg("")
				} else {
					s.logger.
						Info().
						Str("artist", scrobbles.Scrobbles.Scrobble.Artist.Text).
						Str("track", scrobbles.Scrobbles.Scrobble.Track.Text).
						Int("accepted", scrobbles.Scrobbles.Attr.Accepted).
						Int("ignored", scrobbles.Scrobbles.Attr.Ignored).
						Msg("successfull scrobble")
				}

				s.mu.Lock()
				delete(s.jobs, jobID)
				s.mu.Unlock()
				return
			case <-ctx.Done():
				return
			}
		}
	}(ctx, md, jobID, d)

	return nil
}
