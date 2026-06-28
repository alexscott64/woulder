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
	"time"

	"github.com/alexscott64/woulder/backend/internal/database/money"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/storage"
	"github.com/google/uuid"
)

var ErrMoneyForbidden = errors.New("forbidden")
var ErrMoneyInvalidInput = errors.New("invalid input")

const maxGeoJSONBytes = 512 * 1024
const moneyReferenceImportSource = "money_reference"

type MoneyService struct {
	repo           money.Repository
	storage        storage.Storage
	maxUpload      int64
	storageBackend string
	storageBucket  string
	storageRegion  string
	keyPrefix      string
	signedURLTTL   time.Duration
}

type MoneyServiceOptions struct {
	StorageBackend string
	StorageBucket  string
	StorageRegion  string
	KeyPrefix      string
	SignedURLTTL   time.Duration
}

func NewMoneyService(repo money.Repository, store storage.Storage, maxUploadBytes int64) *MoneyService {
	return NewMoneyServiceWithOptions(repo, store, maxUploadBytes, MoneyServiceOptions{StorageBackend: "local", KeyPrefix: "money-creek", SignedURLTTL: 5 * time.Minute})
}

func NewMoneyServiceWithOptions(repo money.Repository, store storage.Storage, maxUploadBytes int64, opts MoneyServiceOptions) *MoneyService {
	if opts.StorageBackend == "" {
		opts.StorageBackend = "local"
	}
	if opts.KeyPrefix == "" {
		opts.KeyPrefix = "money-creek"
	}
	opts.KeyPrefix = normalizeAssetKeyPrefix(opts.KeyPrefix)
	if opts.SignedURLTTL <= 0 {
		opts.SignedURLTTL = 5 * time.Minute
	}
	return &MoneyService{repo: repo, storage: store, maxUpload: maxUploadBytes, storageBackend: opts.StorageBackend, storageBucket: opts.StorageBucket, storageRegion: opts.StorageRegion, keyPrefix: opts.KeyPrefix, signedURLTTL: opts.SignedURLTTL}
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
	features, err := s.repo.ListFeatures(ctx, projectID, models.MoneyFeatureFilter{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	features = filterArchivedDescendants(features)
	counts, err := s.repo.FeatureNoteCounts(ctx, projectID)
	if err != nil {
		return nil, err
	}
	uploads, err := s.repo.PrimaryUploads(ctx, projectID)
	if err != nil {
		return nil, err
	}
	visible := featureIDSet(features)
	return &models.MoneySnapshot{Project: *p, Features: features, NoteCounts: filterNoteCountsByVisibleFeatures(counts, visible), PrimaryUploads: filterPrimaryUploadsByVisibleFeatures(uploads, visible)}, nil
}

func (s *MoneyService) CragSnapshot(ctx context.Context, projectID string) (*models.MoneyCragSnapshot, error) {
	p, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	features, err := s.repo.ListFeatures(ctx, projectID, models.MoneyFeatureFilter{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	features = filterArchivedDescendants(features)
	notes, err := s.repo.ListNotesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	uploads, err := s.repo.ListUploadsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	visible := featureIDSet(features)
	root, trails := BuildMoneyCragTree(features)
	visibleNotes := filterNotesByVisibleFeatures(notes, visible)
	return &models.MoneyCragSnapshot{Project: *p, Root: root, Trails: trails, Notes: visibleNotes, Uploads: filterUploadsByVisibleFeaturesAndNotes(uploads, visibleNotes, visible)}, nil
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
	if filter.Status != "" || filter.IncludeArchived {
		return s.repo.ListFeatures(ctx, projectID, filter)
	}
	filter.IncludeArchived = true
	features, err := s.repo.ListFeatures(ctx, projectID, filter)
	if err != nil {
		return nil, err
	}
	return filterArchivedDescendants(features), nil
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
	current, err := s.repo.GetFeature(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.Status == models.MoneyStatusArchived {
		return nil, ErrMoneyInvalidInput
	}
	f, err := s.featureFromRequest(req)
	if err != nil {
		return nil, err
	}
	if current.FeatureType != f.FeatureType {
		return nil, ErrMoneyInvalidInput
	}
	if req.ParentFeatureID == nil {
		f.ParentFeatureID = current.ParentFeatureID
	}
	if req.ExternalRef == nil {
		f.ExternalRef = current.ExternalRef
	}
	if req.ImportSource == nil {
		f.ImportSource = current.ImportSource
	}
	f.ID = id
	f.UpdatedBy = user.ID
	return s.repo.UpdateFeature(ctx, f)
}

func (s *MoneyService) UpdateAreaGeometry(ctx context.Context, id string, req models.MoneyAreaGeometryRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	current, err := s.repo.GetFeature(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.FeatureType != models.MoneyFeatureArea || current.ParentFeatureID == nil || strings.TrimSpace(*current.ParentFeatureID) == "" {
		return nil, ErrMoneyInvalidInput
	}
	bbox, normalized, err := ValidateAreaPolygonGeoJSON(req.GeoJSON)
	if err != nil {
		return nil, err
	}
	return s.repo.UpdateFeatureGeometry(ctx, id, normalized, *bbox, user.ID)
}

func (s *MoneyService) ArchiveFeature(ctx context.Context, id string, user models.CurrentUser) error {
	return s.ArchiveFeatureWithMode(ctx, id, models.MoneyArchiveModeSubtree, user)
}

func (s *MoneyService) DeleteTrail(ctx context.Context, id string, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	feature, err := s.repo.GetFeature(ctx, id)
	if err != nil {
		return err
	}
	if feature.FeatureType != models.MoneyFeatureTrail || feature.Status == models.MoneyStatusArchived {
		return ErrMoneyInvalidInput
	}
	return s.repo.ArchiveFeature(ctx, id, user.ID)
}

func (s *MoneyService) ArchiveFeatureWithMode(ctx context.Context, id string, mode models.MoneyArchiveMode, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	if mode == "" {
		mode = models.MoneyArchiveModeSubtree
	}
	if mode != models.MoneyArchiveModeSubtree && mode != models.MoneyArchiveModePromoteChildren {
		return ErrMoneyInvalidInput
	}
	features, target, err := s.featuresWithTarget(ctx, id)
	if err != nil {
		return err
	}
	if target.Status == models.MoneyStatusArchived {
		return ErrMoneyInvalidInput
	}
	if target.FeatureType == models.MoneyFeatureTrail {
		return s.repo.ArchiveFeature(ctx, id, user.ID)
	}
	if target.FeatureType != models.MoneyFeatureArea || isMoneyRootFeature(features, target.ID) {
		return ErrMoneyInvalidInput
	}
	if mode == models.MoneyArchiveModePromoteChildren {
		return s.repo.PromoteChildrenAndArchiveFeature(ctx, id, target.ParentFeatureID, user.ID)
	}
	_ = features
	return s.repo.ArchiveFeature(ctx, id, user.ID)
}

func (s *MoneyService) MoveFeatureParent(ctx context.Context, id string, req models.MoneyMoveFeatureRequest, user models.CurrentUser) (*models.MoneyFeature, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	features, target, err := s.featuresWithTarget(ctx, id)
	if err != nil {
		return nil, err
	}
	if target.FeatureType != models.MoneyFeatureArea || isMoneyRootFeature(features, target.ID) || target.Status == models.MoneyStatusArchived {
		return nil, ErrMoneyInvalidInput
	}
	if req.ParentFeatureID != nil {
		parentID := strings.TrimSpace(*req.ParentFeatureID)
		if parentID == "" || parentID == id {
			return nil, ErrMoneyInvalidInput
		}
		req.ParentFeatureID = &parentID
		parent, ok := findFeature(features, parentID)
		if !ok || parent.ProjectID != target.ProjectID || parent.FeatureType != models.MoneyFeatureArea || parent.Status == models.MoneyStatusArchived {
			return nil, ErrMoneyInvalidInput
		}
		if isDescendant(features, parentID, id) || hasArchivedAncestor(features, parentID) {
			return nil, ErrMoneyInvalidInput
		}
	}
	sortOrder := target.SortOrder
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	return s.repo.MoveFeatureParent(ctx, id, req.ParentFeatureID, sortOrder, user.ID)
}

func (s *MoneyService) RestoreFeature(ctx context.Context, id string, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	features, target, err := s.featuresWithTarget(ctx, id)
	if err != nil {
		return err
	}
	if target.FeatureType != models.MoneyFeatureArea || target.ParentFeatureID == nil || strings.TrimSpace(*target.ParentFeatureID) == "" {
		return ErrMoneyInvalidInput
	}
	if hasArchivedAncestor(features, *target.ParentFeatureID) {
		return ErrMoneyInvalidInput
	}
	return s.repo.RestoreFeature(ctx, id, user.ID)
}

func (s *MoneyService) ListTrash(ctx context.Context, projectID string) (*models.MoneyTrashResponse, error) {
	features, err := s.repo.ListFeatures(ctx, projectID, models.MoneyFeatureFilter{IncludeArchived: true})
	if err != nil {
		return nil, err
	}
	return &models.MoneyTrashResponse{Items: buildTrashItems(features)}, nil
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
	blocks := req.Blocks
	if len(blocks) == 0 {
		blocks = json.RawMessage(`[]`)
	}
	if !json.Valid(blocks) || len(blocks) > 128*1024 {
		return nil, ErrMoneyInvalidInput
	}
	return s.repo.UpdateNote(ctx, noteID, body, visibility, cleanTags(req.Tags), blocks, user.ID, user.Role)
}

func (s *MoneyService) DeleteNote(ctx context.Context, noteID string, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	note, err := s.repo.GetNote(ctx, noteID)
	if err != nil {
		return err
	}
	uploads, err := s.uploadsOwnedOnlyByNote(ctx, note)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteNote(ctx, noteID, user.ID, user.Role); err != nil {
		return err
	}
	for _, upload := range uploads {
		if err := s.deleteUploadRecordAndObject(ctx, upload, user); err != nil {
			return err
		}
	}
	return nil
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
	if !allowedUploadContentType(blockKind, contentType) {
		return nil, ErrMoneyInvalidInput
	}
	width, height := imageDimensions(buf)
	uploadID := uuid.NewString()
	safe := safeFilename(fh.Filename, contentType)
	key := s.assetOriginalKey(uploadID, safe)
	stored, err := s.storage.Save(ctx, key, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	bucket := optionalString(s.storageBucket)
	region := optionalString(s.storageRegion)
	etag := optionalString(stored.ETag)
	versionID := optionalString(stored.VersionID)
	return s.repo.CreateUpload(ctx, models.MoneyUpload{ID: uploadID, ProjectID: projectID, FeatureID: featureID, NoteID: noteID, OriginalFilename: safe, StorageKey: stored.StorageKey, ContentType: contentType, ByteSize: stored.ByteSize, Width: width, Height: height, ChecksumSHA256: stored.Checksum, BlockKind: blockKind, Metadata: metadata, AssetKind: "original", StorageBackend: s.storageBackend, StorageBucket: bucket, StorageRegion: region, StorageETag: etag, StorageVersionID: versionID, Visibility: "private", SyncStatus: "available", UploadedBy: user.ID})
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

func (s *MoneyService) SignedUploadDownloadURL(ctx context.Context, uploadID string) (*models.MoneyUploadDownloadURL, error) {
	u, err := s.repo.GetUpload(ctx, uploadID)
	if err != nil {
		return nil, err
	}
	proxyURL := "/api/money/uploads/" + uploadID
	presigner, ok := s.storage.(storage.DownloadPresigner)
	if !ok {
		return &models.MoneyUploadDownloadURL{URL: proxyURL, ExpiresAt: time.Now().Add(s.signedURLTTL), ProxyURL: proxyURL}, nil
	}
	url, err := presigner.SignedGetURL(ctx, u.StorageKey, u.OriginalFilename, u.ContentType, s.signedURLTTL)
	if err != nil {
		return nil, err
	}
	return &models.MoneyUploadDownloadURL{URL: url, ExpiresAt: time.Now().Add(s.signedURLTTL), ProxyURL: proxyURL}, nil
}

func (s *MoneyService) UpdateUploadMetadata(ctx context.Context, uploadID string, req models.MoneyUploadMetadataRequest, user models.CurrentUser) (*models.MoneyUpload, error) {
	if !models.CanWriteMoney(user.Role) {
		return nil, ErrMoneyForbidden
	}
	title := cleanOptional(req.Title, 200)
	comments := cleanOptional(req.Comments, 5000)
	return s.repo.UpdateUploadMetadata(ctx, uploadID, title, comments, user.ID, user.Role)
}

func (s *MoneyService) DeleteUpload(ctx context.Context, uploadID string, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	u, err := s.repo.GetUpload(ctx, uploadID)
	if err != nil {
		return err
	}
	return s.deleteUploadRecordAndObject(ctx, *u, user)
}

func (s *MoneyService) deleteUploadRecordAndObject(ctx context.Context, upload models.MoneyUpload, user models.CurrentUser) error {
	if err := s.repo.DeleteUpload(ctx, upload.ID, user.ID, user.Role); err != nil {
		return err
	}
	if s.storage != nil {
		if err := s.storage.Delete(ctx, upload.StorageKey); err == nil {
			_ = s.repo.MarkUploadPhysicallyDeleted(ctx, upload.ID)
		}
	}
	return nil
}

func (s *MoneyService) uploadsOwnedOnlyByNote(ctx context.Context, note *models.MoneyNote) ([]models.MoneyUpload, error) {
	if note == nil {
		return nil, nil
	}
	noteUploadIDs := uploadIDsFromNoteBlocks(note.Blocks)
	projectUploads, err := s.repo.ListUploadsByProject(ctx, note.ProjectID)
	if err != nil {
		return nil, err
	}
	projectNotes, err := s.repo.ListNotesByProject(ctx, note.ProjectID)
	if err != nil {
		return nil, err
	}
	referencedByOtherNotes := map[string]bool{}
	for _, other := range projectNotes {
		if other.ID == note.ID {
			continue
		}
		for id := range uploadIDsFromNoteBlocks(other.Blocks) {
			referencedByOtherNotes[id] = true
		}
	}
	owned := make([]models.MoneyUpload, 0)
	for _, upload := range projectUploads {
		directlyAttached := upload.NoteID != nil && *upload.NoteID == note.ID
		blockAttached := noteUploadIDs[upload.ID]
		if !directlyAttached && !blockAttached {
			continue
		}
		if referencedByOtherNotes[upload.ID] {
			continue
		}
		owned = append(owned, upload)
	}
	return owned, nil
}

func uploadIDsFromNoteBlocks(blocks json.RawMessage) map[string]bool {
	ids := map[string]bool{}
	if len(blocks) == 0 || !json.Valid(blocks) {
		return ids
	}
	var parsed []models.MoneyNoteBlock
	if err := json.Unmarshal(blocks, &parsed); err != nil {
		return ids
	}
	for _, block := range parsed {
		if block.UploadID != nil && strings.TrimSpace(*block.UploadID) != "" {
			ids[*block.UploadID] = true
		}
	}
	return ids
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
	if req.FeatureType == models.MoneyFeatureTrail {
		var err error
		props, err = normalizeTrailProperties(props)
		if err != nil {
			return models.MoneyFeature{}, err
		}
	}
	bbox, err := ValidateGeoJSON(req.GeoJSON)
	if err != nil {
		return models.MoneyFeature{}, err
	}
	return models.MoneyFeature{ParentFeatureID: req.ParentFeatureID, FeatureType: req.FeatureType, Title: strings.TrimSpace(req.Title), Description: cleanOptional(req.Description, 2000), Status: status, GeoJSON: req.GeoJSON, Style: style, Properties: props, MinLat: &bbox.MinLat, MinLon: &bbox.MinLon, MaxLat: &bbox.MaxLat, MaxLon: &bbox.MaxLon, SortOrder: req.SortOrder, ExternalRef: req.ExternalRef, ImportSource: req.ImportSource}, nil
}

func normalizeTrailProperties(raw json.RawMessage) (json.RawMessage, error) {
	props := map[string]interface{}{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &props); err != nil {
			return nil, ErrMoneyInvalidInput
		}
	}
	category, _ := props["trail_category"].(string)
	category = strings.TrimSpace(category)
	if category == "" {
		category = models.MoneyTrailCategoryConnector
	}
	if !validTrailCategory(category) {
		return nil, ErrMoneyInvalidInput
	}
	props["trail_category"] = category
	if props["trail_destination_feature_id"] != nil {
		id, ok := props["trail_destination_feature_id"].(string)
		if !ok {
			return nil, ErrMoneyInvalidInput
		}
		id = strings.TrimSpace(id)
		if id == "" {
			delete(props, "trail_destination_feature_id")
		} else {
			props["trail_destination_feature_id"] = id
		}
	}
	if props["trail_destination_label"] != nil {
		label, ok := props["trail_destination_label"].(string)
		if !ok {
			return nil, ErrMoneyInvalidInput
		}
		label = strings.TrimSpace(label)
		if len(label) > 200 {
			return nil, ErrMoneyInvalidInput
		}
		if label == "" {
			delete(props, "trail_destination_label")
		} else {
			props["trail_destination_label"] = label
		}
	}
	if (category == models.MoneyTrailCategoryTrailToArea || category == models.MoneyTrailCategoryTrailToDestination) && props["trail_destination_feature_id"] == nil && props["trail_destination_label"] == nil {
		return nil, ErrMoneyInvalidInput
	}
	out, err := json.Marshal(props)
	if err != nil {
		return nil, ErrMoneyInvalidInput
	}
	return out, nil
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

func ValidateAreaPolygonGeoJSON(raw json.RawMessage) (*models.BBox, json.RawMessage, error) {
	if len(raw) == 0 || len(raw) > maxGeoJSONBytes {
		return nil, nil, ErrMoneyInvalidInput
	}
	var obj moneyPolygonGeoJSON
	if json.Unmarshal(raw, &obj) != nil {
		return nil, nil, ErrMoneyInvalidInput
	}
	if obj.Type == "Feature" {
		if obj.Geometry == nil {
			return nil, nil, ErrMoneyInvalidInput
		}
		obj = *obj.Geometry
	}
	if obj.Type != "Polygon" || len(obj.Coordinates) != 1 {
		return nil, nil, ErrMoneyInvalidInput
	}
	ring := obj.Coordinates[0]
	if len(ring) > 0 && samePosition(ring[0], ring[len(ring)-1]) {
		ring = ring[:len(ring)-1]
	}
	if len(ring) < 3 || len(ring) > 500 {
		return nil, nil, ErrMoneyInvalidInput
	}
	seen := map[string]bool{}
	bbox := &models.BBox{MinLat: 1e9, MinLon: 1e9, MaxLat: -1e9, MaxLon: -1e9}
	for _, p := range ring {
		if len(p) != 2 || p[0] < -180 || p[0] > 180 || p[1] < -90 || p[1] > 90 {
			return nil, nil, ErrMoneyInvalidInput
		}
		seen[fmt.Sprintf("%.7f,%.7f", p[0], p[1])] = true
		bbox.MinLon = minFloat(bbox.MinLon, p[0])
		bbox.MaxLon = maxFloat(bbox.MaxLon, p[0])
		bbox.MinLat = minFloat(bbox.MinLat, p[1])
		bbox.MaxLat = maxFloat(bbox.MaxLat, p[1])
	}
	if len(seen) < 3 {
		return nil, nil, ErrMoneyInvalidInput
	}
	closed := append(append([][]float64{}, ring...), ring[0])
	normalized, err := json.Marshal(map[string]interface{}{"type": "Polygon", "coordinates": [][][]float64{closed}})
	if err != nil {
		return nil, nil, ErrMoneyInvalidInput
	}
	return bbox, normalized, nil
}

type moneyPolygonGeoJSON struct {
	Type        string               `json:"type"`
	Coordinates [][][]float64        `json:"coordinates,omitempty"`
	Geometry    *moneyPolygonGeoJSON `json:"geometry,omitempty"`
}

func samePosition(a, b []float64) bool {
	return len(a) == 2 && len(b) == 2 && a[0] == b[0] && a[1] == b[1]
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

func (s *MoneyService) featuresWithTarget(ctx context.Context, id string) ([]models.MoneyFeature, *models.MoneyFeature, error) {
	target, err := s.repo.GetFeature(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	features, err := s.repo.ListFeatures(ctx, target.ProjectID, models.MoneyFeatureFilter{IncludeArchived: true})
	if err != nil {
		return nil, nil, err
	}
	return features, target, nil
}

func subtreeIDs(features []models.MoneyFeature, rootID string) []string {
	children := map[string][]string{}
	for _, f := range features {
		if f.ParentFeatureID != nil {
			children[*f.ParentFeatureID] = append(children[*f.ParentFeatureID], f.ID)
		}
	}
	ids := []string{}
	var walk func(string)
	walk = func(id string) {
		ids = append(ids, id)
		for _, childID := range children[id] {
			walk(childID)
		}
	}
	walk(rootID)
	return ids
}

func isMoneyRootFeature(features []models.MoneyFeature, id string) bool {
	root, _ := BuildMoneyCragTree(features)
	return root != nil && root.Feature.ID == id
}

func findFeature(features []models.MoneyFeature, id string) (models.MoneyFeature, bool) {
	for _, f := range features {
		if f.ID == id {
			return f, true
		}
	}
	return models.MoneyFeature{}, false
}

func isDescendant(features []models.MoneyFeature, id, ancestorID string) bool {
	byID := make(map[string]models.MoneyFeature, len(features))
	for _, f := range features {
		byID[f.ID] = f
	}
	for id != "" {
		f, ok := byID[id]
		if !ok || f.ParentFeatureID == nil {
			return false
		}
		if *f.ParentFeatureID == ancestorID {
			return true
		}
		id = *f.ParentFeatureID
	}
	return false
}

func hasArchivedAncestor(features []models.MoneyFeature, parentID string) bool {
	byID := make(map[string]models.MoneyFeature, len(features))
	for _, f := range features {
		byID[f.ID] = f
	}
	for parentID != "" {
		parent, ok := byID[parentID]
		if !ok {
			return false
		}
		if parent.Status == models.MoneyStatusArchived {
			return true
		}
		if parent.ParentFeatureID == nil {
			return false
		}
		parentID = *parent.ParentFeatureID
	}
	return false
}

func buildTrashItems(features []models.MoneyFeature) []models.MoneyTrashItem {
	byID := make(map[string]models.MoneyFeature, len(features))
	archived := map[string]bool{}
	for _, f := range features {
		byID[f.ID] = f
		if f.Status == models.MoneyStatusArchived {
			archived[f.ID] = true
		}
	}
	items := []models.MoneyTrashItem{}
	for _, f := range features {
		if !archived[f.ID] || f.FeatureType != models.MoneyFeatureArea {
			continue
		}
		if f.ParentFeatureID != nil && archived[*f.ParentFeatureID] {
			continue
		}
		ids := subtreeIDs(features, f.ID)
		desc := len(ids) - 1
		items = append(items, models.MoneyTrashItem{ID: f.ID, Title: f.Title, FeatureType: f.FeatureType, ParentFeatureID: f.ParentFeatureID, Path: featurePath(byID, f.ID), DeletedAt: f.UpdatedAt, UpdatedAt: f.UpdatedAt, DescendantCount: desc})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].DeletedAt.After(items[j].DeletedAt) })
	return items
}

func featurePath(byID map[string]models.MoneyFeature, id string) []string {
	var reversed []string
	for id != "" {
		f, ok := byID[id]
		if !ok {
			break
		}
		reversed = append(reversed, f.Title)
		if f.ParentFeatureID == nil {
			break
		}
		id = *f.ParentFeatureID
	}
	path := make([]string, len(reversed))
	for i := range reversed {
		path[i] = reversed[len(reversed)-1-i]
	}
	return path
}

func filterArchivedDescendants(features []models.MoneyFeature) []models.MoneyFeature {
	visible := make([]models.MoneyFeature, 0, len(features))
	for _, f := range features {
		if f.Status == models.MoneyStatusArchived {
			continue
		}
		if f.ParentFeatureID != nil && hasArchivedAncestor(features, *f.ParentFeatureID) {
			continue
		}
		visible = append(visible, f)
	}
	return visible
}

func featureIDSet(features []models.MoneyFeature) map[string]bool {
	out := make(map[string]bool, len(features))
	for _, f := range features {
		out[f.ID] = true
	}
	return out
}

func filterNotesByVisibleFeatures(notes []models.MoneyNote, visible map[string]bool) []models.MoneyNote {
	out := make([]models.MoneyNote, 0, len(notes))
	for _, n := range notes {
		if n.FeatureID != nil && !visible[*n.FeatureID] {
			continue
		}
		if n.TargetRef != nil && isFeatureTarget(n.TargetType) && !visible[*n.TargetRef] {
			continue
		}
		out = append(out, n)
	}
	return out
}

func filterUploadsByVisibleFeatures(uploads []models.MoneyUpload, visible map[string]bool) []models.MoneyUpload {
	return filterUploadsByVisibleFeaturesAndNotes(uploads, nil, visible)
}

func filterUploadsByVisibleFeaturesAndNotes(uploads []models.MoneyUpload, notes []models.MoneyNote, visible map[string]bool) []models.MoneyUpload {
	visibleNotes := make(map[string]bool, len(notes))
	visibleNoteUploadIDs := map[string]bool{}
	for _, note := range notes {
		visibleNotes[note.ID] = true
		for id := range uploadIDsFromNoteBlocks(note.Blocks) {
			visibleNoteUploadIDs[id] = true
		}
	}
	out := make([]models.MoneyUpload, 0, len(uploads))
	for _, u := range uploads {
		if u.FeatureID != nil && !visible[*u.FeatureID] {
			continue
		}
		if u.NoteID != nil && !visibleNotes[*u.NoteID] && !visibleNoteUploadIDs[u.ID] {
			continue
		}
		out = append(out, u)
	}
	return out
}

func filterNoteCountsByVisibleFeatures(counts map[string]int, visible map[string]bool) map[string]int {
	out := make(map[string]int, len(counts))
	for id, count := range counts {
		if visible[id] {
			out[id] = count
		}
	}
	return out
}

func filterPrimaryUploadsByVisibleFeatures(uploads map[string]models.MoneyUpload, visible map[string]bool) map[string]models.MoneyUpload {
	out := make(map[string]models.MoneyUpload, len(uploads))
	for id, upload := range uploads {
		if visible[id] {
			out[id] = upload
		}
	}
	return out
}

func isFeatureTarget(targetType string) bool {
	return targetType == "feature" || targetType == "area" || targetType == "boulder" || targetType == "trail" || targetType == "problem" || targetType == "point"
}

func permissions(role string) models.MoneyPermissions {
	return models.MoneyPermissions{CanRead: true, CanWrite: models.CanWriteMoney(role), IsAdmin: role == models.RoleAdmin}
}
func validFeatureType(t string) bool {
	return t == models.MoneyFeatureArea || t == models.MoneyFeatureBoulder || t == models.MoneyFeatureProblem || t == models.MoneyFeatureTrail || t == models.MoneyFeatureTopo || t == models.MoneyFeaturePOI || t == models.MoneyFeatureDrawing
}
func validTrailCategory(category string) bool {
	return category == models.MoneyTrailCategoryConnector || category == models.MoneyTrailCategoryApproach || category == models.MoneyTrailCategoryTrailToArea || category == models.MoneyTrailCategoryTrailToDestination
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
func allowedUploadContentType(blockKind, ct string) bool {
	if blockKind != "file" {
		return allowedImage(ct)
	}
	return allowedImage(ct) || ct == "application/pdf" || ct == "text/plain"
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
func (s *MoneyService) assetOriginalKey(uploadID, filename string) string {
	return filepath.ToSlash(filepath.Join(s.keyPrefix, "assets", uploadID, "original", filename))
}

func normalizeAssetKeyPrefix(prefix string) string {
	prefix = filepath.ToSlash(strings.TrimSpace(prefix))
	parts := strings.Split(prefix, "/")
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "." || part == ".." {
			continue
		}
		clean = append(clean, part)
	}
	if len(clean) == 0 {
		return "money-creek"
	}
	return strings.Join(clean, "/")
}

func optionalString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
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
