import { WeatherData } from '../../../types/weather';

/**
 * Pure pest calculation utilities
 * Domain-specific math for pest/insect activity - NO business logic
 */

export type PestLevel = 'low' | 'moderate' | 'high' | 'very_high' | 'extreme';

/**
 * Seasonal multiplier for mosquito activity
 * Based on typical northern hemisphere patterns (WA state latitude ~47-49°N)
 */
export function getSeasonalMosquitoFactor(month: number): number {
  const factors: Record<number, number> = {
    1: 0,     // January - dormant
    2: 0,     // February - dormant
    3: 0.1,   // March - emerging
    4: 0.3,   // April - increasing
    5: 0.6,   // May - active
    6: 0.9,   // June - high activity
    7: 1.0,   // July - peak
    8: 1.0,   // August - peak
    9: 0.7,   // September - declining
    10: 0.3,  // October - late season
    11: 0.1,  // November - dying off
    12: 0,    // December - dormant
  };
  return factors[month] || 0;
}

/**
 * Seasonal multiplier for general pest activity
 */
export function getSeasonalPestFactor(month: number): number {
  const factors: Record<number, number> = {
    1: 0.1,   // January - minimal
    2: 0.1,   // February - minimal
    3: 0.3,   // March - emerging
    4: 0.5,   // April - increasing
    5: 0.8,   // May - active
    6: 1.0,   // June - high
    7: 1.0,   // July - peak
    8: 1.0,   // August - peak
    9: 0.8,   // September - still high
    10: 0.5,  // October - declining
    11: 0.2,  // November - low
    12: 0.1,  // December - minimal
  };
  return factors[month] || 0;
}

/**
 * Convert score (0-100) to pest level category
 */
export function scoreToLevel(score: number): PestLevel {
  if (score >= 80) return 'extreme';
  if (score >= 60) return 'very_high';
  if (score >= 40) return 'high';
  if (score >= 20) return 'moderate';
  return 'low';
}

/**
 * Calculate total rainfall in last N days from historical data
 */
export function calculateRecentRainfall(
  historical: WeatherData[],
  daysAgo: number
): number {
  const cutoffTime = Date.now() - daysAgo * 24 * 60 * 60 * 1000;
  const recentHistorical = historical.filter(
    h => new Date(h.timestamp).getTime() >= cutoffTime
  );
  return recentHistorical.reduce((sum, h) => sum + h.precipitation, 0);
}

/**
 * Get average value from weather data array
 */
export function getAverageValue(
  data: WeatherData[],
  key: 'temperature' | 'humidity' | 'wind_speed'
): number {
  if (data.length === 0) return 0;
  return data.reduce((sum, d) => sum + d[key], 0) / data.length;
}

/**
 * Get maximum value from weather data array
 */
export function getMaxValue(
  data: WeatherData[],
  key: 'temperature' | 'humidity' | 'wind_speed'
): number {
  if (data.length === 0) return 0;
  return Math.max(...data.map(d => d[key]));
}

/**
 * Find worst pest level between two levels
 */
export function getWorstLevel(level1: PestLevel, level2: PestLevel): PestLevel {
  const levelOrder: PestLevel[] = ['low', 'moderate', 'high', 'very_high', 'extreme'];
  const index1 = levelOrder.indexOf(level1);
  const index2 = levelOrder.indexOf(level2);
  return levelOrder[Math.max(index1, index2)];
}
