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
	"sort"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/database/money"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/storage"
	"github.com/google/uuid"
)

var ErrMoneyForbidden = errors.New("forbidden")
var ErrMoneyInvalidInput = errors.New("invalid input")

const maxGeoJSONBytes = 512 * 1024
const moneyReferenceImportSource = "new_money_reference"

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

func (s *MoneyService) CragSnapshot(ctx context.Context, projectID string) (*models.MoneyCragSnapshot, error) {
	p, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	features, err := s.repo.ListFeatures(ctx, projectID, models.MoneyFeatureFilter{})
	if err != nil {
		return nil, err
	}
	notes, err := s.repo.ListNotesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	uploads, err := s.repo.ListUploadsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	root, trails := BuildMoneyCragTree(features)
	return &models.MoneyCragSnapshot{Project: *p, Root: root, Trails: trails, Notes: notes, Uploads: uploads}, nil
}

func BuildMoneyCragTree(features []models.MoneyFeature) (*models.MoneyCragNode, []models.MoneyCragNode) {
	byID := make(map[string]models.MoneyFeature, len(features))
	children := map[string][]string{}
	var rootIDs []string
	var trailIDs []string
	for _, f := range features {
		byID[f.ID] = f
	}
	for _, f := range features {
		if f.FeatureType == models.MoneyFeatureTrail {
			trailIDs = append(trailIDs, f.ID)
			continue
		}
		if f.ParentFeatureID == nil || *f.ParentFeatureID == "" || byID[*f.ParentFeatureID].ID == "" {
			rootIDs = append(rootIDs, f.ID)
			continue
		}
		children[*f.ParentFeatureID] = append(children[*f.ParentFeatureID], f.ID)
	}
	var materialize func(id string) models.MoneyCragNode
	materialize = func(id string) models.MoneyCragNode {
		n := models.MoneyCragNode{Feature: byID[id]}
		for _, childID := range children[id] {
			child := materialize(childID)
			switch child.Feature.FeatureType {
			case models.MoneyFeatureArea:
				n.Children = append(n.Children, child)
			case models.MoneyFeatureBoulder:
				n.Boulders = append(n.Boulders, child)
			case models.MoneyFeatureProblem:
				n.Problems = append(n.Problems, child)
			default:
				n.Children = append(n.Children, child)
			}
		}
		return n
	}
	var root *models.MoneyCragNode
	for _, id := range rootIDs {
		n := materialize(id)
		if n.Feature.FeatureType == models.MoneyFeatureArea && root == nil {
			root = &n
		}
	}
	if root == nil && len(rootIDs) > 0 {
		n := materialize(rootIDs[0])
		root = &n
	}
	trails := make([]models.MoneyCragNode, 0, len(trailIDs))
	for _, id := range trailIDs {
		trails = append(trails, materialize(id))
	}
	return root, trails
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

func (s *MoneyService) CreateArea(ctx context.Context, projectID string, req models.MoneyCragAreaRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	return s.CreateFeature(ctx, projectID, models.MoneyFeatureRequest{ParentFeatureID: req.ParentFeatureID, FeatureType: models.MoneyFeatureArea, Title: req.Title, Description: req.Description, Status: models.MoneyStatusActive, GeoJSON: req.GeoJSON, Properties: req.Properties}, user)
}

func (s *MoneyService) CreateBoulder(ctx context.Context, projectID string, req models.MoneyCragBoulderRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	status := req.DevStatus
	if status == "" {
		status = models.MoneyStatusScouted
	}
	return s.CreateFeature(ctx, projectID, models.MoneyFeatureRequest{ParentFeatureID: &req.ParentFeatureID, FeatureType: models.MoneyFeatureBoulder, Title: req.Title, Description: req.Description, Status: status, GeoJSON: req.GeoJSON, Properties: req.Properties}, user)
}

func (s *MoneyService) CreateProblem(ctx context.Context, projectID string, req models.MoneyCragProblemRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	props := map[string]interface{}{"grade": strings.TrimSpace(req.Grade), "stars": req.Stars, "types": req.Types}
	if req.FA != nil {
		props["fa"] = strings.TrimSpace(*req.FA)
	}
	if len(req.Properties) > 0 {
		var extra map[string]interface{}
		if json.Unmarshal(req.Properties, &extra) == nil {
			for k, v := range extra {
				props[k] = v
			}
		}
	}
	rawProps, _ := json.Marshal(props)
	return s.CreateFeature(ctx, projectID, models.MoneyFeatureRequest{ParentFeatureID: &req.BoulderID, FeatureType: models.MoneyFeatureProblem, Title: req.Name, Description: req.Description, Status: req.Status, GeoJSON: pointGeoJSONFromProperties(rawProps), Properties: rawProps}, user)
}

func (s *MoneyService) UpdateBoulderStatus(ctx context.Context, id string, req models.MoneyBoulderStatusRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	if !validBoulderStatus(req.DevStatus) {
		return nil, ErrMoneyInvalidInput
	}
	f, err := s.repo.GetFeature(ctx, id)
	if err != nil {
		return nil, err
	}
	if f.FeatureType != models.MoneyFeatureBoulder {
		return nil, ErrMoneyInvalidInput
	}
	f.Status = req.DevStatus
	f.UpdatedBy = user.ID
	return s.repo.UpdateFeature(ctx, *f)
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

func (s *MoneyService) ListProjectNotes(ctx context.Context, projectID string) ([]models.MoneyNote, error) {
	return s.repo.ListNotesByProject(ctx, projectID)
}

func (s *MoneyService) CreateProjectNote(ctx context.Context, projectID string, req models.MoneyNoteRequest, user models.CurrentUser) (*models.MoneyNote, error) {
	return s.createNote(ctx, projectID, nil, req, user)
}

func (s *MoneyService) CreateNote(ctx context.Context, featureID string, req models.MoneyNoteRequest, user models.CurrentUser) (*models.MoneyNote, error) {
	feature, err := s.repo.GetFeature(ctx, featureID)
	if err != nil {
		return nil, err
	}
	return s.createNote(ctx, feature.ProjectID, &featureID, req, user)
}

func (s *MoneyService) createNote(ctx context.Context, projectID string, featureID *string, req models.MoneyNoteRequest, user models.CurrentUser) (*models.MoneyNote, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	body := strings.TrimSpace(req.Body)
	if body == "" && len(req.Blocks) <= 2 {
		return nil, ErrMoneyInvalidInput
	}
	if len(body) > 5000 {
		return nil, ErrMoneyInvalidInput
	}
	visibility := req.Visibility
	if visibility == "" {
		visibility = models.MoneyNoteTeam
	}
	if visibility != models.MoneyNoteTeam && visibility != models.MoneyNotePrivate {
		return nil, ErrMoneyInvalidInput
	}
	targetType := req.TargetType
	if targetType == "" {
		targetType = "feature"
	}
	if !validNoteTarget(targetType) {
		return nil, ErrMoneyInvalidInput
	}
	blocks := req.Blocks
	if len(blocks) == 0 {
		blocks = json.RawMessage(`[]`)
	}
	if !json.Valid(blocks) || len(blocks) > 128*1024 {
		return nil, ErrMoneyInvalidInput
	}
	return s.repo.CreateNote(ctx, models.MoneyNote{ProjectID: projectID, FeatureID: featureID, TargetType: targetType, TargetRef: req.TargetRef, Body: body, Visibility: visibility, Tags: cleanTags(req.Tags), Blocks: blocks, CreatedBy: user.ID, UpdatedBy: user.ID})
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
	return s.StoreUploadWithKind(ctx, projectID, featureID, noteID, "photo", nil, fh, user)
}

func (s *MoneyService) StoreUploadWithKind(ctx context.Context, projectID string, featureID, noteID *string, blockKind string, metadata json.RawMessage, fh *multipart.FileHeader, user models.CurrentUser) (*models.MoneyUpload, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	if blockKind == "" {
		blockKind = "photo"
	}
	if !validBlockKind(blockKind) {
		return nil, ErrMoneyInvalidInput
	}
	if len(metadata) == 0 || strings.TrimSpace(string(metadata)) == "" {
		metadata = json.RawMessage(`{}`)
	}
	if !json.Valid(metadata) {
		return nil, ErrMoneyInvalidInput
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
	if blockKind != "file" && !allowedImage(contentType) {
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
	return s.repo.CreateUpload(ctx, models.MoneyUpload{ID: uploadID, ProjectID: projectID, FeatureID: featureID, NoteID: noteID, OriginalFilename: safe, StorageKey: stored.StorageKey, ContentType: contentType, ByteSize: stored.ByteSize, Width: width, Height: height, ChecksumSHA256: stored.Checksum, BlockKind: blockKind, Metadata: metadata, UploadedBy: user.ID})
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
		status = defaultStatus(req.FeatureType)
	}
	if !validFeatureStatus(req.FeatureType, status) {
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
	if !json.Valid(style) || !json.Valid(props) {
		return models.MoneyFeature{}, ErrMoneyInvalidInput
	}
	bbox, err := ValidateGeoJSON(req.GeoJSON)
	if err != nil {
		return models.MoneyFeature{}, err
	}
	return models.MoneyFeature{ParentFeatureID: req.ParentFeatureID, FeatureType: req.FeatureType, Title: strings.TrimSpace(req.Title), Description: cleanOptional(req.Description, 2000), Status: status, GeoJSON: req.GeoJSON, Style: style, Properties: props, MinLat: &bbox.MinLat, MinLon: &bbox.MinLon, MaxLat: &bbox.MaxLat, MaxLon: &bbox.MaxLon, SortOrder: req.SortOrder, ExternalRef: req.ExternalRef, ImportSource: req.ImportSource}, nil
}

func ValidateGeoJSON(raw json.RawMessage) (*models.BBox, error) {
	if len(raw) == 0 || len(raw) > maxGeoJSONBytes {
		return nil, ErrMoneyInvalidInput
	}
	var obj map[string]interface{}
	if json.Unmarshal(raw, &obj) != nil {
		return nil, ErrMoneyInvalidInput
	}
	bbox := &models.BBox{MinLat: 1e9, MinLon: 1e9, MaxLat: -1e9, MaxLon: -1e9}
	if err := walkGeoJSON(obj, bbox); err != nil {
		return nil, err
	}
	if bbox.MinLat == 1e9 {
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
		x, ok1 := arr[0].(float64)
		y, ok2 := arr[1].(float64)
		if ok1 && ok2 {
			validLonLat := x >= -180 && x <= 180 && y >= -90 && y <= 90
			validWorld := x >= -100 && x <= 1100 && y >= -100 && y <= 800
			if !validLonLat && !validWorld {
				return ErrMoneyInvalidInput
			}
			bbox.MinLon = minFloat(bbox.MinLon, x)
			bbox.MaxLon = maxFloat(bbox.MaxLon, x)
			bbox.MinLat = minFloat(bbox.MinLat, y)
			bbox.MaxLat = maxFloat(bbox.MaxLat, y)
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
	return t == models.MoneyFeatureArea || t == models.MoneyFeatureBoulder || t == models.MoneyFeatureProblem || t == models.MoneyFeatureTrail || t == models.MoneyFeatureTopo || t == models.MoneyFeaturePOI || t == models.MoneyFeatureDrawing
}
func validFeatureStatus(featureType, status string) bool {
	switch featureType {
	case models.MoneyFeatureBoulder:
		return validBoulderStatus(status)
	case models.MoneyFeatureProblem:
		return status == models.MoneyStatusProject || status == models.MoneyStatusSent || status == models.MoneyStatusEstablished || status == models.MoneyStatusDraft || status == models.MoneyStatusArchived
	default:
		return status == models.MoneyStatusDraft || status == models.MoneyStatusActive || status == models.MoneyStatusArchived
	}
}
func validBoulderStatus(status string) bool {
	return status == models.MoneyStatusScouted || status == models.MoneyStatusNeedsWork || status == models.MoneyStatusCleaning || status == models.MoneyStatusEstablished || status == models.MoneyStatusArchived
}
func defaultStatus(featureType string) string {
	if featureType == models.MoneyFeatureBoulder {
		return models.MoneyStatusScouted
	}
	if featureType == models.MoneyFeatureProblem {
		return models.MoneyStatusProject
	}
	return models.MoneyStatusActive
}
func validNoteTarget(t string) bool {
	return t == "project" || t == "feature" || t == "area" || t == "boulder" || t == "trail" || t == "point" || t == "none"
}
func validBlockKind(k string) bool {
	return k == "photo" || k == "sketch" || k == "file" || k == "topo"
}
func cleanTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, tag := range tags {
		v := strings.Trim(strings.ToLower(strings.TrimSpace(tag)), "#")
		if v == "" || seen[v] || len(v) > 40 {
			continue
		}
		seen[v] = true
		out = append(out, v)
		if len(out) >= 12 {
			break
		}
	}
	sort.Strings(out)
	return out
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
func pointGeoJSONFromProperties(_ json.RawMessage) json.RawMessage {
	return json.RawMessage(`{"type":"Point","coordinates":[0,0]}`)
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
