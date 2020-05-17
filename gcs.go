package hotconfig

import (
	"context"
	"errors"
	"strings"

	"cloud.google.com/go/storage"
)

// fetcher to fetch config from google cloud storage
type gcsFetcher struct {
	gcs     *storage.Client
	bucket  string
	key     string
	decoder Decoder
}

// parse a google cloud storage url (e.g. gs://bucket/key)
func ParseGCSUrl(url string) (string, string, error) {
	const prefix = "gs://"
	if !strings.HasPrefix(url, prefix) {
		return "", "", errors.New("invalid gs url")
	}
	url = url[len(prefix):]
	i := strings.IndexByte(url, '/')
	if i < 0 {
		return "", "", errors.New("invalid gs url")
	}
	return url[:i], url[i+1:], nil
}

// create a new GCSFetcher from gs url
func NewGCSFetcherFromURI(sc *storage.Client, uri string, decoder Decoder) (Fetcher, error) {
	bucket, key, err := ParseGCSUrl(uri)
	if err != nil {
		return nil, err
	}
	return NewGCSFetcher(sc, bucket, key, decoder), nil
}

// create a new GCSFetcher from bucket and key
func NewGCSFetcher(sc *storage.Client, bucket, key string, decoder Decoder) Fetcher {
	return &gcsFetcher{sc, bucket, key, decoder}
}

// fetch config from gcs
func (f *gcsFetcher) Fetch(ctx context.Context) (interface{}, error) {
	rd, err := f.gcs.Bucket(f.bucket).Object(f.bucket).Key([]byte(f.key)).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	return f.decoder.Decode(rd)
}

// new hot-reloadable config from gcs
func NewGCSConfig(ctx context.Context, sc *storage.Client, uri string, decoder Decoder) (*Config, error) {
	fetcher, err := NewGCSFetcherFromURI(sc, uri, decoder)
	if err != nil {
		return nil, err
	}
	return NewConfig(ctx, fetcher), nil
}
