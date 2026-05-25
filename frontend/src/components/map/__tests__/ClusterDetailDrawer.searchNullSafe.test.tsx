import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ClusterDetailDrawer } from '../ClusterDetailDrawer';
import { HeatMapPoint } from '../../../types/heatmap';

// Mock the heatMapApi so we can simulate the backend returning `null` for
// `routes` (the upstream condition that produced the original crash).
vi.mock('../../../services/api', () => ({
  heatMapApi: {
    searchRoutesInCluster: vi.fn(),
  },
}));

import { heatMapApi } from '../../../services/api';

const makeArea = (id: number, name: string): HeatMapPoint => ({
  mp_area_id: id,
  name,
  latitude: 40,
  longitude: -105,
  activity_score: 10,
  total_ticks: 5,
  active_routes: 3,
  last_activity: new Date().toISOString(),
  unique_climbers: 2,
  has_subareas: false,
});

describe('ClusterDetailDrawer search null-safety', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('does not throw when the search API returns null for routes', async () => {
    // Simulate the backend returning `{ routes: null }` for an empty result.
    (heatMapApi.searchRoutesInCluster as ReturnType<typeof vi.fn>).mockResolvedValue({
      routes: null,
      count: 0,
    });

    const areas = [makeArea(1, 'Area One'), makeArea(2, 'Area Two')];

    render(
      <ClusterDetailDrawer
        areas={areas}
        isOpen={true}
        onClose={() => {}}
        onAreaClick={() => {}}
        dateRange={{ start: new Date('2024-01-01'), end: new Date('2024-12-31') }}
      />
    );

    const input = screen.getByPlaceholderText(/search areas and routes/i);

    // Typing should NOT throw — mirrors the original crash repro.
    expect(() => {
      fireEvent.change(input, { target: { value: 'crack' } });
    }).not.toThrow();

    // Wait for debounced (300ms) search to fire and resolve; drawer must
    // still render the "no results" UX cleanly rather than crashing into
    // the error boundary when filteredAreas tries to map searchResults.
    await waitFor(
      () => {
        expect(heatMapApi.searchRoutesInCluster).toHaveBeenCalled();
        expect(screen.getByText(/no areas or routes found/i)).toBeTruthy();
      },
      { timeout: 2000 }
    );

    // Sanity: input is still mounted (no error boundary unmount).
    expect(screen.getByPlaceholderText(/search areas and routes/i)).toBeTruthy();
  });
});
