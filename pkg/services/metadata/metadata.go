package metadata

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/dhowden/tag"
	"github.com/hcl/audioduration"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported audio format")
)

type (
	Track struct {
		Artist      string
		Name        string
		Timestamp   time.Time
		Album       string
		AlbumArtist string
		Duration    time.Duration
		Number      int
	}

	Service struct{}
)

func (t Track) ToForm() url.Values {
	form := url.Values{}

	form.Add("artist", t.Artist)
	form.Add("track", t.Name)
	form.Add("timestamp", strconv.FormatInt(t.Timestamp.UTC().Unix(), 10))

	if t.Album != "" {
		form.Add("album", t.Album)
	}

	if t.AlbumArtist != "" {
		form.Add("albumArtist", t.AlbumArtist)

		// Some files may be tagged with multiple
		// artists, but we're essentially only interested
		// in the principal artist.
		//
		// If there is an album artist, we'll use that for
		// the main artist tag.
		//
		// Is it the only way? No, but I feel it's more correct.
		if t.Artist != t.AlbumArtist {
			form.Del("artist")
			form.Add("artist", t.AlbumArtist)
		}
	}

	if t.Duration > 0 {
		seconds := int64(t.Duration.Round(time.Second).Abs().Seconds())
		form.Add("duration", strconv.FormatInt(seconds, 10))
	}

	if t.Number > 0 {
		form.Add("trackNumber", strconv.FormatInt(int64(t.Number), 10))
	}

	return form
}

func New() *Service {
	return &Service{}
}

func (s *Service) Read(buff *os.File) (Track, error) {
	data, err := tag.ReadFrom(buff)
	if err != nil {
		return Track{}, err
	}

	_, err = buff.Seek(0, 0)
	if err != nil {
		return Track{}, err
	}

	var duration float64
	switch data.FileType() {
	case tag.FLAC:
		duration, err = audioduration.FLAC(buff)
	case tag.MP3:
		duration, err = audioduration.Mp3(buff)
	case tag.OGG:
		duration, err = audioduration.Ogg(buff)
	case tag.M4A, tag.ALAC:
		duration, err = audioduration.Mp4(buff)
	default:
		err = ErrUnsupportedFormat
	}

	if err != nil {
		return Track{}, err
	}

	trackNumber, _ := data.Track()

	return Track{
		Artist:      data.Artist(),
		Name:        data.Title(),
		Timestamp:   time.Now(),
		Album:       data.Album(),
		AlbumArtist: data.AlbumArtist(),
		Duration:    time.Duration(duration * float64(time.Second)),
		Number:      trackNumber,
	}, nil
}
