package storage

import (
	"context"
	"io"
	"time"
)

type StoredFile struct {
	StorageKey string
	ByteSize   int64
	Checksum   string
	ETag       string
	VersionID  string
}

type Storage interface {
	Save(ctx context.Context, key string, r io.Reader) (StoredFile, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type DownloadPresigner interface {
	SignedGetURL(ctx context.Context, key, filename, contentType string, ttl time.Duration) (string, error)
}
