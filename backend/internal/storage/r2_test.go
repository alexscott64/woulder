package storage

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewR2StorageValidatesRequiredConfigWithoutSecrets(t *testing.T) {
	_, err := NewR2Storage(context.Background(), R2Config{AccountID: "account", Bucket: "woulder"})
	if err == nil {
		t.Fatal("expected missing configuration error")
	}
	if strings.Contains(err.Error(), "account") || strings.Contains(err.Error(), "woulder") {
		t.Fatalf("configuration error should not include provided values: %v", err)
	}
}

func TestNewR2StorageDerivesEndpointAndBucket(t *testing.T) {
	store, err := NewR2Storage(context.Background(), R2Config{AccountID: "account-id", AccessKeyID: "access-key", SecretAccessKey: "secret-key", Bucket: "woulder"})
	if err != nil {
		t.Fatalf("NewR2Storage returned error: %v", err)
	}
	if store.Bucket() != "woulder" {
		t.Fatalf("unexpected bucket: %s", store.Bucket())
	}
}

func TestR2SignedGetURLRejectsUnsafeKey(t *testing.T) {
	store, err := NewR2Storage(context.Background(), R2Config{AccountID: "account-id", AccessKeyID: "access-key", SecretAccessKey: "secret-key", Bucket: "woulder"})
	if err != nil {
		t.Fatalf("NewR2Storage returned error: %v", err)
	}
	if _, err := store.SignedGetURL(context.Background(), "../escape.jpg", "escape.jpg", "image/jpeg", time.Minute); err == nil {
		t.Fatal("expected unsafe key to be rejected")
	}
}

func TestReaderContentLengthDetectsBufferedUpload(t *testing.T) {
	r := bytes.NewReader([]byte("tiny upload"))
	got := readerContentLength(r)
	if got == nil || *got != int64(r.Len()) {
		t.Fatalf("expected content length %d, got %v", r.Len(), got)
	}
}
