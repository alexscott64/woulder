import { describe, expect, it } from 'vitest';
import { parseGeoJSONTrail, parseGPXTrail } from './trailImport';

describe('trailImport', () => {
  it('parses GPX track points into a LineString', () => {
    const parsed = parseGPXTrail('<gpx><trk><name>North Fork</name><trkseg><trkpt lat="47.700" lon="-121.500"/><trkpt lat="47.710" lon="-121.510"/></trkseg></trk></gpx>', 'north.gpx');

    expect(parsed).toEqual({
      title: 'North Fork',
      sourceFormat: 'gpx',
      pointCount: 2,
      geojson: { type: 'LineString', coordinates: [[-121.5, 47.7], [-121.51, 47.71]] },
    });
  });

  it('parses GeoJSON feature LineString names', () => {
    const parsed = parseGeoJSONTrail(JSON.stringify({ type: 'Feature', properties: { name: 'Main Wall trail' }, geometry: { type: 'LineString', coordinates: [[-121.5, 47.7], [-121.51, 47.71]] } }), 'fallback.geojson');

    expect(parsed.title).toBe('Main Wall trail');
    expect(parsed.sourceFormat).toBe('geojson');
    expect(parsed.geojson).toEqual({ type: 'LineString', coordinates: [[-121.5, 47.7], [-121.51, 47.71]] });
  });

  it('rejects files without a usable trail line', () => {
    expect(() => parseGeoJSONTrail('{"type":"Point","coordinates":[-121.5,47.7]}')).toThrow(/LineString/);
  });
});
