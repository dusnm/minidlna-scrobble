package job

import (
	"context"
	"net/url"
	"syscall"
	"time"

	"github.com/dusnm/minidlna-scrobble/pkg/models"
	"github.com/dusnm/minidlna-scrobble/pkg/services/scrobble"
	"github.com/rs/zerolog"
)

type (
	Job struct {
		Ctx   context.Context
		Track models.Track
		Delay time.Duration
	}

	Service struct {
		jobChan         chan Job
		scrobbleService *scrobble.Service
		logger          zerolog.Logger
	}
)

func New(
	scrobbleService *scrobble.Service,
	logger zerolog.Logger,
) *Service {
	return &Service{
		scrobbleService: scrobbleService,
		jobChan:         make(chan Job),
		logger:          logger,
	}
}

func (s *Service) Add(job Job) {
	s.jobChan <- job
}

func (s *Service) Work(ctx context.Context) {
	go func() {
		for {
			select {
			case job := <-s.jobChan:
				s.sendWithDelay(job)
			case <-ctx.Done():
				s.logger.Info().Msg("closing")
				return
			}
		}
	}()
}

func (s *Service) sendWithDelay(job Job) {
	t := time.After(job.Delay)
	go func() {
		select {
		case <-t:
			s.send(job)
		case <-job.Ctx.Done():
			return
		}
	}()
}

func (s *Service) send(job Job) {
	scrobbles, err := s.scrobbleService.Scrobble(job.Ctx, job.Track)
	if err != nil {
		s.logger.
			Error().
			Err(err).
			Msg("")

		switch v := err.(type) {
		case *url.Error:
			// Everything should be retried with no delay in case of a network error
			s.jobChan <- Job{
				Ctx:   job.Ctx,
				Delay: time.Duration(0),
				Track: job.Track,
			}
		case scrobble.ErrorResponse:
			switch v.Code {
			case scrobble.CodeServiceOffline, scrobble.CodeServiceTemporaryUnavailable:
				// Only these codes indicate that the scrobble should be retried
				s.jobChan <- Job{
					Ctx:   job.Ctx,
					Delay: time.Duration(0),
					Track: job.Track,
				}
			case scrobble.CodeInvalidSessionKey:
				// This indicates that the session with last.fm has been revoked
				// and that the user should re-authenticate. This will not be
				// handled for the user, so we'll just terminate the process here.
				s.logger.
					Error().
					Msg("last.fm session invalid, re-authentication required, terminating")

				// Gracefully exit the program by sending a
				// signal it's programmed to intercept.
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			}
		}

		return
	}

	s.logger.
		Info().
		Str("artist", scrobbles.Scrobbles.Scrobble.Artist.Text).
		Str("track", scrobbles.Scrobbles.Scrobble.Track.Text).
		Int("accepted", scrobbles.Scrobbles.Attr.Accepted).
		Int("ignored", scrobbles.Scrobbles.Attr.Ignored).
		Msg("successful scrobble")
}
