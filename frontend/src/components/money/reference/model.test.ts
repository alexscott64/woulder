import { describe, expect, it } from 'vitest';
import { bbox, centroid, flattenBoulders, flattenProblems, geometryPoints, polygonGeoJSON } from './model';
import { MoneyCragNode } from '../../../types/money';

const node = (id: string, title: string, type: 'area' | 'boulder' | 'problem', children: MoneyCragNode[] = [], boulders: MoneyCragNode[] = [], problems: MoneyCragNode[] = []): MoneyCragNode => ({
  feature: { id, project_id: 'p', feature_type: type, title, status: type === 'boulder' ? 'scouted' : type === 'problem' ? 'project' : 'active', geojson: polygonGeoJSON([[0, 0], [10, 0], [10, 10]]), style: {}, properties: {}, sort_order: 0, created_by: 'u', updated_by: 'u', created_at: '', updated_at: '' },
  children,
  boulders,
  problems,
});

describe('reference money crag model helpers', () => {
  it('extracts polygon points and bbox in world coordinates', () => {
    const geo = polygonGeoJSON([[2, 4], [8, 4], [8, 9]]);
    expect(geometryPoints(geo)[0]).toEqual([2, 4]);
    expect(bbox({ ...node('a', 'Area', 'area'), feature: { ...node('a', 'Area', 'area').feature, geojson: geo } })).toEqual([2, 4, 8, 9]);
    expect(centroid([[0, 0], [10, 0], [10, 10]])).toEqual([20 / 3, 10 / 3]);
  });

  it('flattens recursive boulders and problems', () => {
    const problem = node('p1', 'Problem', 'problem');
    const boulder = node('b1', 'Boulder', 'boulder', [], [], [problem]);
    const root = node('a1', 'Root', 'area', [node('a2', 'Child', 'area', [], [boulder])]);
    expect(flattenBoulders(root).map(b => b.feature.id)).toEqual(['b1']);
    expect(flattenProblems(root).map(p => p.feature.id)).toEqual(['p1']);
  });
});
