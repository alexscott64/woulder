package storage

import (
	"context"
	"io"
)

type StoredFile struct {
	StorageKey string
	ByteSize   int64
	Checksum   string
}

type Storage interface {
	Save(ctx context.Context, key string, r io.Reader) (StoredFile, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
