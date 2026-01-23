import { useQuery, useQueries } from '@tanstack/react-query';
import { climbActivityApi } from '../services/api';
import type { AreaActivitySummary, RouteActivitySummary, ClimbHistoryEntry, SearchResult, BoulderDryingStatus } from '../types/weather';

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
  limit = 200
) => {
  return useQuery<RouteActivitySummary[], Error>({
    queryKey: ['routes-activity', locationId, areaId, limit],
    queryFn: () => climbActivityApi.getRoutesOrderedByActivity(locationId, areaId!, limit),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!locationId && !!areaId,
  });
};

/**
 * Hook to fetch recent ticks for a specific route
 */
export const useRecentTicksForRoute = (
  routeId: string | null,
  limit = 5
) => {
  return useQuery<ClimbHistoryEntry[], Error>({
    queryKey: ['route-ticks', routeId, limit],
    queryFn: () => climbActivityApi.getRecentTicksForRoute(routeId!, limit),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!routeId,
  });
};

/**
 * Hook to search all areas and routes in a location by name
 */
export const useSearchInLocation = (
  locationId: number,
  searchQuery: string,
  limit = 50
) => {
  return useQuery<SearchResult[], Error>({
    queryKey: ['search-all', locationId, searchQuery, limit],
    queryFn: () => climbActivityApi.searchInLocation(locationId, searchQuery, limit),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!locationId && searchQuery.length >= 2, // Only search if query is at least 2 chars
  });
};

/**
 * Hook to search all routes in a location by name, grade, or area
 */
export const useSearchRoutesInLocation = (
  locationId: number,
  searchQuery: string,
  limit = 50
) => {
  return useQuery<RouteActivitySummary[], Error>({
    queryKey: ['search-routes', locationId, searchQuery, limit],
    queryFn: () => climbActivityApi.searchRoutesInLocation(locationId, searchQuery, limit),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!locationId && searchQuery.length >= 2, // Only search if query is at least 2 chars
  });
};

/**
 * Hook to fetch boulder-specific drying status for a route
 */
export const useBoulderDryingStatus = (routeId: string | null) => {
  return useQuery<BoulderDryingStatus, Error>({
    queryKey: ['boulder-drying', routeId],
    queryFn: () => climbActivityApi.getBoulderDryingStatus(routeId!),
    staleTime: 10 * 60 * 1000, // 10 minutes
    enabled: !!routeId,
  });
};

/**
 * Hook to fetch boulder-specific drying status for multiple routes
 */
export const useBoulderDryingStatuses = (routeIds: string[]) => {
  return useQueries({
    queries: routeIds.map(routeId => ({
      queryKey: ['boulder-drying', routeId],
      queryFn: () => climbActivityApi.getBoulderDryingStatus(routeId),
      staleTime: 10 * 60 * 1000, // 10 minutes
      enabled: !!routeId,
    })),
  });
};
