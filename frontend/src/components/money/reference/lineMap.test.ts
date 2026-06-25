import { describe, expect, it } from 'vitest';
import { LINE_MAP_CATEGORIES, LINE_MAP_BOUNDS, isBoundedLonLat, moneyCreekLineMapByCategory, moneyCreekLineMapLabels, moneyCreekLineMapPaths } from './lineMap';

describe('Money Creek line map fixture', () => {
  it('loads every supported line category needed by the stylized map', () => {
    for (const category of LINE_MAP_CATEGORIES) {
      expect(moneyCreekLineMapByCategory[category].length, category).toBeGreaterThan(0);
    }
  });

  it('normalizes regenerated source features into DeckGL path data with valid lon/lat positions', () => {
    expect(moneyCreekLineMapPaths.length).toBeGreaterThan(3000);
    const sampledPaths = moneyCreekLineMapPaths.filter((_, index) => index % 50 === 0);
    for (const feature of sampledPaths) {
      expect(feature.path.length, feature.id).toBeGreaterThanOrEqual(2);
      for (const position of feature.path.slice(0, 8)) {
        expect(isBoundedLonLat(position), `${feature.id} ${position.join(',')}`).toBe(true);
      }
    }
  });

  it('covers the crag-to-reservoir corridor without out-of-bounds coordinates', () => {
    const extent = moneyCreekLineMapPaths.reduce((box, feature) => {
      for (const [lon, lat] of feature.path) {
        box[0] = Math.min(box[0], lon);
        box[1] = Math.min(box[1], lat);
        box[2] = Math.max(box[2], lon);
        box[3] = Math.max(box[3], lat);
      }
      return box;
    }, [Number.POSITIVE_INFINITY, Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY, Number.NEGATIVE_INFINITY]);

    expect(extent[0]).toBeGreaterThanOrEqual(LINE_MAP_BOUNDS[0]);
    expect(extent[2]).toBeLessThanOrEqual(LINE_MAP_BOUNDS[2]);
    expect(extent[1]).toBeGreaterThanOrEqual(LINE_MAP_BOUNDS[1]);
    expect(extent[3]).toBeLessThanOrEqual(LINE_MAP_BOUNDS[3]);
    expect(extent[0]).toBeLessThan(-121.64);
    expect(extent[2]).toBeGreaterThan(-121.38);
    expect(extent[1]).toBeLessThan(47.64);
    expect(extent[3]).toBeGreaterThan(47.81);
  });

  it('preserves regenerated source categories and useful road/hydro metadata', () => {
    expect(moneyCreekLineMapByCategory.road.length).toBeGreaterThan(50);
    expect(moneyCreekLineMapByCategory.creek.length).toBeGreaterThan(150);
    expect(moneyCreekLineMapByCategory.reservoir.length).toBeGreaterThan(20);
    expect(moneyCreekLineMapByCategory.road.some(path => path.sourceKind === 'trail' && path.importance === 'minor' && path.surface === 'dirt')).toBe(true);
    expect(moneyCreekLineMapByCategory.creek.some(path => path.source?.endsWith('hydro.geojson') && path.seasonal === false)).toBe(true);
    expect(moneyCreekLineMapByCategory.reservoir.every(path => path.category === 'reservoir')).toBe(true);
  });

  it('provides visible label metadata for stylized roads, water, and crag context', () => {
    expect(moneyCreekLineMapLabels.length).toBeGreaterThanOrEqual(30);
    expect(moneyCreekLineMapLabels.map(label => label.name)).toEqual(expect.arrayContaining([
      'Money Creek',
      'Money Creek Road',
      'South Fork Tolt River',
      'Cleveland Lake',
      'Lake Elizabeth',
    ]));

    const kinds = new Set(moneyCreekLineMapLabels.map(label => label.kind));
    expect(kinds.has('road')).toBe(true);
    expect(kinds.has('creek')).toBe(true);
    expect(kinds.has('reservoir')).toBe(true);
    expect(kinds.has('context')).toBe(true);
    expect(moneyCreekLineMapLabels.some(label => label.priority === 'high' && label.minZoom === 11)).toBe(true);
    expect(moneyCreekLineMapLabels.some(label => label.labelType === 'reservoir' && label.kind === 'reservoir')).toBe(true);
    for (const label of moneyCreekLineMapLabels) {
      expect(label.name.trim().length, label.id).toBeGreaterThan(0);
      expect(label.minZoom, label.id).toBeGreaterThanOrEqual(11);
      expect(isBoundedLonLat(label.position), `${label.id} ${label.position.join(',')}`).toBe(true);
    }
  });

  it('keeps contours unobtrusive but mountain-like with real elevations, intervals, and index flags', () => {
    const allContours = [...moneyCreekLineMapByCategory.contour, ...moneyCreekLineMapByCategory['index-contour']];
    expect(allContours.length).toBeGreaterThan(2500);
    expect(moneyCreekLineMapByCategory['index-contour'].length).toBeGreaterThan(400);
    expect(moneyCreekLineMapByCategory['index-contour'].every(path => typeof path.elevationM === 'number')).toBe(true);
    expect(allContours.every(path => path.intervalM === 10)).toBe(true);
    expect(Math.min(...allContours.map(path => path.elevationM ?? Number.POSITIVE_INFINITY))).toBeLessThanOrEqual(240);
    expect(Math.max(...allContours.map(path => path.elevationM ?? Number.NEGATIVE_INFINITY))).toBeGreaterThanOrEqual(1700);
    expect(moneyCreekLineMapByCategory.contour.some(path => path.path.length >= 13 && path.path[0][0] === path.path[path.path.length - 1][0] && path.path[0][1] === path.path[path.path.length - 1][1])).toBe(true);
  });
});
