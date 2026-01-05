/**
 * Weather data formatting utilities
 */

/**
 * Format dry time for display
 * Converts hours to days for long estimates
 */
export function formatDryTime(hoursUntilDry: number): string {
  if (hoursUntilDry < 1) {
    return '<1h';
  }

  // Show in hours for estimates under 3 days
  if (hoursUntilDry < 72) {
    return `${Math.ceil(hoursUntilDry)}h`;
  }

  // Show in days for longer estimates
  const days = Math.ceil(hoursUntilDry / 24);
  return `${days}d`;
}
