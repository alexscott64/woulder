import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AreaDrillDownView } from '../AreaDrillDownView';
import * as useClimbActivity from '../../hooks/useClimbActivity';
import type { AreaActivitySummary, AreaDryingStats, BoulderDryingStatus } from '../../types/weather';

// Mock the hooks
vi.mock('../../hooks/useClimbActivity');

describe('AreaDrillDownView - Drying Stats Type Handling', () => {
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

  it('should correctly map area drying stats with number keys', async () => {
    // Mock areas data with numeric mp_area_id (as returned by backend)
    const mockAreas: AreaActivitySummary[] = [
      {
        mp_area_id: 108123672, // NUMBER type
        name: 'Lower Wall',
        parent_mp_area_id: 108123669,
        last_climb_at: new Date('2026-01-25T08:00:00Z'),
        total_ticks: 10,
        unique_routes: 5,
        days_since_climb: 2,
        has_subareas: false,
        subarea_count: 0,
      },
      {
        mp_area_id: 108127553, // NUMBER type
        name: 'River Boulders',
        parent_mp_area_id: 108123669,
        last_climb_at: new Date('2026-01-24T08:00:00Z'),
        total_ticks: 70,
        unique_routes: 20,
        days_since_climb: 3,
        has_subareas: true,
        subarea_count: 5,
      },
    ];

    // Mock batch drying stats with STRING keys (as returned by Object.entries)
    // This simulates the API response structure: Record<string, AreaDryingStats>
    const mockDryingStats: Record<string, AreaDryingStats> = {
      '108123672': {
        total_routes: 10,
        dry_count: 8,
        drying_count: 2,
        wet_count: 0,
        percent_dry: 80,
        avg_hours_until_dry: 1.5,
        avg_tree_coverage: 50.7,
        confidence_score: 100,
      },
      '108127553': {
        total_routes: 70,
        dry_count: 35,
        drying_count: 35,
        wet_count: 0,
        percent_dry: 50,
        avg_hours_until_dry: 2.8,
        avg_tree_coverage: 66.2,
        confidence_score: 100,
      },
    };

    // Mock the hooks
    vi.spyOn(useClimbActivity, 'useAreasOrderedByActivity').mockReturnValue({
      data: mockAreas,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useMultipleAreaDryingStats').mockReturnValue({
      data: mockDryingStats,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useSubareasOrderedByActivity').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useRoutesOrderedByActivity').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useBoulderDryingStatuses').mockReturnValue({
      data: {},
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useSearchInLocation').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useAreaDryingStats').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    // Render component
    render(
      <AreaDrillDownView
        locationId={2}
        onNavigateToArea={() => {}}
        onNavigateBack={() => {}}
        currentPath={[]}
      />,
      { wrapper }
    );

    // Wait for areas to render
    await waitFor(() => {
      expect(screen.getByText('Lower Wall')).toBeTruthy();
      expect(screen.getByText('River Boulders')).toBeTruthy();
    });

    // CRITICAL TEST: Verify drying stats are displayed
    // This tests that the Map correctly converts string keys to numbers
    // so that lookups with numeric mp_area_id values work correctly
    await waitFor(() => {
      // Check for "80% Dry" text (from first area with 80% dry)
      expect(screen.getByText('80% Dry')).toBeTruthy();
      // Check for "50% Dry" text (from second area with 50% dry)
      expect(screen.getByText('50% Dry')).toBeTruthy();
    });

    // Verify confidence scores are shown (both areas have 100% confidence)
    const confidenceElements = screen.getAllByText('100% confident');
    expect(confidenceElements.length).toBeGreaterThanOrEqual(1);
  });

  it.skip('should correctly map route drying statuses with number keys', async () => {
    // Mock route data with numeric mp_route_id (as returned by backend)
    const mockRoutes = [
      {
        mp_route_id: 114417549, // NUMBER type
        name: 'Test Boulder',
        rating: 'V3',
        mp_area_id: 108123672,
        last_climb_at: new Date('2026-01-25T08:00:00Z'),
        days_since_climb: 2,
        most_recent_tick: null,
      },
    ];

    // Mock batch boulder drying status with STRING keys (as returned by Object.entries)
    const mockBoulderDryingStatus: Record<string, BoulderDryingStatus> = {
      '114417549': {
        is_dry: false,
        status: 'drying',
        hours_until_dry: 4.5,
        last_rain_timestamp: new Date('2026-02-03T06:00:00Z'),
        confidence_score: 100,
        tree_coverage_percent: 45.0,
        aspect: 'N',
      },
    };

    // Mock the hooks for route view
    vi.spyOn(useClimbActivity, 'useAreasOrderedByActivity').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useSubareasOrderedByActivity').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useRoutesOrderedByActivity').mockReturnValue({
      data: mockRoutes,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useBoulderDryingStatuses').mockReturnValue({
      data: mockBoulderDryingStatus,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useMultipleAreaDryingStats').mockReturnValue({
      data: {},
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useSearchInLocation').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useAreaDryingStats').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    // Render component in route view (currentAreaId set to trigger route display)
    render(
      <AreaDrillDownView
        locationId={2}
        onNavigateToArea={() => {}}
        onNavigateBack={() => {}}
        currentPath={[{ id: '108123672', name: 'Lower Wall' }]}
      />,
      { wrapper }
    );

    // Wait for route to render
    await waitFor(() => {
      expect(screen.getByText('Test Boulder')).toBeTruthy();
    }, { timeout: 3000 });

    // CRITICAL TEST: Verify boulder drying status is displayed
    // This tests that the Map correctly converts string keys to numbers
    // so that lookups with numeric mp_route_id values work correctly
    await waitFor(() => {
      // The exact text will depend on the component's display logic
      // But it should show the "drying" status somewhere
      const dryingElements = screen.queryAllByText(/drying|4.5/i);
      expect(dryingElements.length).toBeGreaterThan(0);
    });
  });

  it('should handle empty drying stats gracefully', async () => {
    const mockAreas: AreaActivitySummary[] = [
      {
        mp_area_id: 108123672,
        name: 'Area Without Stats',
        parent_mp_area_id: 108123669,
        last_climb_at: new Date('2026-01-25T08:00:00Z'),
        total_ticks: 10,
        unique_routes: 5,
        days_since_climb: 2,
        has_subareas: false,
        subarea_count: 0,
      },
    ];

    // Mock empty drying stats
    vi.spyOn(useClimbActivity, 'useAreasOrderedByActivity').mockReturnValue({
      data: mockAreas,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useMultipleAreaDryingStats').mockReturnValue({
      data: {}, // Empty stats
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useSubareasOrderedByActivity').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useRoutesOrderedByActivity').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useBoulderDryingStatuses').mockReturnValue({
      data: {},
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useSearchInLocation').mockReturnValue({
      data: [],
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    vi.spyOn(useClimbActivity, 'useAreaDryingStats').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    render(
      <AreaDrillDownView
        locationId={2}
        onNavigateToArea={() => {}}
        onNavigateBack={() => {}}
        currentPath={[]}
      />,
      { wrapper }
    );

    // Area should still render even without drying stats
    await waitFor(() => {
      expect(screen.getByText('Area Without Stats')).toBeTruthy();
    });

    // Drying stats section should not be present
    expect(screen.queryByText(/% Dry/)).toBeNull();
  });
});
