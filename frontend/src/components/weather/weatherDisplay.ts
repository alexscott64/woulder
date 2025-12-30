/**
 * Weather Display Helpers
 *
 * UI presentation utilities for weather components.
 * These handle colors, labels, icons - presentation concerns only.
 */

import { ConditionLevel } from '../../types/weather';

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
    default:
      return {
        bg: 'bg-gray-100 dark:bg-gray-700',
        text: 'text-gray-700 dark:text-gray-300',
        border: 'border-gray-300 dark:border-gray-600'
      };
  }
}

/**
 * Get OpenWeatherMap icon URL
 */
export function getWeatherIconUrl(iconCode: string): string {
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
