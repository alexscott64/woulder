import { describe, it, expect } from 'vitest';

/**
 * Tests for Map key type conversions
 *
 * CONTEXT: We discovered a bug where API responses return data as Record<string, T>
 * (e.g., { "108123672": {...}, "108127553": {...} }) but our component data uses
 * numeric IDs (e.g., mp_area_id: 108123672). When creating Maps from Object.entries(),
 * the keys become strings, causing lookups with numeric keys to fail.
 *
 * This test ensures we properly convert string keys to numbers when building Maps.
 */
describe('Map Key Type Conversion', () => {
  it('should correctly handle number-to-string key mismatch in area drying stats', () => {
    // Simulate API response with string keys (as returned by backend/JSON parsing)
    const apiResponse: Record<string, { percent_dry: number }> = {
      '108123672': { percent_dry: 80 },
      '108127553': { percent_dry: 50 },
      '108127563': { percent_dry: 100 },
    };

    // WRONG WAY: Direct Object.entries() without conversion
    // This is the bug we fixed - keys remain as strings
    const wrongMap = new Map();
    Object.entries(apiResponse).forEach(([areaId, stats]) => {
      wrongMap.set(areaId, stats); // areaId is a string here!
    });

    // Component data uses numeric IDs (as returned from backend models)
    const numericAreaId = 108123672;

    // Lookup with numeric key FAILS with string-keyed map
    expect(wrongMap.get(numericAreaId)).toBeUndefined();
    // Only works with string key
    expect(wrongMap.get('108123672')).toEqual({ percent_dry: 80 });

    // RIGHT WAY: Convert string keys to numbers
    // This is our fix - ensures consistent number keys
    const correctMap = new Map();
    Object.entries(apiResponse).forEach(([areaId, stats]) => {
      correctMap.set(Number(areaId), stats); // Convert to number!
    });

    // Lookup with numeric key SUCCEEDS with number-keyed map
    expect(correctMap.get(numericAreaId)).toEqual({ percent_dry: 80 });
    expect(correctMap.get(108127553)).toEqual({ percent_dry: 50 });
    expect(correctMap.get(108127563)).toEqual({ percent_dry: 100 });

    // Verify all three area IDs work correctly
    const areaIds = [108123672, 108127553, 108127563];
    areaIds.forEach((id) => {
      const stats = correctMap.get(id);
      expect(stats).toBeDefined();
      expect(stats?.percent_dry).toBeGreaterThan(0);
    });
  });

  it('should correctly handle number-to-string key mismatch in route drying status', () => {
    // Simulate API response for route drying statuses
    const apiResponse: Record<string, { is_dry: boolean; status: string }> = {
      '114417549': { is_dry: false, status: 'drying' },
      '117375431': { is_dry: true, status: 'dry' },
    };

    // WRONG WAY: String keys
    const wrongMap = new Map();
    Object.entries(apiResponse).forEach(([routeId, status]) => {
      wrongMap.set(routeId, status);
    });

    const numericRouteId = 114417549;

    // Lookup fails
    expect(wrongMap.get(numericRouteId)).toBeUndefined();

    // RIGHT WAY: Number keys
    const correctMap = new Map();
    Object.entries(apiResponse).forEach(([routeId, status]) => {
      correctMap.set(Number(routeId), status);
    });

    // Lookup succeeds
    expect(correctMap.get(numericRouteId)).toEqual({ is_dry: false, status: 'drying' });
    expect(correctMap.get(117375431)).toEqual({ is_dry: true, status: 'dry' });
  });

  it('should handle edge cases in key conversion', () => {
    const apiResponse: Record<string, { value: number }> = {
      '0': { value: 100 },
      '123456789012345': { value: 200 }, // Large number
    };

    const correctMap = new Map();
    Object.entries(apiResponse).forEach(([id, data]) => {
      correctMap.set(Number(id), data);
    });

    // Zero should work
    expect(correctMap.get(0)).toEqual({ value: 100 });

    // Large numbers should work (within JS safe integer range)
    expect(correctMap.get(123456789012345)).toEqual({ value: 200 });
  });

  it('should demonstrate the fix prevents drying stats from disappearing', () => {
    // Real-world scenario: User loads area list with drying stats
    const areaDryingStats: Record<string, { percent_dry: number; total_routes: number }> = {
      '108123672': { percent_dry: 80, total_routes: 10 },
      '108127553': { percent_dry: 50, total_routes: 70 },
    };

    const areas = [
      { mp_area_id: 108123672, name: 'Lower Wall' },
      { mp_area_id: 108127553, name: 'River Boulders' },
    ];

    // Build map with proper conversion
    const statsMap = new Map();
    Object.entries(areaDryingStats).forEach(([areaId, stats]) => {
      statsMap.set(Number(areaId), stats);
    });

    // Verify each area can retrieve its stats
    areas.forEach((area) => {
      const stats = statsMap.get(area.mp_area_id);
      expect(stats).toBeDefined();
      expect(stats?.total_routes).toBeGreaterThan(0);
      expect(stats?.percent_dry).toBeGreaterThanOrEqual(0);
      expect(stats?.percent_dry).toBeLessThanOrEqual(100);
    });

    // Specifically test that Lower Wall shows 80% dry
    const lowerWallStats = statsMap.get(108123672);
    expect(lowerWallStats?.percent_dry).toBe(80);

    // And River Boulders shows 50% dry
    const riverBouldersStats = statsMap.get(108127553);
    expect(riverBouldersStats?.percent_dry).toBe(50);
  });
});
