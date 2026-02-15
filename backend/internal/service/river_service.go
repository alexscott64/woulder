package service

import (
	"context"
	"fmt"

	"github.com/alexscott64/woulder/backend/internal/database/rivers"
	"github.com/alexscott64/woulder/backend/internal/models"
	riversClient "github.com/alexscott64/woulder/backend/internal/rivers"
)

type RiverService struct {
	riversRepo  rivers.Repository
	riverClient *riversClient.USGSClient
}

func NewRiverService(riversRepo rivers.Repository, client *riversClient.USGSClient) *RiverService {
	return &RiverService{
		riversRepo:  riversRepo,
		riverClient: client,
	}
}

func (s *RiverService) GetRiverDataForLocation(ctx context.Context, locationID int) ([]models.RiverData, error) {
	// 1. Get rivers from database
	locationRivers, err := s.riversRepo.GetByLocation(ctx, locationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rivers for location %d: %w", locationID, err)
	}

	// 2. Fetch current data from USGS
	var riverDataList []models.RiverData
	for _, river := range locationRivers {
		gaugeFlowCFS, gaugeHeightFt, timestamp, err := s.riverClient.GetRiverData(river.GaugeID)
		if err != nil {
			// Log error but continue with other rivers
			continue
		}

		// Apply flow estimation if needed
		actualFlowCFS := gaugeFlowCFS
		if river.IsEstimated {
			if river.FlowDivisor != nil && *river.FlowDivisor > 0 {
				// Simple divisor method
				actualFlowCFS = gaugeFlowCFS / *river.FlowDivisor
			} else if river.DrainageAreaSqMi != nil && river.GaugeDrainageAreaSqMi != nil {
				// Drainage area ratio method
				actualFlowCFS = riversClient.EstimateFlowFromDrainageRatio(
					gaugeFlowCFS,
					*river.DrainageAreaSqMi,
					*river.GaugeDrainageAreaSqMi,
				)
			}
		}

		status, message, isSafe, percentOfSafe := riversClient.CalculateCrossingStatus(river, actualFlowCFS)

		riverDataList = append(riverDataList, models.RiverData{
			River:         river,
			FlowCFS:       actualFlowCFS,
			GaugeHeightFt: gaugeHeightFt,
			IsSafe:        isSafe,
			Status:        status,
			StatusMessage: message,
			Timestamp:     timestamp,
			PercentOfSafe: percentOfSafe,
		})
	}

	return riverDataList, nil
}

func (s *RiverService) GetRiverDataByID(ctx context.Context, riverID int) (*models.RiverData, error) {
	river, err := s.riversRepo.GetByID(ctx, riverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get river %d: %w", riverID, err)
	}

	gaugeFlowCFS, gaugeHeightFt, timestamp, err := s.riverClient.GetRiverData(river.GaugeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get USGS data for river %d: %w", riverID, err)
	}

	// Apply flow estimation if needed
	actualFlowCFS := gaugeFlowCFS
	if river.IsEstimated {
		if river.FlowDivisor != nil && *river.FlowDivisor > 0 {
			actualFlowCFS = gaugeFlowCFS / *river.FlowDivisor
		} else if river.DrainageAreaSqMi != nil && river.GaugeDrainageAreaSqMi != nil {
			actualFlowCFS = riversClient.EstimateFlowFromDrainageRatio(
				gaugeFlowCFS,
				*river.DrainageAreaSqMi,
				*river.GaugeDrainageAreaSqMi,
			)
		}
	}

	status, message, isSafe, percentOfSafe := riversClient.CalculateCrossingStatus(*river, actualFlowCFS)

	return &models.RiverData{
		River:         *river,
		FlowCFS:       actualFlowCFS,
		GaugeHeightFt: gaugeHeightFt,
		IsSafe:        isSafe,
		Status:        status,
		StatusMessage: message,
		Timestamp:     timestamp,
		PercentOfSafe: percentOfSafe,
	}, nil
}
