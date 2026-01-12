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

/**
 * Format days ago for display
 * Converts days to weeks/months/years for longer periods
 */
export function formatDaysAgo(days: number): string {
  if (days === 0) return 'Today';
  if (days === 1) return 'Yesterday';
  if (days < 7) return `${days} days ago`;
  if (days < 30) {
    const weeks = Math.floor(days / 7);
    return `${weeks} ${weeks === 1 ? 'week' : 'weeks'} ago`;
  }
  if (days < 365) {
    const months = Math.floor(days / 30);
    return `${months} ${months === 1 ? 'month' : 'months'} ago`;
  }
  const years = Math.floor(days / 365);
  return `${years} ${years === 1 ? 'year' : 'years'} ago`;
}
