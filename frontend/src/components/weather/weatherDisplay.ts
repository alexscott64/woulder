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
