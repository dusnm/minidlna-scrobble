package sessioncache

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/services/auth"
)

type (
	Service struct {
		dir string
	}
)

func New() (*Service, error) {
	cacheDir := "/var/cache"
	v, set := os.LookupEnv(constants.XDGCacheDIR)
	if set && v != "" && filepath.IsAbs(v) {
		cacheDir = v
	}

	cacheDir = filepath.Join(cacheDir, "minidlna-scrobbler")
	_, err := os.Stat(cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cacheDir, 0o744)
			if err != nil {
				return nil, err
			}
		}

		return nil, err
	}

	return &Service{
		dir: cacheDir,
	}, nil
}

func (s *Service) Save(data auth.SessionResponse) error {
	fPath := filepath.Join(s.dir, "session.json")
	f, err := os.OpenFile(fPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = f.Write(buff)

	return err
}

func (s *Service) Read() (auth.SessionResponse, error) {
	fPath := filepath.Join(s.dir, "session.json")
	f, err := os.OpenFile(fPath, os.O_RDONLY, 0o644)
	if err != nil {
		return auth.SessionResponse{}, err
	}

	defer f.Close()

	buff, err := io.ReadAll(f)
	if err != nil {
		return auth.SessionResponse{}, err
	}

	var data auth.SessionResponse
	err = json.Unmarshal(buff, &data)
	if err != nil {
		return auth.SessionResponse{}, err
	}

	return data, nil
}
