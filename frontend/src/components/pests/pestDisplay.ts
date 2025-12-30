/**
 * Pest Display Helpers
 *
 * UI presentation utilities for pest components.
 * These handle colors, labels - presentation concerns only.
 */

import { PestLevel } from '../../utils/pests/calculations/pests';

/**
 * Get text color for pest level
 */
export function getPestLevelColor(level: PestLevel): string {
  switch (level) {
    case 'extreme': return 'text-red-600';
    case 'very_high': return 'text-orange-500';
    case 'high': return 'text-yellow-600';
    case 'moderate': return 'text-yellow-500';
    case 'low': return 'text-green-600';
  }
}

/**
 * Get background color for pest level
 */
export function getPestLevelBgColor(level: PestLevel): string {
  switch (level) {
    case 'extreme': return 'bg-red-500';
    case 'very_high': return 'bg-orange-500';
    case 'high': return 'bg-yellow-500';
    case 'moderate': return 'bg-yellow-400';
    case 'low': return 'bg-green-500';
  }
}

/**
 * Get display text for pest level
 */
export function getPestLevelText(level: PestLevel): string {
  switch (level) {
    case 'extreme': return 'Extreme';
    case 'very_high': return 'Very High';
    case 'high': return 'High';
    case 'moderate': return 'Moderate';
    case 'low': return 'Low';
  }
}
