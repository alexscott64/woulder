import { describe, it, expect } from 'vitest';
import {
  getConditionColor,
  getConditionTextColor,
  getConditionLabel,
  getConditionBadgeStyles,
  getWeatherIconUrl,
  getSnowDepthColor,
  getSnowDescription,
  ROCK_CONDITION_COLORS,
  ROCK_CONDITION_LABELS,
  FRICTION_QUALITY_COLORS,
  FRICTION_QUALITY_LABELS,
  formatNextTransition,
  formatSendWindow,
  formatSendWindowRange,
  formatWeekdayLong,
  formatTimeAxisLabel,
  computeWindowGanttPlacement,
  formatSendWindowDetail,
  formatCompactTimeRange,
  formatCompactDuration,
  pickAdaptiveDisplay,
} from '../weatherDisplay';
import type { ConditionLevel, RockTemperatureStatus, SendWindow } from '../../../types/weather';

describe('weatherDisplay', () => {
  describe('getConditionColor', () => {
    it('should return green background for good conditions', () => {
      expect(getConditionColor('good')).toBe('bg-green-500');
    });

    it('should return yellow background for marginal conditions', () => {
      expect(getConditionColor('marginal')).toBe('bg-yellow-500');
    });

    it('should return red background for bad conditions', () => {
      expect(getConditionColor('bad')).toBe('bg-red-500');
    });

    it('should return gray background for unknown conditions', () => {
      expect(getConditionColor('unknown' as ConditionLevel)).toBe('bg-gray-500');
    });
  });

  describe('getConditionTextColor', () => {
    it('should return green text colors for good conditions', () => {
      expect(getConditionTextColor('good')).toBe('text-green-600 dark:text-green-400');
    });

    it('should return yellow text colors for marginal conditions', () => {
      expect(getConditionTextColor('marginal')).toBe('text-yellow-600 dark:text-yellow-400');
    });

    it('should return red text colors for bad conditions', () => {
      expect(getConditionTextColor('bad')).toBe('text-red-600 dark:text-red-400');
    });

    it('should return gray text colors for unknown conditions', () => {
      expect(getConditionTextColor('unknown' as ConditionLevel)).toBe('text-gray-600 dark:text-gray-400');
    });
  });

  describe('getConditionLabel', () => {
    it('should return "Good" for good conditions', () => {
      expect(getConditionLabel('good')).toBe('Good');
    });

    it('should return "Fair" for marginal conditions', () => {
      expect(getConditionLabel('marginal')).toBe('Fair');
    });

    it('should return "Poor" for bad conditions', () => {
      expect(getConditionLabel('bad')).toBe('Poor');
    });

    it('should return "Unknown" for unknown conditions', () => {
      expect(getConditionLabel('unknown' as ConditionLevel)).toBe('Unknown');
    });
  });

  describe('getConditionBadgeStyles', () => {
    it('should return green badge styles for good conditions', () => {
      const styles = getConditionBadgeStyles('good');
      expect(styles.bg).toBe('bg-green-100 dark:bg-green-900/30');
      expect(styles.text).toBe('text-green-700 dark:text-green-300');
      expect(styles.border).toBe('border-green-300 dark:border-green-700');
    });

    it('should return yellow badge styles for marginal conditions', () => {
      const styles = getConditionBadgeStyles('marginal');
      expect(styles.bg).toBe('bg-yellow-100 dark:bg-yellow-900/30');
      expect(styles.text).toBe('text-yellow-700 dark:text-yellow-300');
      expect(styles.border).toBe('border-yellow-300 dark:border-yellow-700');
    });

    it('should return red badge styles for bad conditions', () => {
      const styles = getConditionBadgeStyles('bad');
      expect(styles.bg).toBe('bg-red-100 dark:bg-red-900/30');
      expect(styles.text).toBe('text-red-700 dark:text-red-300');
      expect(styles.border).toBe('border-red-300 dark:border-red-700');
    });

    it('should return gray badge styles for unknown conditions', () => {
      const styles = getConditionBadgeStyles('unknown' as ConditionLevel);
      expect(styles.bg).toBe('bg-gray-100 dark:bg-gray-700');
      expect(styles.text).toBe('text-gray-700 dark:text-gray-300');
      expect(styles.border).toBe('border-gray-300 dark:border-gray-600');
    });

    it('should return all three style properties', () => {
      const styles = getConditionBadgeStyles('good');
      expect(styles).toHaveProperty('bg');
      expect(styles).toHaveProperty('text');
      expect(styles).toHaveProperty('border');
      expect(Object.keys(styles)).toHaveLength(3);
    });
  });

  describe('getWeatherIconUrl', () => {
    it('should generate correct OpenWeatherMap icon URL', () => {
      const iconCode = '01d';
      const expectedUrl = 'https://openweathermap.org/img/wn/01d@2x.png';
      expect(getWeatherIconUrl(iconCode)).toBe(expectedUrl);
    });

    it('should handle different icon codes', () => {
      expect(getWeatherIconUrl('10n')).toBe('https://openweathermap.org/img/wn/10n@2x.png');
      expect(getWeatherIconUrl('13d')).toBe('https://openweathermap.org/img/wn/13d@2x.png');
      expect(getWeatherIconUrl('50d')).toBe('https://openweathermap.org/img/wn/50d@2x.png');
    });

    it('should always use @2x resolution', () => {
      const url = getWeatherIconUrl('01d');
      expect(url).toContain('@2x.png');
    });
  });

  describe('getSnowDepthColor', () => {
    it('should return green for no snow', () => {
      expect(getSnowDepthColor(0)).toBe('text-green-600');
    });

    it('should return yellow for light snow (< 3 inches)', () => {
      expect(getSnowDepthColor(1)).toBe('text-yellow-600');
      expect(getSnowDepthColor(2.9)).toBe('text-yellow-600');
    });

    it('should return orange for moderate snow (3-6 inches)', () => {
      expect(getSnowDepthColor(3)).toBe('text-orange-600');
      expect(getSnowDepthColor(5)).toBe('text-orange-600');
      expect(getSnowDepthColor(5.9)).toBe('text-orange-600');
    });

    it('should return red for deep snow (>= 6 inches)', () => {
      expect(getSnowDepthColor(6)).toBe('text-red-600');
      expect(getSnowDepthColor(12)).toBe('text-red-600');
      expect(getSnowDepthColor(24)).toBe('text-red-600');
    });
  });

  describe('getSnowDescription', () => {
    it('should return "No snow" for 0 inches', () => {
      expect(getSnowDescription(0)).toBe('No snow');
    });

    it('should return "Light snow cover" for < 3 inches', () => {
      expect(getSnowDescription(1)).toBe('Light snow cover');
      expect(getSnowDescription(2.5)).toBe('Light snow cover');
    });

    it('should return "Moderate snow" for 3-6 inches', () => {
      expect(getSnowDescription(3)).toBe('Moderate snow');
      expect(getSnowDescription(5)).toBe('Moderate snow');
      expect(getSnowDescription(5.9)).toBe('Moderate snow');
    });

    it('should return "Heavy snow" for 6-12 inches', () => {
      expect(getSnowDescription(6)).toBe('Heavy snow');
      expect(getSnowDescription(10)).toBe('Heavy snow');
      expect(getSnowDescription(11.9)).toBe('Heavy snow');
    });

    it('should return "Very deep snow" for >= 12 inches', () => {
      expect(getSnowDescription(12)).toBe('Very deep snow');
      expect(getSnowDescription(24)).toBe('Very deep snow');
      expect(getSnowDescription(48)).toBe('Very deep snow');
    });
  });

  // ===== Rock temperature helpers =====

  describe('ROCK_CONDITION_COLORS', () => {
    it('has all 6 condition keys with valid hex strings', () => {
      const keys = ['prime', 'good', 'marginal', 'poor', 'very_poor', 'too_cold'] as const;
      expect(Object.keys(ROCK_CONDITION_COLORS).sort()).toEqual([...keys].sort());
      for (const k of keys) {
        expect(ROCK_CONDITION_COLORS[k]).toMatch(/^#[0-9a-f]{6}$/i);
      }
    });

    it('has matching labels for each condition', () => {
      const keys = ['prime', 'good', 'marginal', 'poor', 'very_poor', 'too_cold'] as const;
      for (const k of keys) {
        expect(typeof ROCK_CONDITION_LABELS[k]).toBe('string');
        expect(ROCK_CONDITION_LABELS[k].length).toBeGreaterThan(0);
      }
    });
  });

  describe('FRICTION_QUALITY_COLORS', () => {
    it('has all 4 friction-quality keys with valid hex strings', () => {
      const keys = ['excellent', 'good', 'reduced', 'poor'] as const;
      expect(Object.keys(FRICTION_QUALITY_COLORS).sort()).toEqual([...keys].sort());
      for (const k of keys) {
        expect(FRICTION_QUALITY_COLORS[k]).toMatch(/^#[0-9a-f]{6}$/i);
        expect(typeof FRICTION_QUALITY_LABELS[k]).toBe('string');
      }
    });
  });

  describe('formatNextTransition', () => {
    it('returns null when no transition is provided', () => {
      expect(formatNextTransition(undefined, 'prime')).toBeNull();
    });

    it('returns a string containing the current label and "until"', () => {
      const result = formatNextTransition(
        { time: '2025-06-15T18:00:00Z', to_condition: 'marginal' },
        'prime'
      );
      expect(result).not.toBeNull();
      expect(result!).toContain('Prime');
      expect(result!).toContain('until');
      // Time formatting is locale-dependent, but should contain a digit
      expect(result!).toMatch(/\d/);
    });
  });

  describe('formatSendWindow', () => {
    const baseWindow: SendWindow = {
      start_time: '2025-06-15T13:00:00Z',
      end_time: '2025-06-15T14:30:00Z',
      duration_h: 1.5,
      condition: 'prime',
      avg_temp_f: 55,
      peak_temp_f: 60,
      dry_throughout: true,
    };

    it('omits "may be damp" suffix when dry_throughout is true', () => {
      const out = formatSendWindow(baseWindow);
      expect(out).not.toContain('may be damp');
      expect(out).toContain('Prime');
    });

    it('includes "may be damp" suffix when dry_throughout is false', () => {
      const out = formatSendWindow({ ...baseWindow, dry_throughout: false });
      expect(out).toContain('may be damp');
    });

    it('formats durations >= 1h as e.g. "1.5h"', () => {
      expect(formatSendWindow(baseWindow)).toContain('1.5h');
    });

    it('formats durations < 1h as minutes (e.g. 0.5h -> "30min")', () => {
      const short: SendWindow = { ...baseWindow, duration_h: 0.5 };
      expect(formatSendWindow(short)).toContain('30min');
      expect(formatSendWindow(short)).not.toContain('0.5h');
    });

    it('uses "Good" tier label when condition is good', () => {
      expect(formatSendWindow({ ...baseWindow, condition: 'good' })).toContain('Good');
    });
  });

  describe('formatSendWindowDetail', () => {
    it('includes start, end, duration and peak temp', () => {
      const w: SendWindow = {
        start_time: '2025-06-15T13:00:00Z',
        end_time: '2025-06-15T17:00:00Z',
        duration_h: 4,
        condition: 'prime',
        avg_temp_f: 55,
        peak_temp_f: 60.4,
        dry_throughout: true,
      };
      const out = formatSendWindowDetail(w);
      expect(out).toContain('→');
      expect(out).toContain('4h');
      // Rounded peak.
      expect(out).toContain('60°F');
    });

    it('formats short durations as minutes', () => {
      const w: SendWindow = {
        start_time: '2025-06-15T13:00:00Z',
        end_time: '2025-06-15T13:30:00Z',
        duration_h: 0.5,
        condition: 'good',
        avg_temp_f: 55,
        peak_temp_f: 58,
        dry_throughout: true,
      };
      expect(formatSendWindowDetail(w)).toContain('30min');
    });
  });

  describe('formatSendWindowRange', () => {
    it('returns a hour-granularity start–end range', () => {
      const w: SendWindow = {
        start_time: '2025-06-15T13:00:00Z',
        end_time: '2025-06-15T17:00:00Z',
        duration_h: 4,
        condition: 'prime',
        avg_temp_f: 60,
        peak_temp_f: 65,
        dry_throughout: true,
      };
      const out = formatSendWindowRange(w);
      // Locale-dependent, but should contain a dash and digits and not seconds/minutes detail
      expect(out).toMatch(/\d+.*[–-].*\d+/);
      expect(out).not.toContain(':00');
    });
  });

  describe('pickAdaptiveDisplay', () => {
    const mkRock = (overrides: Partial<RockTemperatureStatus> = {}): RockTemperatureStatus => ({
      estimated_surface_temp_f: 70,
      air_temp_f: 60,
      temp_differential_f: 10,
      condition: 'prime',
      friction_quality: 'excellent',
      message: '',
      confidence_score: 90,
      rock_type: 'Granite',
      ...overrides,
    });

    it('returns snow when snow depth > 0 (regardless of rock_temp)', () => {
      const result = pickAdaptiveDisplay(5, mkRock());
      expect(result).toEqual({ kind: 'snow', depthInches: 5 });
    });

    it('returns wet/heavy when severity is heavy', () => {
      const result = pickAdaptiveDisplay(0, mkRock({
        condition: 'good',
        condensation: {
          active: true, dewpoint_f: 65, surface_vs_dewpoint: -2,
          severity: 'heavy', reason: 'condensing',
        },
      }));
      expect(result.kind).toBe('wet');
      if (result.kind === 'wet') expect(result.severity).toBe('heavy');
    });

    it('returns hot when condition is very_poor and severity is none', () => {
      const result = pickAdaptiveDisplay(0, mkRock({
        condition: 'very_poor',
        estimated_surface_temp_f: 130,
      }));
      expect(result.kind).toBe('hot');
      if (result.kind === 'hot') expect(result.surfaceF).toBe(130);
    });

    it('returns wet/light when severity is light and condition is good', () => {
      const result = pickAdaptiveDisplay(0, mkRock({
        condition: 'good',
        condensation: {
          active: false, dewpoint_f: 55, surface_vs_dewpoint: 0.5,
          severity: 'light', reason: 'near dewpoint',
        },
      }));
      expect(result.kind).toBe('wet');
      if (result.kind === 'wet') expect(result.severity).toBe('light');
    });

    it('returns rock with condition=prime when severity is none and condition is prime', () => {
      const result = pickAdaptiveDisplay(0, mkRock({ condition: 'prime' }));
      expect(result.kind).toBe('rock');
      if (result.kind === 'rock') expect(result.condition).toBe('prime');
    });

    it('returns unknown when no snow and no rock_temp', () => {
      expect(pickAdaptiveDisplay(null, null)).toEqual({ kind: 'unknown' });
      expect(pickAdaptiveDisplay(undefined, undefined)).toEqual({ kind: 'unknown' });
    });

    it('treats poor temp as priority over light condensation (heat drives the decision)', () => {
      const result = pickAdaptiveDisplay(0, mkRock({
        condition: 'poor',
        condensation: {
          active: false, dewpoint_f: 60, surface_vs_dewpoint: 1,
          severity: 'light', reason: 'near dewpoint',
        },
      }));
      expect(result.kind).toBe('hot');
    });

    it('treats snow=0 as no snow (falls through to rock_temp logic)', () => {
      const result = pickAdaptiveDisplay(0, mkRock({ condition: 'good' }));
      expect(result.kind).toBe('rock');
    });
  
    // ===== Send-window Gantt helpers =====
  
    describe('formatWeekdayLong', () => {
      it('returns "Today" when localDate matches the supplied "today"', () => {
        const today = new Date(2025, 5, 14); // June 14 2025 local
        expect(formatWeekdayLong('2025-06-14', today)).toBe('Today');
      });
  
      it('returns a long weekday name for a non-today date', () => {
        // 2025-06-15 is a Sunday.
        const result = formatWeekdayLong('2025-06-15', new Date(2025, 5, 14));
        expect(result).toBe('Sunday');
      });
  
      it('returns the input string when given a malformed date', () => {
        expect(formatWeekdayLong('not-a-date')).toBe('not-a-date');
      });
  
      it('parses YYYY-MM-DD as a local date (not UTC midnight)', () => {
        // 2025-01-01 is a Wednesday in any timezone when parsed as local.
        // (new Date('2025-01-01') would be UTC midnight, which is Dec 31
        // in the Americas — the helper must avoid that pitfall.)
        const today = new Date(2024, 11, 30); // Dec 30 2024
        const result = formatWeekdayLong('2025-01-01', today);
        expect(result).toBe('Wednesday');
      });
    });
  
    describe('formatTimeAxisLabel', () => {
      it('formats 0 as "12a"', () => {
        expect(formatTimeAxisLabel(0)).toBe('12a');
      });
      it('formats 24 as "12a" (wraps)', () => {
        expect(formatTimeAxisLabel(24)).toBe('12a');
      });
      it('formats 6 as "6a"', () => {
        expect(formatTimeAxisLabel(6)).toBe('6a');
      });
      it('formats 12 as "12p"', () => {
        expect(formatTimeAxisLabel(12)).toBe('12p');
      });
      it('formats 18 as "6p"', () => {
        expect(formatTimeAxisLabel(18)).toBe('6p');
      });
      it('formats 11 as "11a" and 13 as "1p"', () => {
        expect(formatTimeAxisLabel(11)).toBe('11a');
        expect(formatTimeAxisLabel(13)).toBe('1p');
      });
    });
  
    describe('computeWindowGanttPlacement', () => {
      // Build a window that's local-time aware by computing ISO strings
      // from a Date constructed in the test runner's local zone — that
      // way the test works regardless of the runner's TZ.
      const mkLocalISO = (y: number, m: number, d: number, h: number) =>
        new Date(y, m - 1, d, h, 0, 0, 0).toISOString();
  
      const baseWin = (start: string, end: string, hours: number): SendWindow => ({
        start_time: start,
        end_time: end,
        duration_h: hours,
        condition: 'prime',
        avg_temp_f: 55,
        peak_temp_f: 60,
        dry_throughout: true,
      });
  
      it('places a 6h–10h window at 25% left and 16.67% wide', () => {
        const w = baseWin(mkLocalISO(2025, 6, 15, 6), mkLocalISO(2025, 6, 15, 10), 4);
        const p = computeWindowGanttPlacement(w, '2025-06-15');
        expect(p.leftPercent).toBeCloseTo(25, 4);
        expect(p.widthPercent).toBeCloseTo((4 / 24) * 100, 4);
      });
  
      it('places a full-day 0h–24h window at 0% left and 100% wide', () => {
        const w = baseWin(mkLocalISO(2025, 6, 15, 0), mkLocalISO(2025, 6, 16, 0), 24);
        const p = computeWindowGanttPlacement(w, '2025-06-15');
        expect(p.leftPercent).toBeCloseTo(0, 4);
        expect(p.widthPercent).toBeCloseTo(100, 4);
      });
  
      it('clips a window that overshoots the day end', () => {
        // 18:00 -> next day 06:00 = 12h, but only 6h falls on 2025-06-15.
        const w = baseWin(mkLocalISO(2025, 6, 15, 18), mkLocalISO(2025, 6, 16, 6), 12);
        const p = computeWindowGanttPlacement(w, '2025-06-15');
        expect(p.leftPercent).toBeCloseTo(75, 4);
        expect(p.widthPercent).toBeCloseTo(25, 4);
      });
  
      it('returns zero width when the window does not intersect the day', () => {
        const w = baseWin(mkLocalISO(2025, 6, 16, 6), mkLocalISO(2025, 6, 16, 10), 4);
        const p = computeWindowGanttPlacement(w, '2025-06-15');
        expect(p.widthPercent).toBe(0);
        expect(p.leftPercent).toBe(0);
      });
  
      it('returns zero placement for a malformed date', () => {
        const w = baseWin(mkLocalISO(2025, 6, 15, 6), mkLocalISO(2025, 6, 15, 10), 4);
        expect(computeWindowGanttPlacement(w, 'bad')).toEqual({
          leftPercent: 0,
          widthPercent: 0,
        });
      });
    
      describe('formatCompactTimeRange', () => {
        // Build a local-zone ISO so getHours() returns the intended hour
        // regardless of the test runner's timezone.
        const local = (h: number) => {
          const d = new Date(2025, 5, 15, h, 0, 0, 0); // June 15 2025
          return d.toISOString();
        };
    
        it('formats a morning range with lowercase a/p and en-dash', () => {
          // Built off a local Date, so reading via getHours roundtrips.
          const start = new Date(2025, 5, 15, 6, 0).toISOString();
          const end = new Date(2025, 5, 15, 14, 0).toISOString();
          expect(formatCompactTimeRange(start, end)).toBe('6a–2p');
        });
    
        it('special-cases midnight as 12a', () => {
          expect(formatCompactTimeRange(local(0), local(7))).toBe('12a–7a');
        });
    
        it('special-cases noon as 12p', () => {
          expect(formatCompactTimeRange(local(12), local(15))).toBe('12p–3p');
        });
    
        it('handles overnight windows (end hour numerically before start)', () => {
          // 11pm Sunday -> 10am Monday is fine; we render endpoints regardless
          // of date, since the day-card row already scopes the day visually.
          const start = new Date(2025, 5, 15, 23, 0).toISOString();
          const end = new Date(2025, 5, 16, 10, 0).toISOString();
          expect(formatCompactTimeRange(start, end)).toBe('11p–10a');
        });
    
        it('returns empty string for invalid input', () => {
          expect(formatCompactTimeRange('garbage', 'also bad')).toBe('');
        });
      });
    
      describe('formatCompactDuration', () => {
        it('returns whole-hour suffix for integer hours', () => {
          expect(formatCompactDuration(11)).toBe('11h');
          expect(formatCompactDuration(1)).toBe('1h');
        });
    
        it('returns minutes for sub-1h durations', () => {
          expect(formatCompactDuration(0.5)).toBe('30m');
          expect(formatCompactDuration(0.25)).toBe('15m');
        });
    
        it('keeps one decimal for fractional hours', () => {
          expect(formatCompactDuration(1.5)).toBe('1.5h');
          expect(formatCompactDuration(2.7)).toBe('2.7h');
        });
    
        it('handles edge cases gracefully', () => {
          expect(formatCompactDuration(0)).toBe('0m');
          expect(formatCompactDuration(-1)).toBe('0m');
          expect(formatCompactDuration(NaN)).toBe('0m');
        });
      });
    });
  });
});
