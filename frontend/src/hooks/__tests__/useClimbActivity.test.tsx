import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAreasOrderedByActivity, useSubareasOrderedByActivity, useRoutesOrderedByActivity } from '../useClimbActivity';
import * as api from '../../services/api';
import type { AreaActivitySummary, RouteActivitySummary } from '../../types/weather';

// Mock the API module
vi.mock('../../services/api');

describe('useClimbActivity hooks', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });
    vi.clearAllMocks();
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );

  describe('useAreasOrderedByActivity', () => {
    it('should fetch areas for a location', async () => {
      const mockAreas: AreaActivitySummary[] = [
        {
          mp_area_id: '123',
          name: 'Test Area',
          last_climb_at: '2025-01-15T10:00:00Z',
          total_ticks: 10,
          unique_routes: 5,
          days_since_climb: 2,
          has_subareas: true,
          subarea_count: 3,
        },
      ];

      vi.spyOn(api.climbActivityApi, 'getAreasOrderedByActivity').mockResolvedValue(mockAreas);

      const { result } = renderHook(() => useAreasOrderedByActivity(1), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockAreas);
      expect(api.climbActivityApi.getAreasOrderedByActivity).toHaveBeenCalledWith(1);
    });

    it('should not fetch when locationId is falsy', () => {
      const { result } = renderHook(() => useAreasOrderedByActivity(0), { wrapper });

      expect(result.current.isPending).toBe(true);
      expect(result.current.fetchStatus).toBe('idle');
      expect(api.climbActivityApi.getAreasOrderedByActivity).not.toHaveBeenCalled();
    });

    it('should handle errors', async () => {
      const error = new Error('Failed to fetch areas');
      vi.spyOn(api.climbActivityApi, 'getAreasOrderedByActivity').mockRejectedValue(error);

      const { result } = renderHook(() => useAreasOrderedByActivity(1), { wrapper });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toBe(error);
    });
  });

  describe('useSubareasOrderedByActivity', () => {
    it('should fetch subareas for a location and area', async () => {
      const mockSubareas: AreaActivitySummary[] = [
        {
          mp_area_id: '456',
          name: 'Test Subarea',
          parent_mp_area_id: '123',
          last_climb_at: '2025-01-15T10:00:00Z',
          total_ticks: 5,
          unique_routes: 3,
          days_since_climb: 2,
          has_subareas: false,
          subarea_count: 0,
        },
      ];

      vi.spyOn(api.climbActivityApi, 'getSubareasOrderedByActivity').mockResolvedValue(mockSubareas);

      const { result } = renderHook(() => useSubareasOrderedByActivity(1, '123'), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockSubareas);
      expect(api.climbActivityApi.getSubareasOrderedByActivity).toHaveBeenCalledWith(1, '123');
    });

    it('should not fetch when areaId is null', () => {
      const { result } = renderHook(() => useSubareasOrderedByActivity(1, null), { wrapper });

      expect(result.current.isPending).toBe(true);
      expect(result.current.fetchStatus).toBe('idle');
      expect(api.climbActivityApi.getSubareasOrderedByActivity).not.toHaveBeenCalled();
    });

    it('should not fetch when locationId is falsy', () => {
      const { result } = renderHook(() => useSubareasOrderedByActivity(0, '123'), { wrapper });

      expect(result.current.isPending).toBe(true);
      expect(result.current.fetchStatus).toBe('idle');
      expect(api.climbActivityApi.getSubareasOrderedByActivity).not.toHaveBeenCalled();
    });
  });

  describe('useRoutesOrderedByActivity', () => {
    it('should fetch routes for a location and area', async () => {
      const mockRoutes: RouteActivitySummary[] = [
        {
          mp_route_id: '789',
          name: 'Test Route',
          rating: 'V4',
          mp_area_id: '123',
          last_climb_at: '2025-01-15T10:00:00Z',
          most_recent_tick: {
            mp_route_id: '789',
            route_name: 'Test Route',
            route_rating: 'V4',
            mp_area_id: '123',
            area_name: 'Test Area',
            climbed_at: '2025-01-15T10:00:00Z',
            climbed_by: 'John Doe',
            style: 'Flash',
            comment: 'Great problem!',
            days_since_climb: 2,
          },
          days_since_climb: 2,
        },
      ];

      vi.spyOn(api.climbActivityApi, 'getRoutesOrderedByActivity').mockResolvedValue(mockRoutes);

      const { result } = renderHook(() => useRoutesOrderedByActivity(1, '123', 50), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockRoutes);
      expect(api.climbActivityApi.getRoutesOrderedByActivity).toHaveBeenCalledWith(1, '123', 50);
    });

    it('should use default limit of 200', async () => {
      vi.spyOn(api.climbActivityApi, 'getRoutesOrderedByActivity').mockResolvedValue([]);

      const { result } = renderHook(() => useRoutesOrderedByActivity(1, '123'), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(api.climbActivityApi.getRoutesOrderedByActivity).toHaveBeenCalledWith(1, '123', 200);
    });

    it('should not fetch when areaId is null', () => {
      const { result } = renderHook(() => useRoutesOrderedByActivity(1, null), { wrapper });

      expect(result.current.isPending).toBe(true);
      expect(result.current.fetchStatus).toBe('idle');
      expect(api.climbActivityApi.getRoutesOrderedByActivity).not.toHaveBeenCalled();
    });
  });
});
