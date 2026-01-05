import { describe, it, expect } from 'vitest';
import { formatDryTime } from '../formatters';

describe('formatDryTime', () => {
  it('should return "<1h" for values less than 1', () => {
    expect(formatDryTime(0.5)).toBe('<1h');
    expect(formatDryTime(0.1)).toBe('<1h');
    expect(formatDryTime(0)).toBe('<1h');
  });

  it('should return rounded hours for values under 72h', () => {
    expect(formatDryTime(1)).toBe('1h');
    expect(formatDryTime(2.3)).toBe('3h');
    expect(formatDryTime(5.9)).toBe('6h');
    expect(formatDryTime(12)).toBe('12h');
    expect(formatDryTime(24)).toBe('24h');
    expect(formatDryTime(48.2)).toBe('49h');
    expect(formatDryTime(71)).toBe('71h');
  });

  it('should return days for values >= 72h (3 days)', () => {
    expect(formatDryTime(72)).toBe('3d');
    expect(formatDryTime(96)).toBe('4d');
    expect(formatDryTime(168)).toBe('7d');
    expect(formatDryTime(240)).toBe('10d');
  });

  it('should round up to next day for long estimates', () => {
    expect(formatDryTime(73)).toBe('4d'); // 3.04 days -> 4 days
    expect(formatDryTime(169)).toBe('8d'); // 7.04 days -> 8 days
  });
});
