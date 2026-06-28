package storage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestLocalStorageSaveOpenAndChecksum(t *testing.T) {
	store := NewLocalStorage(t.TempDir())
	stored, err := store.Save(context.Background(), "money/project/upload/file.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if stored.ByteSize != 5 || stored.Checksum != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("unexpected stored metadata: %+v", stored)
	}
	r, err := store.Open(context.Background(), stored.StorageKey)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	if string(b) != "hello" {
		t.Fatalf("unexpected content: %q", string(b))
	}
}

func TestLocalStorageRejectsUnsafeKey(t *testing.T) {
	store := NewLocalStorage(t.TempDir())
	if _, err := store.Save(context.Background(), "../escape.txt", strings.NewReader("x")); err == nil {
		t.Fatal("expected unsafe key error")
	}
}
