package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage { return &LocalStorage{baseDir: baseDir} }

func (s *LocalStorage) Save(ctx context.Context, key string, r io.Reader) (StoredFile, error) {
	if err := ctx.Err(); err != nil {
		return StoredFile{}, err
	}
	clean, err := safeKey(key)
	if err != nil {
		return StoredFile{}, err
	}
	path := filepath.Join(s.baseDir, clean)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return StoredFile{}, err
	}
	f, err := os.Create(path)
	if err != nil {
		return StoredFile{}, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), r)
	if err != nil {
		return StoredFile{}, err
	}
	return StoredFile{StorageKey: clean, ByteSize: n, Checksum: hex.EncodeToString(h.Sum(nil))}, nil
}

func (s *LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	clean, err := safeKey(key)
	if err != nil {
		return nil, err
	}
	return os.Open(filepath.Join(s.baseDir, clean))
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	clean, err := safeKey(key)
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(s.baseDir, clean))
}

func safeKey(key string) (string, error) {
	key = filepath.ToSlash(strings.TrimSpace(key))
	if key == "" || strings.Contains(key, "..") || strings.HasPrefix(key, "/") || filepath.IsAbs(key) {
		return "", fmt.Errorf("invalid storage key")
	}
	return filepath.FromSlash(key), nil
}
