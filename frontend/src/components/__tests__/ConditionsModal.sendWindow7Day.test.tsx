/**
 * Regression: the rock-temp "Send Windows" subsection (Gantt bar graph +
 * per-day DayCard "Details" grid) must visually cap at the next 7 days
 * (168h from now). The backend returns the full 16-day forecast horizon,
 * but this section is explicitly labeled "Next 7 days" and must honor it.
 *
 * What the production bug looked like (the live hard-refresh report):
 *   - The Gantt day-axis rendered ~17 day labels ("Today, Mon, Tue, …"),
 *     i.e. the full 16-day forecast.
 *   - The DayCard grid rendered 16 cards, with days 8–16 all showing
 *     "no windows" — a long, ugly empty strip.
 *
 * Why the earlier regression test missed it: the previous version only
 * asserted that out-of-range *window labels* ("peak 81°F") didn't appear.
 * It never asserted the count of day-axis labels or DayCards, so the
 * underlying iteration over `daily_forecast` (which drives both the day
 * rows and the cards) went unchecked.
 *
 * This test enforces:
 *   1. The Gantt day-axis renders at most 8 rows (today + 7).
 *   2. The DayCard grid renders at most 8 cards (today + 7).
 *   3. Out-of-range days never leak "no windows" cards into the document.
 *   4. In-range send-windows still render in both the Gantt and DayCards.
 *   5. Out-of-range send-windows are filtered everywhere (defense-in-depth).
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConditionsModal } from '../ConditionsModal';
import type {
  RockTemperatureStatus,
  SendWindow,
  DailyRockTemp,
  WeatherCondition,
} from '../../types/weather';

// Pin "now" to a deterministic moment. Noon UTC mid-month keeps the local
// date stable across the common CI/dev timezones (UTC, America/Los_Angeles,
// etc.) since +/-12h from noon UTC stays within the same calendar day in
// most zones we care about.
const NOW = new Date('2026-04-15T12:00:00Z');
const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

// The cap implemented in RockTempTabContent uses
//   cutoffMs = Date.now() + 7*24h
// and compares local_date midnight (in browser-local TZ) against it.
// With NOW pinned to noon UTC, the local midnight of day-7-from-now is
// strictly before NOW + 168h in every timezone west of UTC+12 (i.e.
// every realistic CI/dev zone), so day-7 is INCLUDED → 8 visible days
// (index 0 = today, through index 7 = today+7d). We assert exactly 8.
const EXPECTED_VISIBLE_DAYS = 8;

const todayCondition: WeatherCondition = {
  level: 'good',
  reasons: ['Dry conditions'],
};

/** Build a YYYY-MM-DD key the same way RockTempTabContent does (local time). */
function localDateKey(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

/**
 * Build a SendWindow whose start_time is `offsetMs` from NOW. Uses a unique
 * `peak_temp_f` so the rendered "peak NN°F" text uniquely identifies the
 * window in the DayCard, and the rendered title="… peak NN°F" uniquely
 * identifies the bar in the Gantt.
 */
function buildWindow(offsetMs: number, peakF: number, durationH = 3): SendWindow {
  const start = new Date(NOW.getTime() + offsetMs);
  const end = new Date(start.getTime() + durationH * HOUR);
  return {
    start_time: start.toISOString(),
    end_time: end.toISOString(),
    duration_h: durationH,
    condition: 'good',
    avg_temp_f: peakF - 2,
    peak_temp_f: peakF,
    dry_throughout: true,
  };
}

/** Build a daily_forecast covering 16 local days starting from "today". */
function buildDailyForecast(): DailyRockTemp[] {
  const days: DailyRockTemp[] = [];
  for (let i = 0; i < 16; i++) {
    const d = new Date(NOW.getTime() + i * DAY);
    days.push({
      local_date: localDateKey(d),
      peak_surface_temp_f: 60,
      min_surface_temp_f: 40,
      peak_condition: 'good',
      overall_condition: 'good',
      has_condensation: false,
      window_count: 0,
    });
  }
  return days;
}

describe('ConditionsModal - send window 7-day cutoff (regression)', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW);
    // happy-dom lacks scrollIntoView; the modal calls it after mount.
    Element.prototype.scrollIntoView = vi.fn() as unknown as typeof Element.prototype.scrollIntoView;
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it('caps Gantt day-axis, DayCard grid, and send-window rendering to the next 7 days', () => {
    // Unique peak temps so each window is individually identifiable in the
    // rendered DOM (DayCard text + Gantt bar title attribute).
    //
    // In-range (start < NOW+168h):
    const inRangeShortly = buildWindow(12 * HOUR, 71);                 // +12h
    const inRangeMid = buildWindow(3 * DAY, 72);                        // +3d
    const inRangeStraddling = buildWindow(6 * DAY + 20 * HOUR, 73, 6); // +6d20h, ends past 7d (kept whole)
    // Out-of-range (start ≥ NOW+168h):
    const outOfRangeAt7d = buildWindow(7 * DAY + HOUR, 81);             // +7d1h — HIDDEN
    const outOfRangeFar = buildWindow(10 * DAY, 82);                    // +10d  — HIDDEN

    const status: RockTemperatureStatus = {
      estimated_surface_temp_f: 58,
      air_temp_f: 55,
      temp_differential_f: 3,
      condition: 'good',
      friction_quality: 'good',
      message: 'Surface temperature is in the ideal range for friction.',
      confidence_score: 0.9,
      rock_type: 'Granite',
      send_windows: [
        inRangeShortly,
        inRangeMid,
        inRangeStraddling,
        outOfRangeAt7d,
        outOfRangeFar,
      ],
      daily_forecast: buildDailyForecast(),
    };

    render(
      <ConditionsModal
        locationName="Test Crag"
        rockTempStatus={status}
        todayCondition={todayCondition}
        initialFocus="rock-surface-temperature"
        onClose={() => {}}
      />,
    );

    // ---- 1. Gantt day-axis row count -----------------------------------
    // Each Gantt day-row's left-column weekday label carries
    // data-testid="send-window-day-axis-label". The pre-fix bug rendered
    // 16 labels (one per backend forecast day); the fix caps to 7-8.
    const dayAxisLabels = screen.getAllByTestId('send-window-day-axis-label');
    expect(dayAxisLabels.length).toBeLessThanOrEqual(EXPECTED_VISIBLE_DAYS);
    expect(dayAxisLabels.length).toBe(EXPECTED_VISIBLE_DAYS);
    // Guard against future regressions ballooning the axis back to 16+.
    expect(dayAxisLabels.length).toBeLessThan(16);
    // The leftmost label must be "Today".
    expect(dayAxisLabels[0].textContent).toContain('Today');

    // ---- 2. DayCard grid count -----------------------------------------
    const dayCards = screen.getAllByTestId('send-window-daycard');
    expect(dayCards.length).toBeLessThanOrEqual(EXPECTED_VISIBLE_DAYS);
    expect(dayCards.length).toBe(EXPECTED_VISIBLE_DAYS);
    expect(dayCards.length).toBeLessThan(16);

    // ---- 3. "no windows" leak from out-of-range days -------------------
    // Mock data: only the 3 in-range windows have send_windows; days 1, 2,
    // 4, 5, 6, 7 (relative offsets without a window) legitimately show
    // "no windows". Days 8-15 (out of range) MUST NOT be rendered at all
    // — if they were, we'd see strictly more "no windows" cards than the
    // count of in-range empty days.
    //
    // In-range days that legitimately lack a window: of the 8 visible
    // days, 3 contain windows (today, +3d, +6d-+7d straddler shows on
    // both day 6 and day 7), so at most 5 cards say "no windows".
    const noWindowsMatches = screen.queryAllByText(/no windows/i);
    expect(noWindowsMatches.length).toBeLessThanOrEqual(5);

    // ---- 4. In-range send-windows render (bar + card) ------------------
    const bars = screen.getAllByTestId('send-window-bar');
    const barTitles = bars.map((b) => b.getAttribute('title') ?? '');

    expect(barTitles.some((t) => t.includes('peak 71°F'))).toBe(true);
    expect(barTitles.some((t) => t.includes('peak 72°F'))).toBe(true);
    expect(barTitles.some((t) => t.includes('peak 73°F'))).toBe(true);

    expect(screen.getAllByText(/peak 71°F/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/peak 72°F/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/peak 73°F/).length).toBeGreaterThan(0);

    // ---- 5. Out-of-range send-windows are gone everywhere --------------
    expect(barTitles.some((t) => t.includes('peak 81°F'))).toBe(false);
    expect(barTitles.some((t) => t.includes('peak 82°F'))).toBe(false);
    expect(screen.queryByText(/peak 81°F/)).toBeNull();
    expect(screen.queryByText(/peak 82°F/)).toBeNull();
  });
});
