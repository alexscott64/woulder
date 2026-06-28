package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/alexscott64/woulder/backend/internal/storage"
)

func TestValidateGeoJSONDerivesBBox(t *testing.T) {
	raw := json.RawMessage(`{"type":"LineString","coordinates":[[-121.52,47.71],[-121.50,47.73]]}`)
	bbox, err := ValidateGeoJSON(raw)
	if err != nil {
		t.Fatalf("ValidateGeoJSON returned error: %v", err)
	}
	if bbox.MinLon != -121.52 || bbox.MaxLon != -121.50 || bbox.MinLat != 47.71 || bbox.MaxLat != 47.73 {
		t.Fatalf("unexpected bbox: %+v", bbox)
	}
}

func TestValidateAreaPolygonGeoJSONClosesRingAndDerivesBBox(t *testing.T) {
	raw := json.RawMessage(`{"type":"Polygon","coordinates":[[[-121.52,47.71],[-121.50,47.71],[-121.51,47.73]]]}`)
	bbox, normalized, err := ValidateAreaPolygonGeoJSON(raw)
	if err != nil {
		t.Fatalf("ValidateAreaPolygonGeoJSON returned error: %v", err)
	}
	if bbox.MinLon != -121.52 || bbox.MaxLon != -121.50 || bbox.MinLat != 47.71 || bbox.MaxLat != 47.73 {
		t.Fatalf("unexpected bbox: %+v", bbox)
	}
	var polygon struct {
		Coordinates [][][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(normalized, &polygon); err != nil {
		t.Fatalf("normalized polygon is invalid JSON: %v", err)
	}
	ring := polygon.Coordinates[0]
	first, last := ring[0], ring[len(ring)-1]
	if first[0] != last[0] || first[1] != last[1] {
		t.Fatalf("expected closed ring, got %v", ring)
	}
}

func TestValidateAreaPolygonGeoJSONRejectsWorldCoordinates(t *testing.T) {
	raw := json.RawMessage(`{"type":"Polygon","coordinates":[[[0,0],[100,0],[100,100],[0,0]]]}`)
	if _, _, err := ValidateAreaPolygonGeoJSON(raw); err == nil {
		t.Fatal("expected WGS84 polygon validation error")
	}
}

func TestFilterArchivedDescendantsHidesSubtree(t *testing.T) {
	root := moneyTestFeature("root", "", models.MoneyFeatureArea, models.MoneyStatusActive)
	area := moneyTestFeature("area", "root", models.MoneyFeatureArea, models.MoneyStatusArchived)
	boulder := moneyTestFeature("boulder", "area", models.MoneyFeatureBoulder, models.MoneyStatusScouted)
	problem := moneyTestFeature("problem", "boulder", models.MoneyFeatureProblem, models.MoneyStatusProject)
	sibling := moneyTestFeature("sibling", "root", models.MoneyFeatureArea, models.MoneyStatusActive)

	visible := filterArchivedDescendants([]models.MoneyFeature{root, area, boulder, problem, sibling})
	ids := make([]string, 0, len(visible))
	for _, f := range visible {
		ids = append(ids, f.ID)
	}
	if got, want := ids, []string{"root", "sibling"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("unexpected visible ids: got %v want %v", got, want)
	}
}

func TestBuildTrashItemsReturnsTopLevelDeletedAreasWithDescendants(t *testing.T) {
	root := moneyTestFeature("root", "", models.MoneyFeatureArea, models.MoneyStatusActive)
	area := moneyTestFeature("area", "root", models.MoneyFeatureArea, models.MoneyStatusArchived)
	child := moneyTestFeature("child", "area", models.MoneyFeatureArea, models.MoneyStatusActive)
	boulder := moneyTestFeature("boulder", "child", models.MoneyFeatureBoulder, models.MoneyStatusScouted)
	nestedDeleted := moneyTestFeature("nested", "area", models.MoneyFeatureArea, models.MoneyStatusArchived)

	items := buildTrashItems([]models.MoneyFeature{root, area, child, boulder, nestedDeleted})
	if len(items) != 1 {
		t.Fatalf("expected one top-level trash item, got %d", len(items))
	}
	if items[0].ID != "area" || items[0].DescendantCount != 3 {
		t.Fatalf("unexpected trash item: %+v", items[0])
	}
	if got := items[0].Path; len(got) != 2 || got[0] != "Root" || got[1] != "Area" {
		t.Fatalf("unexpected path: %v", got)
	}
}

func moneyTestFeature(id, parentID, featureType, status string) models.MoneyFeature {
	var parent *string
	if parentID != "" {
		parent = &parentID
	}
	return models.MoneyFeature{ID: id, ParentFeatureID: parent, FeatureType: featureType, Title: strings.Title(id), Status: status, UpdatedAt: time.Unix(int64(len(id)), 0)}
}

func TestValidateAreaPolygonGeoJSONRejectsTooFewDistinctVertices(t *testing.T) {
	raw := json.RawMessage(`{"type":"Polygon","coordinates":[[[-121.52,47.71],[-121.52,47.71],[-121.51,47.73],[-121.52,47.71]]]}`)
	if _, _, err := ValidateAreaPolygonGeoJSON(raw); err == nil {
		t.Fatal("expected distinct vertex validation error")
	}
}

func TestValidateGeoJSONRejectsUnsupportedGeometry(t *testing.T) {
	raw := json.RawMessage(`{"type":"MultiPoint","coordinates":[[-121.52,47.71]]}`)
	if _, err := ValidateGeoJSON(raw); err == nil {
		t.Fatal("expected unsupported geometry error")
	}
}

func TestValidateGeoJSONRejectsOutOfRangeCoordinates(t *testing.T) {
	raw := json.RawMessage(`{"type":"Point","coordinates":[-181,47.71]}`)
	if _, err := ValidateGeoJSON(raw); err == nil {
		t.Fatal("expected coordinate range error")
	}
}

func TestMoneyServiceAssetOriginalKeyUsesMoneyCreekPrefix(t *testing.T) {
	svc := NewMoneyServiceWithOptions(nil, nil, 0, MoneyServiceOptions{KeyPrefix: "money-creek"})
	got := svc.assetOriginalKey("asset-id", "photo.jpg")
	want := "money-creek/assets/asset-id/original/photo.jpg"
	if got != want {
		t.Fatalf("unexpected asset key: got %q want %q", got, want)
	}
}

func TestMoneyServiceAssetKeyPrefixIsNormalized(t *testing.T) {
	svc := NewMoneyServiceWithOptions(nil, nil, 0, MoneyServiceOptions{KeyPrefix: "/money-creek/../tmp//"})
	got := svc.assetOriginalKey("asset-id", "photo.jpg")
	want := "money-creek/tmp/assets/asset-id/original/photo.jpg"
	if got != want {
		t.Fatalf("unexpected normalized asset key: got %q want %q", got, want)
	}
}

func TestStoreUploadStoresChecksumAndPrivateMetadata(t *testing.T) {
	repo := &moneyUploadRepo{}
	store := &captureStorage{}
	svc := NewMoneyServiceWithOptions(repo, store, 1024, MoneyServiceOptions{StorageBackend: "r2", StorageBucket: "woulder", StorageRegion: "auto", KeyPrefix: "money-creek"})
	jpeg := []byte{0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43, 0x00, 0xff, 0xd9}
	fh := multipartFileHeader(t, jpeg, "photo.jpg")

	upload, err := svc.StoreUploadWithKind(context.Background(), "project-1", moneyStrPtr("feature-1"), nil, "photo", json.RawMessage(`{"camera":"phone"}`), fh, models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper})
	if err != nil {
		t.Fatalf("StoreUploadWithKind returned error: %v", err)
	}
	if upload.StorageBackend != "r2" || upload.StorageBucket == nil || *upload.StorageBucket != "woulder" || upload.Visibility != "private" || upload.SyncStatus != "available" {
		t.Fatalf("unexpected storage metadata: %+v", upload)
	}
	if upload.ChecksumSHA256 != "b94d27b9934d3e08a52e52d7da7dabfadeb4c484efe37a5380ee9088f7ace2ef" {
		t.Fatalf("unexpected checksum: %s", upload.ChecksumSHA256)
	}
	if upload.ByteSize != int64(len(jpeg)) || store.savedKey != upload.StorageKey {
		t.Fatalf("unexpected stored file metadata: upload=%+v key=%q", upload, store.savedKey)
	}
}

func TestSignedUploadDownloadURLUsesPresignerWhenAvailable(t *testing.T) {
	repo := &moneyUploadRepo{upload: &models.MoneyUpload{ID: "upload-1", StorageKey: "money-creek/assets/upload-1/original/photo.jpg", OriginalFilename: "photo.jpg", ContentType: "image/jpeg"}}
	store := &captureStorage{signedURL: "https://example.invalid/signed"}
	svc := NewMoneyServiceWithOptions(repo, store, 1024, MoneyServiceOptions{SignedURLTTL: time.Minute})

	resp, err := svc.SignedUploadDownloadURL(context.Background(), "upload-1")
	if err != nil {
		t.Fatalf("SignedUploadDownloadURL returned error: %v", err)
	}
	if resp.URL != store.signedURL || resp.ProxyURL != "/api/money/uploads/upload-1" {
		t.Fatalf("unexpected download response: %+v", resp)
	}
	if store.presignedKey != repo.upload.StorageKey || store.presignedFilename != "photo.jpg" || store.presignedContentType != "image/jpeg" {
		t.Fatalf("unexpected presign inputs: %+v", store)
	}
}

func TestUpdateUploadMetadataTrimsAndSavesTitleComments(t *testing.T) {
	repo := &moneyUploadRepo{upload: &models.MoneyUpload{ID: "upload-1", OriginalFilename: "photo.jpg", UploadedBy: "user-1"}}
	svc := NewMoneyService(repo, nil, 1024)
	title := "  Topo overview  "
	comments := "  Shows the main face.  "

	upload, err := svc.UpdateUploadMetadata(context.Background(), "upload-1", models.MoneyUploadMetadataRequest{Title: &title, Comments: &comments}, models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper})
	if err != nil {
		t.Fatalf("UpdateUploadMetadata returned error: %v", err)
	}
	if upload.Title == nil || *upload.Title != "Topo overview" || upload.Comments == nil || *upload.Comments != "Shows the main face." {
		t.Fatalf("unexpected upload metadata: %+v", upload)
	}
}

func TestDeleteUploadSoftDeletesThenMarksPhysicalDelete(t *testing.T) {
	repo := &moneyUploadRepo{upload: &models.MoneyUpload{ID: "upload-1", StorageKey: "money-creek/assets/upload-1/original/photo.jpg", UploadedBy: "user-1"}}
	store := &captureStorage{}
	svc := NewMoneyService(repo, store, 1024)

	if err := svc.DeleteUpload(context.Background(), "upload-1", models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper}); err != nil {
		t.Fatalf("DeleteUpload returned error: %v", err)
	}
	if !repo.deleted || !repo.markedPhysical || store.deletedKey != repo.upload.StorageKey {
		t.Fatalf("expected soft and physical delete markers, repo=%+v store=%+v", repo, store)
	}
}

func TestDeleteNoteDeletesOnlyUploadsOwnedByThatNote(t *testing.T) {
	noteID := "note-1"
	sharedUploadID := "shared-upload"
	noteOnlyUploadID := "note-only-upload"
	directUploadID := "direct-upload"
	repo := &moneyUploadRepo{
		note:  &models.MoneyNote{ID: noteID, ProjectID: "project-1", Blocks: json.RawMessage(`[{"kind":"photo","upload_id":"shared-upload"},{"kind":"photo","upload_id":"note-only-upload"}]`)},
		notes: []models.MoneyNote{{ID: "note-2", ProjectID: "project-1", Blocks: json.RawMessage(`[{"kind":"photo","upload_id":"shared-upload"}]`)}},
		uploads: []models.MoneyUpload{
			{ID: sharedUploadID, ProjectID: "project-1", StorageKey: "money-creek/shared.jpg", UploadedBy: "user-1"},
			{ID: noteOnlyUploadID, ProjectID: "project-1", StorageKey: "money-creek/note-only.jpg", UploadedBy: "user-1"},
			{ID: directUploadID, ProjectID: "project-1", NoteID: &noteID, StorageKey: "money-creek/direct.jpg", UploadedBy: "user-1"},
		},
	}
	store := &captureStorage{}
	svc := NewMoneyService(repo, store, 1024)

	if err := svc.DeleteNote(context.Background(), noteID, models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper}); err != nil {
		t.Fatalf("DeleteNote returned error: %v", err)
	}
	if !repo.noteDeleted {
		t.Fatal("expected note to be soft deleted")
	}
	if got := strings.Join(repo.deletedUploadIDs, ","); got != "note-only-upload,direct-upload" {
		t.Fatalf("expected only note-owned uploads to be deleted, got %q", got)
	}
	if got := strings.Join(store.deletedKeys, ","); got != "money-creek/note-only.jpg,money-creek/direct.jpg" {
		t.Fatalf("expected note-owned storage objects to be deleted, got %q", got)
	}
}

func TestAllowedUploadContentType(t *testing.T) {
	if !allowedUploadContentType("photo", "image/jpeg") {
		t.Fatal("expected jpeg photo to be allowed")
	}
	if allowedUploadContentType("photo", "application/pdf") {
		t.Fatal("expected pdf photo to be rejected")
	}
	if !allowedUploadContentType("file", "application/pdf") {
		t.Fatal("expected pdf file to be allowed")
	}
}

func TestUpdateTrailLabelAndDestinationMetadata(t *testing.T) {
	repo := &moneyUploadRepo{feature: &models.MoneyFeature{ID: "trail-1", FeatureType: models.MoneyFeatureTrail, Status: models.MoneyStatusActive}}
	svc := NewMoneyService(repo, nil, 0)

	updated, err := svc.UpdateFeature(context.Background(), "trail-1", models.MoneyFeatureRequest{
		FeatureType: models.MoneyFeatureTrail,
		Title:       "New approach label",
		Status:      models.MoneyStatusActive,
		GeoJSON:     json.RawMessage(`{"type":"LineString","coordinates":[[-121.52,47.71],[-121.51,47.72]]}`),
		Properties:  json.RawMessage(`{"trail_category":"trail_to_area","trail_destination_feature_id":" area-7 ","trail_destination_label":""}`),
	}, models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper})
	if err != nil {
		t.Fatalf("UpdateFeature returned error: %v", err)
	}
	if updated.Title != "New approach label" || updated.UpdatedBy != "user-1" {
		t.Fatalf("unexpected updated trail: %+v", updated)
	}
	var props map[string]string
	if err := json.Unmarshal(updated.Properties, &props); err != nil {
		t.Fatalf("updated properties are invalid JSON: %v", err)
	}
	if props["trail_category"] != models.MoneyTrailCategoryTrailToArea || props["trail_destination_feature_id"] != "area-7" {
		t.Fatalf("unexpected trail properties: %v", props)
	}
	if _, ok := props["trail_destination_label"]; ok {
		t.Fatalf("expected empty destination label to be omitted: %v", props)
	}
}

func TestUpdateBoulderRenamePreservesParentGeometryAndMetadata(t *testing.T) {
	parentID := "area-7"
	externalRef := "ref-1"
	importSource := "money_reference"
	repo := &moneyUploadRepo{feature: &models.MoneyFeature{
		ID:              "boulder-1",
		ProjectID:       "project-1",
		ParentFeatureID: &parentID,
		FeatureType:     models.MoneyFeatureBoulder,
		Title:           "Old name",
		Status:          models.MoneyStatusScouted,
		ExternalRef:     &externalRef,
		ImportSource:    &importSource,
	}}
	svc := NewMoneyService(repo, nil, 0)

	updated, err := svc.UpdateFeature(context.Background(), "boulder-1", models.MoneyFeatureRequest{
		FeatureType: models.MoneyFeatureBoulder,
		Title:       "New name",
		Status:      models.MoneyStatusScouted,
		GeoJSON:     json.RawMessage(`{"type":"Polygon","coordinates":[[[-121.52,47.71],[-121.51,47.71],[-121.51,47.72],[-121.52,47.71]]]}`),
		Properties:  json.RawMessage(`{"landing":"flat"}`),
		Style:       json.RawMessage(`{"color":"green"}`),
		SortOrder:   4,
	}, models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper})
	if err != nil {
		t.Fatalf("UpdateFeature returned error: %v", err)
	}
	if updated.Title != "New name" || updated.ParentFeatureID == nil || *updated.ParentFeatureID != parentID || updated.Status != models.MoneyStatusScouted {
		t.Fatalf("unexpected updated boulder: %+v", updated)
	}
	if updated.ExternalRef == nil || *updated.ExternalRef != externalRef || updated.ImportSource == nil || *updated.ImportSource != importSource {
		t.Fatalf("expected import metadata preserved, got external=%v source=%v", updated.ExternalRef, updated.ImportSource)
	}
	if updated.MinLat == nil || *updated.MinLat != 47.71 || updated.MaxLon == nil || *updated.MaxLon != -121.51 {
		t.Fatalf("expected bbox from preserved geometry, got %+v", updated)
	}
	var props map[string]string
	if err := json.Unmarshal(updated.Properties, &props); err != nil || props["landing"] != "flat" {
		t.Fatalf("unexpected properties: %v err=%v", props, err)
	}
}

func TestUpdateFeatureRejectsTypeChange(t *testing.T) {
	repo := &moneyUploadRepo{feature: &models.MoneyFeature{ID: "boulder-1", FeatureType: models.MoneyFeatureBoulder, Status: models.MoneyStatusScouted}}
	svc := NewMoneyService(repo, nil, 0)

	_, err := svc.UpdateFeature(context.Background(), "boulder-1", models.MoneyFeatureRequest{
		FeatureType: models.MoneyFeatureTrail,
		Title:       "Not a trail",
		Status:      models.MoneyStatusActive,
		GeoJSON:     json.RawMessage(`{"type":"LineString","coordinates":[[-121.52,47.71],[-121.51,47.72]]}`),
	}, models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper})
	if !errors.Is(err, ErrMoneyInvalidInput) {
		t.Fatalf("expected ErrMoneyInvalidInput, got %v", err)
	}
}

func TestUpdateTrailDestinationRequiresTarget(t *testing.T) {
	repo := &moneyUploadRepo{feature: &models.MoneyFeature{ID: "trail-1", FeatureType: models.MoneyFeatureTrail, Status: models.MoneyStatusActive}}
	svc := NewMoneyService(repo, nil, 0)

	_, err := svc.UpdateFeature(context.Background(), "trail-1", models.MoneyFeatureRequest{
		FeatureType: models.MoneyFeatureTrail,
		Title:       "Trail",
		Status:      models.MoneyStatusActive,
		GeoJSON:     json.RawMessage(`{"type":"LineString","coordinates":[[-121.52,47.71],[-121.51,47.72]]}`),
		Properties:  json.RawMessage(`{"trail_category":"trail_to_area"}`),
	}, models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper})
	if !errors.Is(err, ErrMoneyInvalidInput) {
		t.Fatalf("expected ErrMoneyInvalidInput, got %v", err)
	}
}

func TestDeleteTrailArchivesTrail(t *testing.T) {
	repo := &moneyUploadRepo{feature: &models.MoneyFeature{ID: "trail-1", FeatureType: models.MoneyFeatureTrail, Status: models.MoneyStatusActive}}
	svc := NewMoneyService(repo, nil, 0)

	if err := svc.DeleteTrail(context.Background(), "trail-1", models.CurrentUser{ID: "user-1", Role: models.RoleDeveloper}); err != nil {
		t.Fatalf("DeleteTrail returned error: %v", err)
	}
	if repo.archivedID != "trail-1" || repo.archivedBy != "user-1" {
		t.Fatalf("expected trail archive, got id=%q by=%q", repo.archivedID, repo.archivedBy)
	}
}

func moneyStrPtr(v string) *string { return &v }

type moneyUploadRepo struct {
	feature          *models.MoneyFeature
	note             *models.MoneyNote
	notes            []models.MoneyNote
	upload           *models.MoneyUpload
	uploads          []models.MoneyUpload
	created          *models.MoneyUpload
	deleted          bool
	deletedUploadIDs []string
	markedPhysical   bool
	archivedID       string
	archivedBy       string
	noteDeleted      bool
}

func (r *moneyUploadRepo) GetProjectBySlug(ctx context.Context, slug string) (*models.MoneyProject, error) {
	return nil, nil
}
func (r *moneyUploadRepo) GetProjectByID(ctx context.Context, id string) (*models.MoneyProject, error) {
	return nil, nil
}
func (r *moneyUploadRepo) ListFeatures(ctx context.Context, projectID string, filter models.MoneyFeatureFilter) ([]models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyUploadRepo) GetFeature(ctx context.Context, id string) (*models.MoneyFeature, error) {
	return r.feature, nil
}
func (r *moneyUploadRepo) CreateFeature(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyUploadRepo) UpdateFeature(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error) {
	r.feature = &feature
	return &feature, nil
}
func (r *moneyUploadRepo) UpdateFeatureGeometry(ctx context.Context, id string, geojson []byte, bbox models.BBox, updatedBy string) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyUploadRepo) UpsertFeatureByExternalRef(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyUploadRepo) ArchiveFeature(ctx context.Context, id, userID string) error {
	r.archivedID = id
	r.archivedBy = userID
	return nil
}
func (r *moneyUploadRepo) PromoteChildrenAndArchiveFeature(ctx context.Context, id string, parentID *string, userID string) error {
	return nil
}
func (r *moneyUploadRepo) RestoreFeature(ctx context.Context, id, userID string) error { return nil }
func (r *moneyUploadRepo) MoveFeatureParent(ctx context.Context, id string, parentID *string, sortOrder int, userID string) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyUploadRepo) ListTrash(ctx context.Context, projectID string) ([]models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyUploadRepo) ListNotes(ctx context.Context, featureID string) ([]models.MoneyNote, error) {
	return nil, nil
}
func (r *moneyUploadRepo) ListNotesByProject(ctx context.Context, projectID string) ([]models.MoneyNote, error) {
	return r.notes, nil
}
func (r *moneyUploadRepo) GetNote(ctx context.Context, noteID string) (*models.MoneyNote, error) {
	return r.note, nil
}
func (r *moneyUploadRepo) CreateNote(ctx context.Context, note models.MoneyNote) (*models.MoneyNote, error) {
	return nil, nil
}
func (r *moneyUploadRepo) UpdateNote(ctx context.Context, noteID, body, visibility string, tags []string, blocks []byte, userID, role string) (*models.MoneyNote, error) {
	return &models.MoneyNote{ID: noteID, Body: body, Visibility: visibility, Tags: tags, Blocks: blocks, UpdatedBy: userID}, nil
}
func (r *moneyUploadRepo) DeleteNote(ctx context.Context, noteID, userID, role string) error {
	r.noteDeleted = true
	return nil
}
func (r *moneyUploadRepo) CreateUpload(ctx context.Context, upload models.MoneyUpload) (*models.MoneyUpload, error) {
	r.created = &upload
	return &upload, nil
}
func (r *moneyUploadRepo) GetUpload(ctx context.Context, id string) (*models.MoneyUpload, error) {
	return r.upload, nil
}
func (r *moneyUploadRepo) ListUploadsByFeature(ctx context.Context, featureID string) ([]models.MoneyUpload, error) {
	return nil, nil
}
func (r *moneyUploadRepo) ListUploadsByProject(ctx context.Context, projectID string) ([]models.MoneyUpload, error) {
	return r.uploads, nil
}
func (r *moneyUploadRepo) DeleteUpload(ctx context.Context, uploadID, userID, role string) error {
	r.deleted = true
	r.deletedUploadIDs = append(r.deletedUploadIDs, uploadID)
	return nil
}
func (r *moneyUploadRepo) MarkUploadPhysicallyDeleted(ctx context.Context, uploadID string) error {
	r.markedPhysical = true
	return nil
}
func (r *moneyUploadRepo) UpdateUploadMetadata(ctx context.Context, uploadID string, title, comments *string, userID, role string) (*models.MoneyUpload, error) {
	if r.upload == nil {
		r.upload = &models.MoneyUpload{ID: uploadID}
	}
	r.upload.Title = title
	r.upload.Comments = comments
	return r.upload, nil
}
func (r *moneyUploadRepo) FeatureNoteCounts(ctx context.Context, projectID string) (map[string]int, error) {
	return nil, nil
}
func (r *moneyUploadRepo) PrimaryUploads(ctx context.Context, projectID string) (map[string]models.MoneyUpload, error) {
	return nil, nil
}

type captureStorage struct {
	savedKey             string
	deletedKey           string
	deletedKeys          []string
	signedURL            string
	presignedKey         string
	presignedFilename    string
	presignedContentType string
	presignedTTL         time.Duration
}

func (s *captureStorage) Save(ctx context.Context, key string, r io.Reader) (storage.StoredFile, error) {
	s.savedKey = key
	b, _ := io.ReadAll(r)
	return storage.StoredFile{StorageKey: key, ByteSize: int64(len(b)), Checksum: "b94d27b9934d3e08a52e52d7da7dabfadeb4c484efe37a5380ee9088f7ace2ef", ETag: "etag", VersionID: "version"}, nil
}
func (s *captureStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (s *captureStorage) Delete(ctx context.Context, key string) error {
	s.deletedKey = key
	s.deletedKeys = append(s.deletedKeys, key)
	return nil
}
func (s *captureStorage) SignedGetURL(ctx context.Context, key, filename, contentType string, ttl time.Duration) (string, error) {
	s.presignedKey = key
	s.presignedFilename = filename
	s.presignedContentType = contentType
	s.presignedTTL = ttl
	return s.signedURL, nil
}

func multipartFileHeader(t *testing.T, content []byte, filename string) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "image/jpeg")
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("CreatePart returned error: %v", err)
	}
	_, _ = part.Write(content)
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart writer close returned error: %v", err)
	}
	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(int64(body.Len()))
	if err != nil {
		t.Fatalf("ReadForm returned error: %v", err)
	}
	files := form.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected one file, got %d", len(files))
	}
	return files[0]
}
