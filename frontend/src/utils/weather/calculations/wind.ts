/**
 * Pure wind calculation utilities
 * Domain-specific math for wind data - NO business logic
 */

// Wind speed thresholds (mph)
export const CALM_MAX = 5;
export const LIGHT_MAX = 12;
export const MODERATE_MAX = 20;
export const STRONG_MAX = 30;

/**
 * Get wind category based on speed
 */
export function getCategory(windSpeedMph: number): 'calm' | 'light' | 'moderate' | 'strong' | 'dangerous' {
  if (windSpeedMph <= CALM_MAX) return 'calm';
  if (windSpeedMph <= LIGHT_MAX) return 'light';
  if (windSpeedMph <= MODERATE_MAX) return 'moderate';
  if (windSpeedMph <= STRONG_MAX) return 'strong';
  return 'dangerous';
}

/**
 * Check if wind speed aids evaporation/drying
 * Ideal drying wind: 5-20 mph
 */
export function aidsDrying(windSpeedMph: number): boolean {
  return windSpeedMph > CALM_MAX && windSpeedMph < MODERATE_MAX;
}

/**
 * Convert wind direction degrees to compass bearing (N, NE, E, etc.)
 */
export function degreesToCompass(degrees: number): string {
  const directions = ['N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW'];
  const index = Math.round(degrees / 45) % 8;
  return directions[index];
}

/**
 * Convert wind direction degrees to cardinal direction name
 */
export function degreesToCardinal(degrees: number): string {
  const cardinals = [
    'North', 'Northeast', 'East', 'Southeast',
    'South', 'Southwest', 'West', 'Northwest'
  ];
  const index = Math.round(degrees / 45) % 8;
  return cardinals[index];
}

/**
 * Convert mph to km/h
 */
export function mphToKmh(mph: number): number {
  return mph * 1.60934;
}

/**
 * Convert km/h to mph
 */
export function kmhToMph(kmh: number): number {
  return kmh / 1.60934;
}

/**
 * Convert mph to knots
 */
export function mphToKnots(mph: number): number {
  return mph * 0.868976;
}

/**
 * Get Beaufort scale number (0-12) with description
 */
export function getBeaufortScale(windSpeedMph: number): {
  scale: number;
  description: string;
} {
  if (windSpeedMph < 1) return { scale: 0, description: 'Calm' };
  if (windSpeedMph <= 3) return { scale: 1, description: 'Light air' };
  if (windSpeedMph <= 7) return { scale: 2, description: 'Light breeze' };
  if (windSpeedMph <= 12) return { scale: 3, description: 'Gentle breeze' };
  if (windSpeedMph <= 18) return { scale: 4, description: 'Moderate breeze' };
  if (windSpeedMph <= 24) return { scale: 5, description: 'Fresh breeze' };
  if (windSpeedMph <= 31) return { scale: 6, description: 'Strong breeze' };
  if (windSpeedMph <= 38) return { scale: 7, description: 'High wind' };
  if (windSpeedMph <= 46) return { scale: 8, description: 'Gale' };
  if (windSpeedMph <= 54) return { scale: 9, description: 'Strong gale' };
  if (windSpeedMph <= 63) return { scale: 10, description: 'Storm' };
  if (windSpeedMph <= 72) return { scale: 11, description: 'Violent storm' };
  return { scale: 12, description: 'Hurricane' };
}

/**
 * Check if wind speed is safe for climbing
 */
export function isSafeForClimbing(windSpeedMph: number): boolean {
  return windSpeedMph <= MODERATE_MAX;
}
