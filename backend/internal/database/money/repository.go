package money

import (
	"context"

	"github.com/alexscott64/woulder/backend/internal/models"
)

type Repository interface {
	GetProjectBySlug(ctx context.Context, slug string) (*models.MoneyProject, error)
	GetProjectByID(ctx context.Context, id string) (*models.MoneyProject, error)
	ListFeatures(ctx context.Context, projectID string, filter models.MoneyFeatureFilter) ([]models.MoneyFeature, error)
	GetFeature(ctx context.Context, id string) (*models.MoneyFeature, error)
	CreateFeature(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error)
	UpdateFeature(ctx context.Context, feature models.MoneyFeature) (*models.MoneyFeature, error)
	ArchiveFeature(ctx context.Context, id, userID string) error
	ListNotes(ctx context.Context, featureID string) ([]models.MoneyNote, error)
	CreateNote(ctx context.Context, note models.MoneyNote) (*models.MoneyNote, error)
	UpdateNote(ctx context.Context, noteID, body, visibility, userID, role string) (*models.MoneyNote, error)
	DeleteNote(ctx context.Context, noteID, userID, role string) error
	CreateUpload(ctx context.Context, upload models.MoneyUpload) (*models.MoneyUpload, error)
	GetUpload(ctx context.Context, id string) (*models.MoneyUpload, error)
	ListUploadsByFeature(ctx context.Context, featureID string) ([]models.MoneyUpload, error)
	DeleteUpload(ctx context.Context, uploadID, userID, role string) error
	FeatureNoteCounts(ctx context.Context, projectID string) (map[string]int, error)
	PrimaryUploads(ctx context.Context, projectID string) (map[string]models.MoneyUpload, error)
}
