package watcher

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dusnm/minidlna-scrobble/pkg/config"
	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/helpers"
	"github.com/dusnm/minidlna-scrobble/pkg/logparser"
	"github.com/dusnm/minidlna-scrobble/pkg/services/metadata"
	"github.com/dusnm/minidlna-scrobble/pkg/services/scrobble"
	"github.com/fsnotify/fsnotify"
)

type (
	Job struct {
		Cancel     context.CancelFunc
		FinishedAt time.Time
	}

	Service struct {
		cfg             *config.Config
		mu              *sync.Mutex
		metadata        *metadata.Service
		scrobbleService *scrobble.Service
		jobs            map[string]context.CancelFunc
	}
)

func New(
	cfg *config.Config,
	metadataService *metadata.Service,
	scrobbleService *scrobble.Service,
) *Service {
	return &Service{
		cfg:             cfg,
		mu:              &sync.Mutex{},
		metadata:        metadataService,
		scrobbleService: scrobbleService,
		jobs:            make(map[string]context.CancelFunc, 0),
	}
}

func (s *Service) Watch(ctx context.Context) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}

				if event.Name == s.cfg.LogFile && event.Has(fsnotify.Write) {
					line, err := s.lastLine()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
						continue
					}

					if strings.HasPrefix(line, constants.MagicLogValue) {
						// Cancel any previously enqueed jobs
						// if they didn't complete by now, they don't count
						s.cancelJobs()

						parsed, err := logparser.ParseLine(strings.NewReader(line))
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
							continue
						}

						f, err := os.OpenFile(parsed.Filepath, os.O_RDONLY, 0o644)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
							continue
						}

						md, err := s.metadata.Read(f)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
							continue
						}

						f.Close()

						npResp, err := s.scrobbleService.SendNowPlaying(ctx, md)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
							continue
						}

						if npResp.NowPlaying.IgnoredMessage.Code == "0" {
							err = s.enqueeScrobble(ctx, md)
							if err != nil {
								fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
								continue
							}
						}
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}

				fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			}
		}
	}()

	if err = w.Add(filepath.Dir(s.cfg.LogFile)); err != nil {
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

	for _, cancel := range s.jobs {
		cancel()
	}

	s.jobs = make(map[string]context.CancelFunc, 0)
}

func (s *Service) enqueeScrobble(ctx context.Context, md metadata.Track) error {
	ctx, cancel := context.WithCancel(ctx)
	if md.Duration <= time.Second*30 {
		// Not worth scrobbling
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
		md metadata.Track,
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
					fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
				}

				fmt.Printf(
					"scrobbled: %s - %s, ignored: %d",
					scrobbles.Scrobbles.Scrobble.Artist.Text,
					scrobbles.Scrobbles.Scrobble.Track.Text,
					scrobbles.Scrobbles.Attr.Ignored,
				)

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
