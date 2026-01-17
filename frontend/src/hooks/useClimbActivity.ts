import { useQuery } from '@tanstack/react-query';
import { climbActivityApi } from '../services/api';
import type { AreaActivitySummary, RouteActivitySummary } from '../types/weather';

/**
 * Hook to fetch areas ordered by recent climb activity for a location
 */
export const useAreasOrderedByActivity = (locationId: number) => {
  return useQuery<AreaActivitySummary[], Error>({
    queryKey: ['areas-activity', locationId],
    queryFn: () => climbActivityApi.getAreasOrderedByActivity(locationId),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!locationId,
  });
};

/**
 * Hook to fetch subareas of a parent area ordered by recent climb activity
 */
export const useSubareasOrderedByActivity = (
  locationId: number,
  areaId: string | null
) => {
  return useQuery<AreaActivitySummary[], Error>({
    queryKey: ['subareas-activity', locationId, areaId],
    queryFn: () => climbActivityApi.getSubareasOrderedByActivity(locationId, areaId!),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!locationId && !!areaId,
  });
};

/**
 * Hook to fetch routes in an area ordered by recent climb activity
 */
export const useRoutesOrderedByActivity = (
  locationId: number,
  areaId: string | null,
  limit = 50
) => {
  return useQuery<RouteActivitySummary[], Error>({
    queryKey: ['routes-activity', locationId, areaId, limit],
    queryFn: () => climbActivityApi.getRoutesOrderedByActivity(locationId, areaId!, limit),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!locationId && !!areaId,
  });
};
