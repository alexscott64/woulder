/**
 * Weather Display Helpers
 *
 * UI presentation utilities for weather components.
 * These handle colors, labels, icons - presentation concerns only.
 */

import type {
  ConditionLevel,
  RockCondition,
  FrictionQuality,
  RockTemperatureStatus,
  SendWindow,
  RockTempTransition,
} from '../../types/weather';

/**
 * Get background color class for condition level
 */
export function getConditionColor(level: ConditionLevel): string {
  switch (level) {
    case 'good':
      return 'bg-green-500';
    case 'marginal':
      return 'bg-yellow-500';
    case 'bad':
      return 'bg-red-500';
    case 'do_not_climb':
      return 'bg-red-700';
    default:
      return 'bg-gray-500';
  }
}

/**
 * Get text color class for condition level
 */
export function getConditionTextColor(level: ConditionLevel): string {
  switch (level) {
    case 'good':
      return 'text-green-600 dark:text-green-400';
    case 'marginal':
      return 'text-yellow-600 dark:text-yellow-400';
    case 'bad':
      return 'text-red-600 dark:text-red-400';
    case 'do_not_climb':
      return 'text-red-700 dark:text-red-500';
    default:
      return 'text-gray-600 dark:text-gray-400';
  }
}

/**
 * Get human-readable label for condition level
 */
export function getConditionLabel(level: ConditionLevel): string {
  switch (level) {
    case 'good':
      return 'Good';
    case 'marginal':
      return 'Fair';
    case 'bad':
      return 'Poor';
    case 'do_not_climb':
      return 'Do Not Climb';
    default:
      return 'Unknown';
  }
}

/**
 * Get badge styles for condition level (background + text + border colors)
 */
export function getConditionBadgeStyles(level: ConditionLevel): {
  bg: string;
  text: string;
  border: string;
} {
  switch (level) {
    case 'good':
      return {
        bg: 'bg-green-100 dark:bg-green-900/30',
        text: 'text-green-700 dark:text-green-300',
        border: 'border-green-300 dark:border-green-700'
      };
    case 'marginal':
      return {
        bg: 'bg-yellow-100 dark:bg-yellow-900/30',
        text: 'text-yellow-700 dark:text-yellow-300',
        border: 'border-yellow-300 dark:border-yellow-700'
      };
    case 'bad':
      return {
        bg: 'bg-red-100 dark:bg-red-900/30',
        text: 'text-red-700 dark:text-red-300',
        border: 'border-red-300 dark:border-red-700'
      };
    case 'do_not_climb':
      return {
        bg: 'bg-red-200 dark:bg-red-900/50',
        text: 'text-red-900 dark:text-red-200',
        border: 'border-red-500 dark:border-red-600'
      };
    default:
      return {
        bg: 'bg-gray-100 dark:bg-gray-700',
        text: 'text-gray-700 dark:text-gray-300',
        border: 'border-gray-300 dark:border-gray-600'
      };
  }
}

/**
 * Check if a timestamp is during daytime based on sunrise/sunset
 */
export function isDaytime(timestamp: string, sunrise: string | undefined, sunset: string | undefined): boolean {
  if (!sunrise || !sunset) {
    // Fallback to simple hour check if no sun times available
    const hour = new Date(timestamp).getHours();
    return hour >= 6 && hour < 20;
  }

  const time = new Date(timestamp).getTime();
  const sunriseTime = new Date(sunrise).getTime();
  const sunsetTime = new Date(sunset).getTime();

  return time >= sunriseTime && time < sunsetTime;
}

/**
 * Get OpenWeatherMap icon URL, correcting day/night indicator based on actual sunrise/sunset
 */
export function getWeatherIconUrl(iconCode: string, timestamp?: string, sunrise?: string, sunset?: string): string {
  // If we have timestamp and sun times, correct the day/night indicator
  if (timestamp && (sunrise || sunset)) {
    const isDay = isDaytime(timestamp, sunrise, sunset);
    const baseIcon = iconCode.substring(0, 2); // e.g., "01" from "01d" or "01n"
    const correctedIcon = baseIcon + (isDay ? 'd' : 'n');
    return `https://openweathermap.org/img/wn/${correctedIcon}@2x.png`;
  }

  return `https://openweathermap.org/img/wn/${iconCode}@2x.png`;
}

/**
 * Get color class for snow depth based on climbing conditions
 */
export function getSnowDepthColor(depth: number): string {
  if (depth === 0) return 'text-green-600'; // No snow = good
  if (depth < 3) return 'text-yellow-600'; // Light snow = marginal
  if (depth < 6) return 'text-orange-600'; // Moderate snow = challenging
  return 'text-red-600'; // Deep snow = bad
}

/**
 * Get human-readable description of snow conditions
 */
export function getSnowDescription(depth: number): string {
  if (depth === 0) return 'No snow';
  if (depth < 3) return 'Light snow cover';
  if (depth < 6) return 'Moderate snow';
  if (depth < 12) return 'Heavy snow';
  return 'Very deep snow';
}

// ===== Rock Temperature display helpers =====

/**
 * Hex colors for each rock surface-temp condition tier.
 * Used inline (style={{ backgroundColor }}) so timeline strips render
 * reliably regardless of Tailwind's JIT purge.
 */
export const ROCK_CONDITION_COLORS: Record<RockCondition, string> = {
  prime:     '#3b82f6', // blue-500
  good:      '#10b981', // emerald-500
  marginal:  '#eab308', // yellow-500
  poor:      '#f97316', // orange-500
  very_poor: '#dc2626', // red-600
  too_cold:  '#7dd3fc', // sky-300
};

export const ROCK_CONDITION_LABELS: Record<RockCondition, string> = {
  prime:     'Prime',
  good:      'Good',
  marginal:  'Marginal',
  poor:      'Poor',
  very_poor: 'Very Poor',
  too_cold:  'Too Cold',
};

export const FRICTION_QUALITY_COLORS: Record<FrictionQuality, string> = {
  excellent: '#3b82f6',
  good:      '#10b981',
  reduced:   '#eab308',
  poor:      '#dc2626',
};

export const FRICTION_QUALITY_LABELS: Record<FrictionQuality, string> = {
  excellent: 'Excellent',
  good:      'Good',
  reduced:   'Reduced',
  poor:      'Poor',
};

/**
 * Tailwind class for the diagonal-stripe overlay used to indicate
 * condensing hours over a base color. Apply alongside the base
 * background color (set via inline style).
 */
export const CONDENSATION_OVERLAY_CLASS =
  'bg-[repeating-linear-gradient(45deg,_transparent_0px,_transparent_4px,_rgba(255,255,255,0.35)_4px,_rgba(255,255,255,0.35)_6px)]';

/**
 * Format the "current condition until next transition" line.
 * Returns null if no transition is provided.
 * Example: "Prime until 11:00 AM"
 */
export function formatNextTransition(
  t: RockTempTransition | undefined,
  currentCondition: RockCondition
): string | null {
  if (!t) return null;
  const time = new Date(t.time);
  const timeStr = time.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
  const currentLabel = ROCK_CONDITION_LABELS[currentCondition];
  return `${currentLabel} until ${timeStr}`;
}

/**
 * Format a send window line for the modal list.
 * Example: "6:00 AM – 10:00 AM • 4h Prime" or "6:00 AM – 6:30 AM • 30min Good (may be damp early)"
 */
export function formatSendWindow(w: SendWindow): string {
  const start = new Date(w.start_time).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
  const end = new Date(w.end_time).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
  const dur = w.duration_h >= 1
    ? `${Math.round(w.duration_h * 10) / 10}h`
    : `${Math.round(w.duration_h * 60)}min`;
  const tier = w.condition === 'prime' ? 'Prime' : 'Good';
  const damp = w.dry_throughout ? '' : ' (may be damp early)';
  return `${start} – ${end} • ${dur} ${tier}${damp}`;
}

/**
 * Format just the start–end time range of a send window, hour granularity.
 * Example: "6–10 AM"
 */
export function formatSendWindowRange(w: SendWindow): string {
  const start = new Date(w.start_time).toLocaleTimeString(undefined, { hour: 'numeric' });
  const end = new Date(w.end_time).toLocaleTimeString(undefined, { hour: 'numeric' });
  return `${start}–${end}`;
}

// ===== Send-window Gantt helpers =====

/**
 * Day-of-week + month/day label for a backend `local_date` (YYYY-MM-DD).
 * Returns "Today" if the date matches today (in the browser's local zone).
 * Otherwise returns e.g. "Saturday" (long weekday name).
 *
 * `today` is an optional injection point for tests; defaults to `new Date()`.
 */
export function formatWeekdayLong(localDate: string, today?: Date): string {
  // Parse YYYY-MM-DD as a *local* date so the weekday matches the
  // user's intuition. Splitting avoids the UTC-midnight pitfall of
  // `new Date('2025-06-15')`.
  const [y, m, d] = localDate.split('-').map((p) => parseInt(p, 10));
  if (!y || !m || !d) return localDate;
  const date = new Date(y, m - 1, d);

  const now = today ?? new Date();
  if (
    now.getFullYear() === y &&
    now.getMonth() + 1 === m &&
    now.getDate() === d
  ) {
    return 'Today';
  }
  return date.toLocaleDateString(undefined, { weekday: 'long' });
}

/**
 * Compact time-axis label for an hour-of-day 0..24.
 * 0 / 24 → "12a", 6 → "6a", 12 → "12p", 18 → "6p".
 */
export function formatTimeAxisLabel(hour: number): string {
  const h = ((hour % 24) + 24) % 24;
  if (h === 0) return '12a';
  if (h === 12) return '12p';
  if (h < 12) return `${h}a`;
  return `${h - 12}p`;
}

/**
 * Compute where a SendWindow should render along a 24-hour day-row,
 * expressed as left/width percentages of the row width.
 *
 * The day row represents `dayLocalDate 00:00 -> 24:00` in the user's
 * browser timezone (which, for the typical case of viewing one's own
 * climbing area, matches the location's IANA timezone — the backend
 * already splits multi-day windows at local midnight, so each window
 * lies within a single local day).
 *
 * If the window does not intersect the day at all, returns
 * `{ leftPercent: 0, widthPercent: 0 }`. Callers should treat that as
 * "do not render".
 */
export function computeWindowGanttPlacement(
  window: SendWindow,
  dayLocalDate: string
): { leftPercent: number; widthPercent: number } {
  const [y, m, d] = dayLocalDate.split('-').map((p) => parseInt(p, 10));
  if (!y || !m || !d) return { leftPercent: 0, widthPercent: 0 };

  const dayStart = new Date(y, m - 1, d, 0, 0, 0, 0).getTime();
  const dayEnd = dayStart + 24 * 3600 * 1000;

  const start = new Date(window.start_time).getTime();
  const end = new Date(window.end_time).getTime();

  // Clip to the day.
  const clippedStart = Math.max(start, dayStart);
  const clippedEnd = Math.min(end, dayEnd);
  if (clippedEnd <= clippedStart) {
    return { leftPercent: 0, widthPercent: 0 };
  }

  const leftPercent = ((clippedStart - dayStart) / (24 * 3600 * 1000)) * 100;
  const widthPercent = ((clippedEnd - clippedStart) / (24 * 3600 * 1000)) * 100;
  return { leftPercent, widthPercent };
}

/**
 * Compact send-window line used in the per-day detail list.
 * Example: "11 PM → 10 AM • 11h • peak 55°F"
 */
export function formatSendWindowDetail(w: SendWindow): string {
  const start = new Date(w.start_time).toLocaleTimeString(undefined, { hour: 'numeric' });
  const end = new Date(w.end_time).toLocaleTimeString(undefined, { hour: 'numeric' });
  const dur = w.duration_h >= 1
    ? `${Math.round(w.duration_h * 10) / 10}h`
    : `${Math.round(w.duration_h * 60)}min`;
  return `${start} → ${end} • ${dur} • peak ${Math.round(w.peak_temp_f)}°F`;
}

/**
 * Compact hour label used in dense day cards.
 * Lowercase a/p, no space, no leading zero.
 * Examples: 0 → "12a", 6 → "6a", 12 → "12p", 23 → "11p".
 */
function formatCompactHour(hour: number): string {
  const h = ((hour % 24) + 24) % 24;
  if (h === 0) return '12a';
  if (h === 12) return '12p';
  if (h < 12) return `${h}a`;
  return `${h - 12}p`;
}

/**
 * Compact "start–end" hour range for a send window, used in the
 * day-card grid. Uses an en-dash and lowercase am/pm marker.
 *
 * Examples:
 *   "11p–10a" (overnight)
 *   "6a–2p"
 *   "12a–7a"
 *
 * The minute component of the underlying timestamps is intentionally
 * dropped — these cards are for at-a-glance scanning, not minute-precise
 * planning. Use `formatSendWindowDetail` when minutes matter.
 */
export function formatCompactTimeRange(startISO: string, endISO: string): string {
  const start = new Date(startISO);
  const end = new Date(endISO);
  if (isNaN(start.getTime()) || isNaN(end.getTime())) return '';
  return `${formatCompactHour(start.getHours())}–${formatCompactHour(end.getHours())}`;
}

/**
 * Compact duration label for a send window, used alongside
 * `formatCompactTimeRange`.
 *
 * Examples:
 *   "11h"   (whole hours ≥ 1)
 *   "1.5h"  (fractional hours, one decimal)
 *   "30m"   (sub-1h, rounded to whole minutes)
 */
export function formatCompactDuration(hours: number): string {
  if (!isFinite(hours) || hours <= 0) return '0m';
  if (hours < 1) {
    return `${Math.round(hours * 60)}m`;
  }
  // Whole-hour fast-path avoids "11.0h"
  const rounded = Math.round(hours * 10) / 10;
  if (rounded === Math.floor(rounded)) {
    return `${Math.floor(rounded)}h`;
  }
  return `${rounded}h`;
}

/**
 * Tagged enum returned by pickAdaptiveDisplay so the WeatherCard's
 * adaptive 3rd column can switch on .kind without re-deriving priorities.
 */
export type AdaptiveDisplayMode =
  | { kind: 'snow';    depthInches: number }
  | { kind: 'wet';     severity: 'heavy' | 'light'; clearsAt?: string }
  | { kind: 'hot';     surfaceF: number; nextTransition?: RockTempTransition }
  | { kind: 'rock';    surfaceF: number; condition: RockCondition; nextTransition?: RockTempTransition }
  | { kind: 'unknown' };

/**
 * Pick what the adaptive 3rd metric column should display.
 * Priority: snow > heavy condensation > poor/very_poor temp > light condensation > everything else.
 * Heat is treated as a stronger climbing-decision driver than light surface dampness,
 * which is why poor/very_poor outranks light condensation.
 */
export function pickAdaptiveDisplay(
  snowDepth: number | null | undefined,
  rockTemp: RockTemperatureStatus | null | undefined,
): AdaptiveDisplayMode {
  if (snowDepth != null && snowDepth > 0) {
    return { kind: 'snow', depthInches: snowDepth };
  }
  if (!rockTemp) return { kind: 'unknown' };

  const sev = rockTemp.condensation?.severity ?? 'none';
  if (sev === 'heavy') {
    return { kind: 'wet', severity: 'heavy', clearsAt: rockTemp.condensation?.clears_at };
  }
  if (rockTemp.condition === 'poor' || rockTemp.condition === 'very_poor') {
    return {
      kind: 'hot',
      surfaceF: rockTemp.estimated_surface_temp_f,
      nextTransition: rockTemp.next_transition,
    };
  }
  if (sev === 'light') {
    return { kind: 'wet', severity: 'light', clearsAt: rockTemp.condensation?.clears_at };
  }
  // prime, good, marginal, too_cold
  return {
    kind: 'rock',
    surfaceF: rockTemp.estimated_surface_temp_f,
    condition: rockTemp.condition,
    nextTransition: rockTemp.next_transition,
  };
}
