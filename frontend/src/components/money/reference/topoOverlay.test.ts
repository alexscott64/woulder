import { describe, expect, it } from 'vitest';
import { MoneyFeature } from '../../../types/money';
import { NormalizedPoint, TopoOverlay, displayStartMarkerLabel, labelForStartMarkerType, overlayPathD, scalePoint, startMarkerSegments, upsertTopoOverlay } from './topoOverlay';

describe('topo overlay helpers', () => {
  it('renders normalized coordinates at any target scale', () => {
    expect(scalePoint([0.25, 0.5], 800, 600)).toEqual([200, 300]);
    expect(overlayPathD([[0, 0], [0.5, 0.25], [1, 1]], 400, 200)).toBe('M 0.0 0.0 L 200.0 50.0 L 400.0 200.0');
  });

  it('scales normalized start markers with the target photo size', () => {
    expect(startMarkerSegments([0.25, 0.5], 20, 800, 600)).toEqual([
      [190, 290, 210, 310],
      [210, 290, 190, 310],
    ]);
  });

  it('labels new and legacy start markers for display', () => {
    expect(labelForStartMarkerType('generic')).toBe('X');
    expect(labelForStartMarkerType('left')).toBe('L');
    expect(labelForStartMarkerType('right')).toBe('R');
    expect(displayStartMarkerLabel({ id: 'start-l', point: [0.2, 0.7], label: 'Start L' })).toBe('L');
    expect(displayStartMarkerLabel({ id: 'start-r', point: [0.7, 0.5], label: 'Start R' })).toBe('R');
    expect(displayStartMarkerLabel({ id: 'start-x', point: [0.5, 0.5], label: 'X', type: 'generic' })).toBe('X');
  });

  it('saves a topo overlay linked to a problem and upload photo with typed start markers', () => {
    const feature = { id: 'problem-1', properties: { grade: 'V2' } } as unknown as MoneyFeature;
    const points: NormalizedPoint[] = [[0.1, 0.2], [0.3, 0.4]];
    const overlay: TopoOverlay = { id: 'topo-1', upload_id: 'upload-1', problem_id: 'problem-1', color: '#F97316', width: 5, order: 0, paths: [{ id: 'path-1', points }], starts: [{ id: 'start-x', point: [0.15, 0.85], label: 'X', type: 'generic' }, { id: 'start-r', point: [0.24, 0.82], label: 'R', type: 'right' }] };

    expect(upsertTopoOverlay(feature, overlay)).toEqual({
      grade: 'V2',
      topo_overlays: [overlay],
    });
  });
});
