/**
 * Utilities for evaluating climbing conditions based on weather data
 */

/**
 * Get temperature color based on climbing comfort
 *
 * Ideal climbing temps: 45-75°F
 * Marginal: 35-45°F or 75-85°F
 * Poor: <35°F or >85°F
 */
export function getTempColor(temp: number): string {
  if (temp >= 45 && temp <= 75) return 'text-green-600'; // Good climbing temps
  if ((temp >= 35 && temp < 45) || (temp > 75 && temp <= 85)) return 'text-yellow-600'; // Marginal
  return 'text-red-600'; // Too cold or too hot
}

/**
 * Get precipitation color for climbing conditions
 *
 * No precipitation: Good
 * Light (<0.1"): Marginal
 * Significant (>0.1"): Bad
 */
export function getPrecipColor(precip: number): string {
  if (precip === 0) return 'text-green-600'; // No rain = good
  if (precip < 0.1) return 'text-yellow-600'; // Light rain = marginal
  return 'text-red-600'; // Significant rain = bad
}

/**
 * Get human-readable description of climbing conditions
 */
export function getClimbingDescription(temp: number, precip: number, windSpeed: number): {
  level: 'good' | 'marginal' | 'bad';
  reasons: string[];
} {
  const reasons: string[] = [];
  let level: 'good' | 'marginal' | 'bad' = 'good';

  // Temperature assessment
  if (temp < 35) {
    reasons.push('Very cold');
    level = 'bad';
  } else if (temp < 45) {
    reasons.push('Cold');
    if (level === 'good') level = 'marginal';
  } else if (temp > 85) {
    reasons.push('Very hot');
    level = 'bad';
  } else if (temp > 75) {
    reasons.push('Hot');
    if (level === 'good') level = 'marginal';
  }

  // Precipitation assessment
  if (precip > 0.1) {
    reasons.push('Significant precipitation');
    level = 'bad';
  } else if (precip > 0) {
    reasons.push('Light precipitation');
    if (level === 'good') level = 'marginal';
  }

  // Wind assessment
  if (windSpeed > 25) {
    reasons.push('Very windy');
    level = 'bad';
  } else if (windSpeed > 15) {
    reasons.push('Windy');
    if (level === 'good') level = 'marginal';
  }

  if (reasons.length === 0) {
    reasons.push('Ideal conditions');
  }

  return { level, reasons };
}
