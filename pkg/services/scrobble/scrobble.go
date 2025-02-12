package scrobble

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dusnm/minidlna-scrobble/pkg/config"
	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/helpers"
	"github.com/dusnm/minidlna-scrobble/pkg/models"
	"github.com/dusnm/minidlna-scrobble/pkg/services/sessioncache"
)

const (
	CodeInvalidSessionKey           = 9
	CodeServiceOffline              = 11
	CodeServiceTemporaryUnavailable = 16
)

type (
	Service struct {
		cfg          config.Credentials
		sessionCache *sessioncache.Service
		client       *http.Client
	}

	ErrorResponse struct {
		Message string `json:"message"`
		Code    uint   `json:"error"`
	}

	Ignored struct {
		Code string `json:"code"`
		Text string `json:"#text"`
	}

	Correction struct {
		Corrected string `json:"corrected"`
		Text      string `json:"#text"`
	}

	NowPlayingResponse struct {
		NowPlaying struct {
			IgnoredMessage Ignored              `json:"ignoredMessage"`
			Artist         struct{ Correction } `json:"artist"`
			Track          struct{ Correction } `json:"track"`
			AlbumArtist    struct{ Correction } `json:"albumArtist"`
			Album          struct{ Correction } `json:"album"`
		} `json:"nowplaying"`
	}

	ScrobbleResponse struct {
		Scrobbles struct {
			Scrobble struct {
				IgnoredMessage Ignored              `json:"ignoredMessage"`
				Artist         struct{ Correction } `json:"artist"`
				Track          struct{ Correction } `json:"track"`
				AlbumArtist    struct{ Correction } `json:"albumArtist"`
				Album          struct{ Correction } `json:"album"`
				Timestamp      string               `json:"timestamp"`
			} `json:"scrobble"`
			Attr struct {
				Ignored  int `json:"ignored"`
				Accepted int `json:"accepted"`
			} `json:"@attr"`
		} `json:"scrobbles"`
	}
)

func (er ErrorResponse) Error() string {
	return fmt.Sprintf("request failed with: message - %s, code - %d", er.Message, er.Code)
}

func New(
	cfg config.Credentials,
	sessionCache *sessioncache.Service,
) *Service {
	return &Service{
		cfg:          cfg,
		sessionCache: sessionCache,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (s *Service) SendNowPlaying(
	ctx context.Context,
	data models.Track,
) (NowPlayingResponse, error) {
	if err := ctx.Err(); err != nil {
		return NowPlayingResponse{}, err
	}

	session, err := s.sessionCache.Read()
	if err != nil {
		return NowPlayingResponse{}, err
	}

	form := data.ToForm()
	form.Add("format", "json")
	form.Add("method", "track.updateNowPlaying")
	form.Add("api_key", s.cfg.APIKey)
	form.Add("sk", session.Session.Key)
	form.Add("api_sig", helpers.CalculateSignature(form, s.cfg.SharedSecret))

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		constants.APIBaseURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return NowPlayingResponse{}, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := s.client.Do(request)
	if err != nil {
		return NowPlayingResponse{}, err
	}

	defer response.Body.Close()

	buff, err := io.ReadAll(response.Body)
	if err != nil {
		return NowPlayingResponse{}, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		var errResp ErrorResponse
		if err := json.Unmarshal(buff, &errResp); err != nil {
			return NowPlayingResponse{}, err
		}

		return NowPlayingResponse{}, errResp
	}

	var npResp NowPlayingResponse
	if err := json.Unmarshal(buff, &npResp); err != nil {
		return NowPlayingResponse{}, err
	}

	return npResp, nil
}

func (s *Service) Scrobble(ctx context.Context, data models.Track) (ScrobbleResponse, error) {
	if err := ctx.Err(); err != nil {
		return ScrobbleResponse{}, err
	}

	session, err := s.sessionCache.Read()
	if err != nil {
		return ScrobbleResponse{}, err
	}

	form := data.ToForm()
	form.Add("format", "json")
	form.Add("method", "track.scrobble")
	form.Add("api_key", s.cfg.APIKey)
	form.Add("sk", session.Session.Key)
	form.Add("api_sig", helpers.CalculateSignature(form, s.cfg.SharedSecret))

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		constants.APIBaseURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return ScrobbleResponse{}, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := s.client.Do(request)
	if err != nil {
		return ScrobbleResponse{}, err
	}

	defer response.Body.Close()

	buff, err := io.ReadAll(response.Body)
	if err != nil {
		return ScrobbleResponse{}, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		var errResp ErrorResponse
		if err := json.Unmarshal(buff, &errResp); err != nil {
			return ScrobbleResponse{}, err
		}

		return ScrobbleResponse{}, errResp
	}

	var scrobbleResponse ScrobbleResponse
	if err = json.Unmarshal(buff, &scrobbleResponse); err != nil {
		return ScrobbleResponse{}, err
	}

	return scrobbleResponse, nil
}
