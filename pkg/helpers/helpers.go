package helpers

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	replacer *strings.Replacer

	ErrInvalidDurationFormat = errors.New("invalid duration format")
)

func init() {
	replacements := []string{
		"&amp;amp;", "&",
	}

	replacer = strings.NewReplacer(replacements...)
}

func CalculateSignature(
	params url.Values,
	sharedSecret string,
) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		// The format parameter should not be a part of the signature
		if key != "format" {
			keys = append(keys, key)
		}
	}

	slices.Sort(keys)

	b := bytes.Buffer{}
	for _, key := range keys {
		b.WriteString(key)
		b.WriteString(params.Get(key))
	}

	b.WriteString(sharedSecret)

	sum := md5.Sum(b.Bytes())

	return hex.EncodeToString(sum[:])
}

func RandomID() (string, error) {
	buff := make([]byte, 20)
	_, err := rand.Read(buff)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buff), nil
}

func ParseDBDuration(duration string) (time.Duration, error) {
	parts := strings.Split(duration, ":")
	if len(parts) != 3 {
		return time.Duration(0), ErrInvalidDurationFormat
	}

	secParts := strings.Split(parts[2], ".")
	if len(secParts) != 2 {
		return time.Duration(0), ErrInvalidDurationFormat
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Duration(0), err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Duration(0), err
	}

	seconds, err := strconv.Atoi(secParts[0])
	if err != nil {
		return time.Duration(0), err
	}

	miliseconds, err := strconv.Atoi(secParts[1])
	if err != nil {
		return time.Duration(0), err
	}

	result := time.Hour * time.Duration(hours)
	result += time.Minute * time.Duration(minutes)
	result += time.Second * time.Duration(seconds)
	result += time.Millisecond * time.Duration(miliseconds)

	return result, nil
}

func ReplaceSpecialChars(in string) string {
	return replacer.Replace(in)
}
