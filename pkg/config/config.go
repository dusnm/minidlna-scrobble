package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dusnm/minidlna-scrobble/pkg/constants"
)

var (
	ErrDBFilePathNotAbsolute  = errors.New("the path to the minidlna database must be absolute")
	ErrLogFilePathNotAbsolute = errors.New("the path to the minidlna log file must be absolute")
	ErrAPIKeyMissing          = errors.New("you must supply the api key")
	ErrSharedSecretMissing    = errors.New("you must supply the shared secret")
)

type (
	ErrConfigFileNotFound struct {
		Path string
	}

	Credentials struct {
		APIKey       string `json:"api_key"`
		SharedSecret string `json:"shared_secret"`
	}

	Config struct {
		DBFile      string      `json:"db_file"`
		LogFile     string      `json:"log_file"`
		Credentials Credentials `json:"credentials"`
	}
)

func (e ErrConfigFileNotFound) Error() string {
	return fmt.Sprintf("config file not found at: %s", e.Path)
}

func New() (*Config, error) {
	configDir := "/etc"
	v, set := os.LookupEnv(constants.XDGConfigDir)
	if set && v != "" && filepath.IsAbs(v) {
		configDir = v
	}

	configPath := filepath.Join(configDir, "minidlna-scrobbler", "config.json")
	f, err := os.OpenFile(configPath, os.O_RDONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrConfigFileNotFound{Path: configPath}
		}

		return nil, err
	}

	defer f.Close()

	cfg, err := unmarshall(f)
	if err != nil {
		return nil, err
	}

	if err = validate(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func unmarshall(data io.Reader) (Config, error) {
	var cfg Config
	decoder := json.NewDecoder(data)
	for {
		if err := decoder.Decode(&cfg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return Config{}, err
		}
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if !filepath.IsAbs(cfg.DBFile) {
		return ErrDBFilePathNotAbsolute
	}

	if !filepath.IsAbs(cfg.LogFile) {
		return ErrLogFilePathNotAbsolute
	}

	if cfg.Credentials.APIKey == "" {
		return ErrAPIKeyMissing
	}

	if cfg.Credentials.SharedSecret == "" {
		return ErrSharedSecretMissing
	}

	return nil
}
