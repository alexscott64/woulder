/**
 * Pure temperature calculation utilities
 * Domain-specific math for temperature data - NO business logic
 */

// Climbing temperature thresholds (Fahrenheit)
export const IDEAL_MIN = 41;
export const IDEAL_MAX = 70;
export const COLD_MIN = 30;
export const COLD_MAX = 40;
export const WARM_MIN = 71;
export const WARM_MAX = 85;
export const FREEZING = 32;

/**
 * Get temperature category based on climbing comfort
 */
export function getCategory(tempF: number): 'too-cold' | 'cold' | 'ideal' | 'warm' | 'too-hot' {
  if (tempF < COLD_MIN) return 'too-cold';
  if (tempF <= COLD_MAX) return 'cold';
  if (tempF <= IDEAL_MAX) return 'ideal';
  if (tempF <= WARM_MAX) return 'warm';
  return 'too-hot';
}

/**
 * Convert Fahrenheit to Celsius
 */
export function fahrenheitToCelsius(tempF: number): number {
  return (tempF - 32) * (5 / 9);
}

/**
 * Convert Celsius to Fahrenheit
 */
export function celsiusToFahrenheit(tempC: number): number {
  return (tempC * 9 / 5) + 32;
}

/**
 * Check if temperature is at or below freezing
 */
export function isFreezing(tempF: number): boolean {
  return tempF <= FREEZING;
}

/**
 * Check if temperature supports fast drying (>55°F)
 */
export function supportsDrying(tempF: number): boolean {
  return tempF > 55;
}

/**
 * Adjust temperature for elevation (lapse rate: ~3.5°F per 1000ft)
 */
export function adjustForElevation(seaLevelTempF: number, elevationFt: number): number {
  const LAPSE_RATE = 3.5 / 1000; // degrees F per foot
  return seaLevelTempF - (elevationFt * LAPSE_RATE);
}

/**
 * Calculate wind chill temperature
 * Only applies when temp <= 50°F and wind > 3 mph
 */
export function calculateWindChill(tempF: number, windSpeedMph: number): number {
  if (tempF > 50 || windSpeedMph <= 3) return tempF;

  const windChill = 35.74
    + (0.6215 * tempF)
    - (35.75 * Math.pow(windSpeedMph, 0.16))
    + (0.4275 * tempF * Math.pow(windSpeedMph, 0.16));

  return windChill;
}

/**
 * Calculate heat index
 * Only applies when temp >= 80°F and humidity > 40%
 */
export function calculateHeatIndex(tempF: number, humidityPercent: number): number {
  if (tempF < 80 || humidityPercent <= 40) return tempF;

  const T = tempF;
  const RH = humidityPercent;

  const heatIndex = -42.379
    + (2.04901523 * T)
    + (10.14333127 * RH)
    - (0.22475541 * T * RH)
    - (0.00683783 * T * T)
    - (0.05481717 * RH * RH)
    + (0.00122874 * T * T * RH)
    + (0.00085282 * T * RH * RH)
    - (0.00000199 * T * T * RH * RH);

  return heatIndex;
}

/**
 * Calculate feels-like temperature considering wind chill and heat index
 */
export function calculateFeelsLike(
  tempF: number,
  windSpeedMph: number,
  humidityPercent: number
): number {
  // Wind chill for cold temps
  if (tempF <= 50 && windSpeedMph > 3) {
    return calculateWindChill(tempF, windSpeedMph);
  }

  // Heat index for hot temps
  if (tempF >= 80 && humidityPercent > 40) {
    return calculateHeatIndex(tempF, humidityPercent);
  }

  // No adjustment needed
  return tempF;
}
