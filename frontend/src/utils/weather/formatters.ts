/**
 * Weather data formatting utilities
 */

/**
 * Format dry time for display
 * Handles the special case of 999h (unknown dry time for snow/ice)
 */
export function formatDryTime(hoursUntilDry: number): string {
  if (hoursUntilDry >= 999) {
    return 'Unknown (snow/ice)';
  }

  if (hoursUntilDry < 1) {
    return '<1h';
  }

  return `${Math.ceil(hoursUntilDry)}h`;
}
