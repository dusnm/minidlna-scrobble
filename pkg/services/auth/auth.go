package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dusnm/minidlna-scrobble/pkg/config"
	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/helpers"
)

type (
	Service struct {
		cfg    config.Credentials
		client *http.Client
	}

	TokenResponse struct {
		Token string `json:"token"`
	}

	SessionResponse struct {
		Session struct {
			Name       string `json:"name"`
			Key        string `json:"key"`
			Subscriber uint   `json:"subscriber"`
		} `json:"session"`
	}

	ErrorResponse struct {
		Message string `json:"message"`
		Code    uint   `json:"error"`
	}
)

func (er ErrorResponse) Error() string {
	return fmt.Sprintf("request failed with: message - %s, code - %d", er.Message, er.Code)
}

func New(
	cfg config.Credentials,
) *Service {
	return &Service{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (s *Service) GetToken(ctx context.Context) (string, error) {
	query := url.Values{}
	query.Add("format", "json")
	query.Add("method", "auth.gettoken")
	query.Add("api_key", s.cfg.APIKey)
	query.Add("api_sig", helpers.CalculateSignature(query, s.cfg.SharedSecret))

	// The URL is coming from a constant, parsing can never fail
	u, _ := url.Parse(constants.APIBaseURL)

	u.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		u.String(),
		nil,
	)
	if err != nil {
		return "", err
	}

	response, err := s.client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	buff, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Cover any error
	if response.StatusCode >= http.StatusBadRequest {
		var errMsg ErrorResponse
		if err := json.Unmarshal(buff, &errMsg); err != nil {
			return "", err
		}

		return "", errMsg
	}

	var data TokenResponse
	if err = json.Unmarshal(buff, &data); err != nil {
		return "", err
	}

	return data.Token, nil
}

func (s *Service) GetSessionKey(ctx context.Context, token string) (SessionResponse, error) {
	query := url.Values{}
	query.Add("format", "json")
	query.Add("method", "auth.getsession")
	query.Add("token", token)
	query.Add("api_key", s.cfg.APIKey)
	query.Add("api_sig", helpers.CalculateSignature(query, s.cfg.SharedSecret))

	// The URL is coming from a constant, parsing can never fail
	u, _ := url.Parse(constants.APIBaseURL)

	u.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		u.String(),
		nil,
	)
	if err != nil {
		return SessionResponse{}, err
	}

	response, err := s.client.Do(request)
	if err != nil {
		return SessionResponse{}, err
	}

	defer response.Body.Close()

	buff, err := io.ReadAll(response.Body)
	if err != nil {
		return SessionResponse{}, err
	}

	// Cover any error
	if response.StatusCode >= http.StatusBadRequest {
		var errMsg ErrorResponse
		if err := json.Unmarshal(buff, &errMsg); err != nil {
			return SessionResponse{}, err
		}

		return SessionResponse{}, errMsg
	}

	var data SessionResponse
	if err = json.Unmarshal(buff, &data); err != nil {
		return SessionResponse{}, err
	}

	return data, nil
}
