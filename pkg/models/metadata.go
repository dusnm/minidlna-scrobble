package models

import (
	"net/url"
	"strconv"
	"time"
)

type (
	Track struct {
		Artist    string
		Name      string
		Timestamp time.Time
		Album     string
		Duration  time.Duration
		Number    int
	}
)

func (t Track) ToForm() url.Values {
	form := url.Values{}

	form.Add("artist", t.Artist)
	form.Add("track", t.Name)
	form.Add("timestamp", strconv.FormatInt(t.Timestamp.UTC().Unix(), 10))

	if t.Album != "" {
		form.Add("album", t.Album)
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
