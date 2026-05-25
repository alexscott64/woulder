/**
 * Per-day "Send Windows" section in the condition-details modal.
 *
 * When a user clicks a day's condition button (Good/Fair/Poor) in the
 * 6-day forecast, the `ConditionDetailsModal` opens with that day's
 * details. This test locks in:
 *
 *   1. The new section renders exactly ONE Gantt row (`send-window-bar-row`)
 *      — single-day strip, not the 7-day grid.
 *   2. Exactly ONE DayCard renders (`send-window-daycard`) wrapped in the
 *      single-day container (`send-window-single-day`).
 *   3. The two in-range send-windows for the target day appear (by their
 *      unique "peak NN°F" text in the DayCard).
 *   4. Send-windows from OTHER days do NOT appear in the modal — only the
 *      target day's windows are shown.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConditionDetailsModal } from '../ConditionDetailsModal';
import type {
  RockTemperatureStatus,
  SendWindow,
  DailyRockTemp,
} from '../../types/weather';

// Pin "now" to noon UTC mid-month so the browser-local date computation
// inside SendWindowDayView is stable across the typical CI/dev zones.
const NOW = new Date('2026-04-15T12:00:00Z');
const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

/** YYYY-MM-DD in browser-local time (matches SendWindowDayView). */
function localDateKey(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

/**
 * Build a SendWindow at `offsetMs` from NOW with a unique peak temp so
 * its rendered "peak NN°F" text is individually identifiable in the DOM.
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

/** Build a 16-day daily_forecast covering the full backend horizon. */
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

describe('ConditionDetailsModal - per-day Send Windows section', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(NOW);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it('renders a single-day Gantt strip + DayCard scoped to the target date', () => {
    // Pick "today + 2d" as our target day so it's clearly mid-week.
    const targetDate = localDateKey(new Date(NOW.getTime() + 2 * DAY));

    // Two windows ON the target day (different start hours).
    const dayWindowA = buildWindow(2 * DAY + 8 * HOUR, 71);    // +2d 8a, peak 71
    const dayWindowB = buildWindow(2 * DAY + 14 * HOUR, 72);   // +2d 2p, peak 72
    // Windows on OTHER days that must NOT appear in the modal.
    const otherDay1 = buildWindow(0 * DAY + 9 * HOUR, 81);     // today,  peak 81
    const otherDay2 = buildWindow(5 * DAY + 10 * HOUR, 82);    // +5d,    peak 82

    const status: RockTemperatureStatus = {
      estimated_surface_temp_f: 58,
      air_temp_f: 55,
      temp_differential_f: 3,
      condition: 'good',
      friction_quality: 'good',
      message: 'Surface temperature is in the ideal range for friction.',
      confidence_score: 0.9,
      rock_type: 'Granite',
      send_windows: [dayWindowA, dayWindowB, otherDay1, otherDay2],
      daily_forecast: buildDailyForecast(),
    };

    render(
      <ConditionDetailsModal
        locationName="Wed Forecast"
        conditionLevel="good"
        conditionLabel="Good"
        reasons={['Dry conditions', 'Mild winds']}
        rockTempStatus={status}
        targetDate={targetDate}
        onClose={() => {}}
      />,
    );

    // ---- Sanity: header + Contributing Factors still render ------------
    expect(screen.getByText(/Climbing Conditions/i)).toBeTruthy();
    expect(screen.getByText(/Contributing Factors/i)).toBeTruthy();
    expect(screen.getByText(/Send Windows/i)).toBeTruthy();

    // ---- 1. Single-day container ---------------------------------------
    const singleDayContainers = screen.getAllByTestId('send-window-single-day');
    expect(singleDayContainers.length).toBe(1);

    // ---- 2. Exactly one DayCard ----------------------------------------
    const dayCards = screen.getAllByTestId('send-window-daycard');
    expect(dayCards.length).toBe(1);

    // ---- 3. Exactly one Gantt row --------------------------------------
    const ganttRows = screen.getAllByTestId('send-window-bar-row');
    expect(ganttRows.length).toBe(1);

    // ---- 4. The two in-range windows render in the DayCard -------------
    // DayCard renders "peak NN°F" text once per window.
    expect(screen.getAllByText(/peak 71°F/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/peak 72°F/).length).toBeGreaterThan(0);

    // The Gantt row should also carry two `send-window-bar` overlays for
    // the target day's two windows (and nothing from other days).
    const bars = screen.getAllByTestId('send-window-bar');
    expect(bars.length).toBe(2);
    const barTitles = bars.map((b) => b.getAttribute('title') ?? '');
    expect(barTitles.some((t) => t.includes('peak 71°F'))).toBe(true);
    expect(barTitles.some((t) => t.includes('peak 72°F'))).toBe(true);

    // ---- 5. Other-day windows are NOT rendered -------------------------
    expect(screen.queryByText(/peak 81°F/)).toBeNull();
    expect(screen.queryByText(/peak 82°F/)).toBeNull();
    expect(barTitles.some((t) => t.includes('peak 81°F'))).toBe(false);
    expect(barTitles.some((t) => t.includes('peak 82°F'))).toBe(false);
  });

  it('omits the Send Windows section when no rockTempStatus is provided', () => {
    render(
      <ConditionDetailsModal
        locationName="Wed Forecast"
        conditionLevel="good"
        conditionLabel="Good"
        reasons={['Dry conditions']}
        onClose={() => {}}
      />,
    );

    // Header + factors render…
    expect(screen.getByText(/Climbing Conditions/i)).toBeTruthy();
    expect(screen.getByText(/Contributing Factors/i)).toBeTruthy();
    // …but the Send Windows section is absent.
    expect(screen.queryByText(/Send Windows/i)).toBeNull();
    expect(screen.queryByTestId('send-window-single-day')).toBeNull();
    expect(screen.queryByTestId('send-window-daycard')).toBeNull();
    expect(screen.queryByTestId('send-window-bar-row')).toBeNull();
  });
});
