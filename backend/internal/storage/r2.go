package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Endpoint        string
	Region          string
	PublicBaseURL   string
}

type R2Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

func NewR2Storage(ctx context.Context, cfg R2Config) (*R2Storage, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" && strings.TrimSpace(cfg.AccountID) != "" {
		cfg.Endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", strings.TrimSpace(cfg.AccountID))
	}
	if strings.TrimSpace(cfg.Region) == "" {
		cfg.Region = "auto"
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" || strings.TrimSpace(cfg.SecretAccessKey) == "" || strings.TrimSpace(cfg.Bucket) == "" || strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("missing R2 storage configuration")
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &R2Storage{client: client, presigner: s3.NewPresignClient(client), bucket: cfg.Bucket}, nil
}

func (s *R2Storage) Save(ctx context.Context, key string, r io.Reader) (StoredFile, error) {
	clean, err := safeKey(key)
	if err != nil {
		return StoredFile{}, err
	}
	objectKey := filepathKey(clean)
	contentLength := readerContentLength(r)
	h := sha256.New()
	counter := &countingReader{r: io.TeeReader(r, h)}
	out, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          counter,
		ContentLength: contentLength,
	})
	if err != nil {
		return StoredFile{}, err
	}
	stored := StoredFile{StorageKey: objectKey, ByteSize: counter.n, Checksum: hex.EncodeToString(h.Sum(nil))}
	if out.ETag != nil {
		stored.ETag = strings.Trim(*out.ETag, "\"")
	}
	if out.VersionId != nil {
		stored.VersionID = *out.VersionId
	}
	return stored, nil
}

func (s *R2Storage) Bucket() string { return s.bucket }

type countingReader struct {
	r io.Reader
	n int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.n += int64(n)
	return n, err
}

func (s *R2Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	clean, err := safeKey(key)
	if err != nil {
		return nil, err
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(filepathKey(clean))})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *R2Storage) Delete(ctx context.Context, key string) error {
	clean, err := safeKey(key)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(filepathKey(clean))})
	return err
}

func (s *R2Storage) SignedGetURL(ctx context.Context, key, filename, contentType string, ttl time.Duration) (string, error) {
	clean, err := safeKey(key)
	if err != nil {
		return "", err
	}
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filepathKey(clean)),
	}
	if strings.TrimSpace(filename) != "" {
		input.ResponseContentDisposition = aws.String(fmt.Sprintf("inline; filename=\"%s\"", strings.ReplaceAll(filename, "\"", "")))
	}
	if strings.TrimSpace(contentType) != "" {
		input.ResponseContentType = aws.String(contentType)
	}
	out, err := s.presigner.PresignGetObject(ctx, input, func(o *s3.PresignOptions) { o.Expires = ttl })
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func readerContentLength(r io.Reader) *int64 {
	switch v := r.(type) {
	case interface{ Len() int }:
		length := int64(v.Len())
		return &length
	case interface{ Size() int64 }:
		length := v.Size()
		return &length
	}
	return nil
}

func filepathKey(key string) string {
	return strings.ReplaceAll(key, "\\", "/")
}
