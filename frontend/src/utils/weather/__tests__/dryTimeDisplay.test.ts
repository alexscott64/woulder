import { describe, it, expect } from 'vitest';
import { formatDryTime } from '../formatters';

describe('formatDryTime', () => {
  it('should return "Unknown (snow/ice)" for 999h', () => {
    expect(formatDryTime(999)).toBe('Unknown (snow/ice)');
  });

  it('should return "Unknown (snow/ice)" for values >= 999', () => {
    expect(formatDryTime(1000)).toBe('Unknown (snow/ice)');
    expect(formatDryTime(9999)).toBe('Unknown (snow/ice)');
  });

  it('should return "<1h" for values less than 1', () => {
    expect(formatDryTime(0.5)).toBe('<1h');
    expect(formatDryTime(0.1)).toBe('<1h');
    expect(formatDryTime(0)).toBe('<1h');
  });

  it('should return rounded hours for normal values', () => {
    expect(formatDryTime(1)).toBe('1h');
    expect(formatDryTime(2.3)).toBe('3h');
    expect(formatDryTime(5.9)).toBe('6h');
    expect(formatDryTime(12)).toBe('12h');
    expect(formatDryTime(24)).toBe('24h');
    expect(formatDryTime(48.2)).toBe('49h');
  });

  it('should handle edge cases around 999', () => {
    expect(formatDryTime(998)).toBe('998h');
    expect(formatDryTime(998.9)).toBe('999h');
    expect(formatDryTime(999)).toBe('Unknown (snow/ice)');
  });
});
