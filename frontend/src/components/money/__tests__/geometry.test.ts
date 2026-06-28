import { describe, expect, it } from 'vitest';
import { buildGeoJSON, getFeatureCoordinates, minimumPointCount } from '../geometry';
import { MoneyFeature } from '../../../types/money';

const baseFeature: MoneyFeature = {
  id: 'feature-1',
  project_id: 'project-1',
  feature_type: 'trail',
  title: 'Trail',
  status: 'draft',
  geojson: { type: 'LineString', coordinates: [[-121.5, 47.7], [-121.51, 47.71]] },
  style: {},
  properties: {},
  created_by: 'user-1',
  updated_by: 'user-1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

describe('money geometry helpers', () => {
  it('builds a closed polygon for topo drawings', () => {
    const geojson = buildGeoJSON('topo', [[-121.5, 47.7], [-121.51, 47.71], [-121.49, 47.72]]);
    expect(geojson.type).toBe('Polygon');
    if (geojson.type === 'Polygon') {
      expect(geojson.coordinates[0][0]).toEqual(geojson.coordinates[0][geojson.coordinates[0].length - 1]);
    }
  });

  it('extracts feature coordinates from line geometry', () => {
    expect(getFeatureCoordinates(baseFeature)).toEqual([[-121.5, 47.7], [-121.51, 47.71]]);
  });

  it('sets minimum point counts by feature type', () => {
    expect(minimumPointCount('poi')).toBe(1);
    expect(minimumPointCount('trail')).toBe(2);
    expect(minimumPointCount('topo')).toBe(3);
  });
});
