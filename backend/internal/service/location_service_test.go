package service

import (
	"context"
	"errors"
	"testing"

	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestLocationService_GetAllLocations(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context) ([]models.Location, error)
		want    int
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func(ctx context.Context) ([]models.Location, error) {
				return []models.Location{
					{ID: 1, Name: "Location 1"},
					{ID: 2, Name: "Location 2"},
				}, nil
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "database error",
			mockFn: func(ctx context.Context) ([]models.Location, error) {
				return nil, errors.New("database error")
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationsRepo := &MockLocationsRepository{
				GetAllFn: tt.mockFn,
			}
			mockAreasRepo := &MockAreasRepository{}

			service := NewLocationService(mockLocationsRepo, mockAreasRepo)
			locations, err := service.GetAllLocations(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, locations, tt.want)
			}
		})
	}
}

func TestLocationService_GetLocation(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, id int) (*models.Location, error)
		id      int
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func(ctx context.Context, id int) (*models.Location, error) {
				return &models.Location{
					ID:   id,
					Name: "Test Location",
				}, nil
			},
			id:      1,
			wantErr: false,
		},
		{
			name: "not found",
			mockFn: func(ctx context.Context, id int) (*models.Location, error) {
				return nil, errors.New("not found")
			},
			id:      999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationsRepo := &MockLocationsRepository{
				GetByIDFn: tt.mockFn,
			}
			mockAreasRepo := &MockAreasRepository{}

			service := NewLocationService(mockLocationsRepo, mockAreasRepo)
			location, err := service.GetLocation(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, location)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, location)
				assert.Equal(t, tt.id, location.ID)
			}
		})
	}
}

func TestLocationService_GetLocationsByArea(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, areaID int) ([]models.Location, error)
		areaID  int
		want    int
		wantErr bool
	}{
		{
			name: "success with results",
			mockFn: func(ctx context.Context, areaID int) ([]models.Location, error) {
				return []models.Location{
					{ID: 1, Name: "Location 1", AreaID: areaID},
					{ID: 2, Name: "Location 2", AreaID: areaID},
				}, nil
			},
			areaID:  1,
			want:    2,
			wantErr: false,
		},
		{
			name: "success with empty results",
			mockFn: func(ctx context.Context, areaID int) ([]models.Location, error) {
				return []models.Location{}, nil
			},
			areaID:  999,
			want:    0,
			wantErr: false,
		},
		{
			name: "database error",
			mockFn: func(ctx context.Context, areaID int) ([]models.Location, error) {
				return nil, errors.New("database error")
			},
			areaID:  1,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationsRepo := &MockLocationsRepository{
				GetByAreaFn: tt.mockFn,
			}
			mockAreasRepo := &MockAreasRepository{}

			service := NewLocationService(mockLocationsRepo, mockAreasRepo)
			locations, err := service.GetLocationsByArea(context.Background(), tt.areaID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, locations, tt.want)
			}
		})
	}
}

func TestLocationService_GetAllAreas(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context) ([]models.Area, error)
		want    int
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func(ctx context.Context) ([]models.Area, error) {
				return []models.Area{
					{ID: 1, Name: "Area 1"},
					{ID: 2, Name: "Area 2"},
				}, nil
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "database error",
			mockFn: func(ctx context.Context) ([]models.Area, error) {
				return nil, errors.New("database error")
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationsRepo := &MockLocationsRepository{}
			mockAreasRepo := &MockAreasRepository{
				GetAllFn: tt.mockFn,
			}

			service := NewLocationService(mockLocationsRepo, mockAreasRepo)
			areas, err := service.GetAllAreas(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, areas, tt.want)
			}
		})
	}
}

func TestLocationService_GetAreasWithLocationCounts(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context) ([]models.AreaWithLocationCount, error)
		want    int
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func(ctx context.Context) ([]models.AreaWithLocationCount, error) {
				return []models.AreaWithLocationCount{
					{Area: models.Area{ID: 1, Name: "Area 1"}, LocationCount: 5},
					{Area: models.Area{ID: 2, Name: "Area 2"}, LocationCount: 3},
				}, nil
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "database error",
			mockFn: func(ctx context.Context) ([]models.AreaWithLocationCount, error) {
				return nil, errors.New("database error")
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationsRepo := &MockLocationsRepository{}
			mockAreasRepo := &MockAreasRepository{
				GetAllWithLocationCountsFn: tt.mockFn,
			}

			service := NewLocationService(mockLocationsRepo, mockAreasRepo)
			areas, err := service.GetAreasWithLocationCounts(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, areas, tt.want)
			}
		})
	}
}

func TestLocationService_GetAreaByID(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, id int) (*models.Area, error)
		id      int
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func(ctx context.Context, id int) (*models.Area, error) {
				return &models.Area{
					ID:   id,
					Name: "Test Area",
				}, nil
			},
			id:      1,
			wantErr: false,
		},
		{
			name: "not found",
			mockFn: func(ctx context.Context, id int) (*models.Area, error) {
				return nil, errors.New("not found")
			},
			id:      999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocationsRepo := &MockLocationsRepository{}
			mockAreasRepo := &MockAreasRepository{
				GetByIDFn: tt.mockFn,
			}

			service := NewLocationService(mockLocationsRepo, mockAreasRepo)
			area, err := service.GetAreaByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, area)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, area)
				assert.Equal(t, tt.id, area.ID)
			}
		})
	}
}
