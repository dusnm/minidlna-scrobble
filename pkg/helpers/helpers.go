package helpers

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"slices"
)

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
