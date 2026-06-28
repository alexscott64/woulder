import { describe, expect, it } from 'vitest';
import { shouldFitAreaViewport } from './viewport';

describe('shouldFitAreaViewport', () => {
  it('allows the first area bounds fit before viewport interaction', () => {
    expect(shouldFitAreaViewport({ previousFocusBBoxKey: null, nextFocusBBoxKey: 'root-bounds', editing: false, hasViewportInteraction: false })).toBe(true);
  });

  it('does not refit when the area bounds key has not changed', () => {
    expect(shouldFitAreaViewport({ previousFocusBBoxKey: 'root-bounds', nextFocusBBoxKey: 'root-bounds', editing: false, hasViewportInteraction: false })).toBe(false);
  });

  it('does not refit changed area bounds after the user has zoomed or panned', () => {
    expect(shouldFitAreaViewport({ previousFocusBBoxKey: 'root-bounds', nextFocusBBoxKey: 'nested-area-bounds', editing: false, hasViewportInteraction: true })).toBe(false);
  });

  it('does not refit while area geometry is being edited', () => {
    expect(shouldFitAreaViewport({ previousFocusBBoxKey: 'root-bounds', nextFocusBBoxKey: 'edited-area-bounds', editing: true, hasViewportInteraction: false })).toBe(false);
  });
});
