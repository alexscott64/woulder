package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/database/money"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/storage"
	"github.com/google/uuid"
)

var ErrMoneyForbidden = errors.New("forbidden")
var ErrMoneyInvalidInput = errors.New("invalid input")

const maxGeoJSONBytes = 256 * 1024

type MoneyService struct {
	repo      money.Repository
	storage   storage.Storage
	maxUpload int64
}

func NewMoneyService(repo money.Repository, store storage.Storage, maxUploadBytes int64) *MoneyService {
	return &MoneyService{repo: repo, storage: store, maxUpload: maxUploadBytes}
}

func (s *MoneyService) GetProjectBySlug(ctx context.Context, slug string, user models.CurrentUser) (*models.MoneyProjectResponse, error) {
	p, err := s.repo.GetProjectBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &models.MoneyProjectResponse{Project: *p, User: user, Permissions: permissions(user.Role)}, nil
}

func (s *MoneyService) Snapshot(ctx context.Context, projectID string) (*models.MoneySnapshot, error) {
	p, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	features, err := s.repo.ListFeatures(ctx, projectID, models.MoneyFeatureFilter{})
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.FeatureNoteCounts(ctx, projectID)
	if err != nil {
		return nil, err
	}
	uploads, err := s.repo.PrimaryUploads(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &models.MoneySnapshot{Project: *p, Features: features, NoteCounts: counts, PrimaryUploads: uploads}, nil
}

func (s *MoneyService) ListFeatures(ctx context.Context, projectID string, filter models.MoneyFeatureFilter) ([]models.MoneyFeature, error) {
	return s.repo.ListFeatures(ctx, projectID, filter)
}

func (s *MoneyService) GetFeatureDetail(ctx context.Context, id string) (*models.MoneyFeatureDetail, error) {
	f, err := s.repo.GetFeature(ctx, id)
	if err != nil {
		return nil, err
	}
	notes, err := s.repo.ListNotes(ctx, id)
	if err != nil {
		return nil, err
	}
	uploads, err := s.repo.ListUploadsByFeature(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.MoneyFeatureDetail{Feature: *f, Notes: notes, Uploads: uploads}, nil
}

func (s *MoneyService) CreateFeature(ctx context.Context, projectID string, req models.MoneyFeatureRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	f, err := s.featureFromRequest(req)
	if err != nil {
		return nil, err
	}
	f.ProjectID = projectID
	f.CreatedBy = user.ID
	f.UpdatedBy = user.ID
	return s.repo.CreateFeature(ctx, f)
}

func (s *MoneyService) UpdateFeature(ctx context.Context, id string, req models.MoneyFeatureRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	f, err := s.featureFromRequest(req)
	if err != nil {
		return nil, err
	}
	f.ID = id
	f.UpdatedBy = user.ID
	return s.repo.UpdateFeature(ctx, f)
}

func (s *MoneyService) ArchiveFeature(ctx context.Context, id string, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	return s.repo.ArchiveFeature(ctx, id, user.ID)
}

func (s *MoneyService) ListNotes(ctx context.Context, featureID string) ([]models.MoneyNote, error) {
	return s.repo.ListNotes(ctx, featureID)
}

func (s *MoneyService) CreateNote(ctx context.Context, featureID string, req models.MoneyNoteRequest, user models.CurrentUser) (*models.MoneyNote, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	feature, err := s.repo.GetFeature(ctx, featureID)
	if err != nil {
		return nil, err
	}
	body := strings.TrimSpace(req.Body)
	if body == "" || len(body) > 5000 {
		return nil, ErrMoneyInvalidInput
	}
	visibility := req.Visibility
	if visibility == "" {
		visibility = models.MoneyNoteTeam
	}
	if visibility != models.MoneyNoteTeam && visibility != models.MoneyNotePrivate {
		return nil, ErrMoneyInvalidInput
	}
	return s.repo.CreateNote(ctx, models.MoneyNote{ProjectID: feature.ProjectID, FeatureID: &featureID, Body: body, Visibility: visibility, CreatedBy: user.ID, UpdatedBy: user.ID})
}

func (s *MoneyService) UpdateNote(ctx context.Context, noteID string, req models.MoneyNoteRequest, user models.CurrentUser) (*models.MoneyNote, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	body := strings.TrimSpace(req.Body)
	if body == "" || len(body) > 5000 {
		return nil, ErrMoneyInvalidInput
	}
	visibility := req.Visibility
	if visibility == "" {
		visibility = models.MoneyNoteTeam
	}
	return s.repo.UpdateNote(ctx, noteID, body, visibility, user.ID, user.Role)
}

func (s *MoneyService) DeleteNote(ctx context.Context, noteID string, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	return s.repo.DeleteNote(ctx, noteID, user.ID, user.Role)
}

func (s *MoneyService) StoreUpload(ctx context.Context, projectID string, featureID, noteID *string, fh *multipart.FileHeader, user models.CurrentUser) (*models.MoneyUpload, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	if fh == nil || fh.Size <= 0 || fh.Size > s.maxUpload {
		return nil, ErrMoneyInvalidInput
	}
	file, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	limited := io.LimitReader(file, s.maxUpload+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(buf)) > s.maxUpload {
		return nil, ErrMoneyInvalidInput
	}
	contentType := http.DetectContentType(buf[:min(len(buf), 512)])
	if !allowedImage(contentType) {
		return nil, ErrMoneyInvalidInput
	}
	width, height := imageDimensions(buf)
	uploadID := uuid.NewString()
	safe := safeFilename(fh.Filename, contentType)
	key := filepath.ToSlash(filepath.Join("money", projectID, uploadID, safe))
	stored, err := s.storage.Save(ctx, key, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	return s.repo.CreateUpload(ctx, models.MoneyUpload{ID: uploadID, ProjectID: projectID, FeatureID: featureID, NoteID: noteID, OriginalFilename: safe, StorageKey: stored.StorageKey, ContentType: contentType, ByteSize: stored.ByteSize, Width: width, Height: height, ChecksumSHA256: stored.Checksum, UploadedBy: user.ID})
}

func (s *MoneyService) OpenUpload(ctx context.Context, uploadID string) (*models.MoneyUpload, io.ReadCloser, error) {
	u, err := s.repo.GetUpload(ctx, uploadID)
	if err != nil {
		return nil, nil, err
	}
	r, err := s.storage.Open(ctx, u.StorageKey)
	if err != nil {
		return nil, nil, err
	}
	return u, r, nil
}

func (s *MoneyService) DeleteUpload(ctx context.Context, uploadID string, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	u, err := s.repo.GetUpload(ctx, uploadID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteUpload(ctx, uploadID, user.ID, user.Role); err != nil {
		return err
	}
	_ = s.storage.Delete(ctx, u.StorageKey)
	return nil
}

func (s *MoneyService) featureFromRequest(req models.MoneyFeatureRequest) (models.MoneyFeature, error) {
	if !validFeatureType(req.FeatureType) || strings.TrimSpace(req.Title) == "" || len(req.Title) > 200 {
		return models.MoneyFeature{}, ErrMoneyInvalidInput
	}
	status := req.Status
	if status == "" {
		status = models.MoneyStatusDraft
	}
	if status != models.MoneyStatusDraft && status != models.MoneyStatusActive && status != models.MoneyStatusArchived {
		return models.MoneyFeature{}, ErrMoneyInvalidInput
	}
	style := req.Style
	if len(style) == 0 {
		style = json.RawMessage(`{}`)
	}
	props := req.Properties
	if len(props) == 0 {
		props = json.RawMessage(`{}`)
	}
	bbox, err := ValidateGeoJSON(req.GeoJSON)
	if err != nil {
		return models.MoneyFeature{}, err
	}
	return models.MoneyFeature{FeatureType: req.FeatureType, Title: strings.TrimSpace(req.Title), Description: cleanOptional(req.Description, 2000), Status: status, GeoJSON: req.GeoJSON, Style: style, Properties: props, MinLat: &bbox.MinLat, MinLon: &bbox.MinLon, MaxLat: &bbox.MaxLat, MaxLon: &bbox.MaxLon}, nil
}

func ValidateGeoJSON(raw json.RawMessage) (*models.BBox, error) {
	if len(raw) == 0 || len(raw) > maxGeoJSONBytes {
		return nil, ErrMoneyInvalidInput
	}
	var obj map[string]interface{}
	if json.Unmarshal(raw, &obj) != nil {
		return nil, ErrMoneyInvalidInput
	}
	bbox := &models.BBox{MinLat: 91, MinLon: 181, MaxLat: -91, MaxLon: -181}
	if err := walkGeoJSON(obj, bbox); err != nil {
		return nil, err
	}
	if bbox.MinLat == 91 {
		return nil, ErrMoneyInvalidInput
	}
	return bbox, nil
}

func walkGeoJSON(obj map[string]interface{}, bbox *models.BBox) error {
	t, _ := obj["type"].(string)
	switch t {
	case "Point", "LineString", "Polygon":
		return walkCoords(obj["coordinates"], bbox)
	case "Feature":
		g, ok := obj["geometry"].(map[string]interface{})
		if !ok {
			return ErrMoneyInvalidInput
		}
		return walkGeoJSON(g, bbox)
	case "FeatureCollection":
		arr, ok := obj["features"].([]interface{})
		if !ok {
			return ErrMoneyInvalidInput
		}
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				return ErrMoneyInvalidInput
			}
			if err := walkGeoJSON(m, bbox); err != nil {
				return err
			}
		}
		return nil
	default:
		return ErrMoneyInvalidInput
	}
}

func walkCoords(v interface{}, bbox *models.BBox) error {
	arr, ok := v.([]interface{})
	if !ok {
		return ErrMoneyInvalidInput
	}
	if len(arr) >= 2 {
		lon, ok1 := arr[0].(float64)
		lat, ok2 := arr[1].(float64)
		if ok1 && ok2 {
			if lon < -180 || lon > 180 || lat < -90 || lat > 90 {
				return ErrMoneyInvalidInput
			}
			bbox.MinLon = minFloat(bbox.MinLon, lon)
			bbox.MaxLon = maxFloat(bbox.MaxLon, lon)
			bbox.MinLat = minFloat(bbox.MinLat, lat)
			bbox.MaxLat = maxFloat(bbox.MaxLat, lat)
			return nil
		}
	}
	for _, child := range arr {
		if err := walkCoords(child, bbox); err != nil {
			return err
		}
	}
	return nil
}

func permissions(role string) models.MoneyPermissions {
	return models.MoneyPermissions{CanRead: true, CanWrite: models.CanWriteMoney(role), IsAdmin: role == models.RoleAdmin}
}
func validFeatureType(t string) bool {
	return t == models.MoneyFeatureTrail || t == models.MoneyFeatureTopo || t == models.MoneyFeaturePOI || t == models.MoneyFeatureDrawing
}
func cleanOptional(s *string, max int) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if len(v) > max {
		v = v[:max]
	}
	if v == "" {
		return nil
	}
	return &v
}
func allowedImage(ct string) bool {
	return ct == "image/jpeg" || ct == "image/png" || ct == "image/webp"
}
func imageDimensions(buf []byte) (*int, *int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		return nil, nil
	}
	return &cfg.Width, &cfg.Height
}
func safeFilename(name, ct string) string {
	base := regexp.MustCompile(`[^a-zA-Z0-9._-]+`).ReplaceAllString(filepath.Base(name), "-")
	base = strings.Trim(base, ".-")
	if base == "" {
		base = "upload"
	}
	if filepath.Ext(base) == "" {
		base += extForContentType(ct)
	}
	return base
}
func extForContentType(ct string) string {
	if ct == "image/png" {
		return ".png"
	}
	if ct == "image/webp" {
		return ".webp"
	}
	return ".jpg"
}
func ParseBBox(s string) (*models.BBox, error) {
	var b models.BBox
	if _, err := fmt.Sscanf(s, "%f,%f,%f,%f", &b.MinLon, &b.MinLat, &b.MaxLon, &b.MaxLat); err != nil {
		return nil, ErrMoneyInvalidInput
	}
	return &b, nil
}
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
