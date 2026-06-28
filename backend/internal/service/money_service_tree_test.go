package service

import (
	"context"
	"testing"

	"github.com/alexscott64/woulder/backend/internal/models"
)

type moneyTreeRepo struct {
	features []models.MoneyFeature
}

func (r *moneyTreeRepo) GetProjectBySlug(ctx context.Context, slug string) (*models.MoneyProject, error) {
	return nil, nil
}
func (r *moneyTreeRepo) GetProjectByID(ctx context.Context, id string) (*models.MoneyProject, error) {
	return nil, nil
}
func (r *moneyTreeRepo) ListFeatures(ctx context.Context, projectID string, filter models.MoneyFeatureFilter) ([]models.MoneyFeature, error) {
	out := make([]models.MoneyFeature, len(r.features))
	copy(out, r.features)
	return out, nil
}
func (r *moneyTreeRepo) GetFeature(ctx context.Context, id string) (*models.MoneyFeature, error) {
	for i := range r.features {
		if r.features[i].ID == id {
			return &r.features[i], nil
		}
	}
	return nil, ErrMoneyInvalidInput
}
func (r *moneyTreeRepo) CreateFeature(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyTreeRepo) UpdateFeature(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyTreeRepo) UpdateFeatureGeometry(ctx context.Context, id string, geojson []byte, bbox models.BBox, updatedBy string) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyTreeRepo) UpsertFeatureByExternalRef(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyTreeRepo) ArchiveFeature(ctx context.Context, id, userID string) error {
	for i := range r.features {
		if r.features[i].ID == id {
			r.features[i].Status = models.MoneyStatusArchived
		}
	}
	return nil
}
func (r *moneyTreeRepo) PromoteChildrenAndArchiveFeature(ctx context.Context, id string, parentID *string, userID string) error {
	for i := range r.features {
		if r.features[i].ParentFeatureID != nil && *r.features[i].ParentFeatureID == id && r.features[i].Status != models.MoneyStatusArchived {
			r.features[i].ParentFeatureID = parentID
		}
	}
	return r.ArchiveFeature(ctx, id, userID)
}
func (r *moneyTreeRepo) RestoreFeature(ctx context.Context, id, userID string) error { return nil }
func (r *moneyTreeRepo) MoveFeatureParent(ctx context.Context, id string, parentID *string, sortOrder int, userID string) (*models.MoneyFeature, error) {
	for i := range r.features {
		if r.features[i].ID == id {
			r.features[i].ParentFeatureID = parentID
			r.features[i].SortOrder = sortOrder
			return &r.features[i], nil
		}
	}
	return nil, ErrMoneyInvalidInput
}
func (r *moneyTreeRepo) ListTrash(ctx context.Context, projectID string) ([]models.MoneyFeature, error) {
	return nil, nil
}
func (r *moneyTreeRepo) ListNotes(ctx context.Context, featureID string) ([]models.MoneyNote, error) {
	return nil, nil
}
func (r *moneyTreeRepo) ListNotesByProject(ctx context.Context, projectID string) ([]models.MoneyNote, error) {
	return nil, nil
}
func (r *moneyTreeRepo) CreateNote(ctx context.Context, note models.MoneyNote) (*models.MoneyNote, error) {
	return nil, nil
}
func (r *moneyTreeRepo) UpdateNote(ctx context.Context, noteID, body, visibility, userID, role string) (*models.MoneyNote, error) {
	return nil, nil
}
func (r *moneyTreeRepo) DeleteNote(ctx context.Context, noteID, userID, role string) error {
	return nil
}
func (r *moneyTreeRepo) CreateUpload(ctx context.Context, upload models.MoneyUpload) (*models.MoneyUpload, error) {
	return nil, nil
}
func (r *moneyTreeRepo) GetUpload(ctx context.Context, id string) (*models.MoneyUpload, error) {
	return nil, nil
}
func (r *moneyTreeRepo) ListUploadsByFeature(ctx context.Context, featureID string) ([]models.MoneyUpload, error) {
	return nil, nil
}
func (r *moneyTreeRepo) ListUploadsByProject(ctx context.Context, projectID string) ([]models.MoneyUpload, error) {
	return nil, nil
}
func (r *moneyTreeRepo) DeleteUpload(ctx context.Context, uploadID, userID, role string) error {
	return nil
}
func (r *moneyTreeRepo) MarkUploadPhysicallyDeleted(ctx context.Context, uploadID string) error {
	return nil
}
func (r *moneyTreeRepo) FeatureNoteCounts(ctx context.Context, projectID string) (map[string]int, error) {
	return nil, nil
}
func (r *moneyTreeRepo) PrimaryUploads(ctx context.Context, projectID string) (map[string]models.MoneyUpload, error) {
	return nil, nil
}

func TestArchiveFeaturePromoteChildrenReparentsDirectChildrenOnly(t *testing.T) {
	repo := &moneyTreeRepo{features: []models.MoneyFeature{
		moneyProjectFeature("root", "", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("area", "root", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("child", "area", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("boulder", "child", models.MoneyFeatureBoulder, models.MoneyStatusScouted, "p1"),
	}}
	svc := NewMoneyService(repo, nil, 0)
	user := models.CurrentUser{ID: "u1", Role: models.RoleDeveloper}

	if err := svc.ArchiveFeatureWithMode(context.Background(), "area", models.MoneyArchiveModePromoteChildren, user); err != nil {
		t.Fatalf("ArchiveFeatureWithMode returned error: %v", err)
	}
	area, _ := repo.GetFeature(context.Background(), "area")
	child, _ := repo.GetFeature(context.Background(), "child")
	boulder, _ := repo.GetFeature(context.Background(), "boulder")
	if area.Status != models.MoneyStatusArchived {
		t.Fatalf("expected selected area archived, got %s", area.Status)
	}
	if child.ParentFeatureID == nil || *child.ParentFeatureID != "root" {
		t.Fatalf("expected child promoted to root, got %v", child.ParentFeatureID)
	}
	if boulder.ParentFeatureID == nil || *boulder.ParentFeatureID != "child" {
		t.Fatalf("expected nested hierarchy preserved, got %v", boulder.ParentFeatureID)
	}
}

func TestMoveFeatureParentRejectsRootSelfDescendantAndArchived(t *testing.T) {
	features := []models.MoneyFeature{
		moneyProjectFeature("root", "", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("area", "root", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("child", "area", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("archived", "root", models.MoneyFeatureArea, models.MoneyStatusArchived, "p1"),
	}
	user := models.CurrentUser{ID: "u1", Role: models.RoleDeveloper}
	cases := []struct{ name, id, parent string }{
		{"root", "root", "area"},
		{"self", "area", "area"},
		{"descendant", "area", "child"},
		{"archived-parent", "area", "archived"},
		{"archived-target", "archived", "area"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &moneyTreeRepo{features: append([]models.MoneyFeature(nil), features...)}
			svc := NewMoneyService(repo, nil, 0)
			parentID := tc.parent
			if _, err := svc.MoveFeatureParent(context.Background(), tc.id, models.MoneyMoveFeatureRequest{ParentFeatureID: &parentID}, user); err != ErrMoneyInvalidInput {
				t.Fatalf("expected invalid input, got %v", err)
			}
		})
	}
}

func TestMoveFeatureParentAllowsValidMove(t *testing.T) {
	repo := &moneyTreeRepo{features: []models.MoneyFeature{
		moneyProjectFeature("root", "", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("area", "root", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
		moneyProjectFeature("target", "root", models.MoneyFeatureArea, models.MoneyStatusActive, "p1"),
	}}
	svc := NewMoneyService(repo, nil, 0)
	parentID := "target"
	sortOrder := 7
	moved, err := svc.MoveFeatureParent(context.Background(), "area", models.MoneyMoveFeatureRequest{ParentFeatureID: &parentID, SortOrder: &sortOrder}, models.CurrentUser{ID: "u1", Role: models.RoleDeveloper})
	if err != nil {
		t.Fatalf("MoveFeatureParent returned error: %v", err)
	}
	if moved.ParentFeatureID == nil || *moved.ParentFeatureID != "target" || moved.SortOrder != 7 {
		t.Fatalf("unexpected moved feature: %+v", moved)
	}
}

func moneyProjectFeature(id, parentID, featureType, status, projectID string) models.MoneyFeature {
	f := moneyTestFeature(id, parentID, featureType, status)
	f.ProjectID = projectID
	return f
}
