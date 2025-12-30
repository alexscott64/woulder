import { WeatherData } from '../../../types/weather';

/**
 * Pure precipitation calculation utilities
 * Domain-specific math for precipitation data - NO business logic
 *
 * NOTE: WeatherData.precipitation represents HOURLY precipitation amounts (inches per hour)
 * Backend samples Open-Meteo hourly data at 3-hour intervals for forecasts,
 * but provides full hourly data for historical lookups.
 */

/**
 * Get precipitation amount (already in inches per hour)
 * Each data point represents 1 hour of precipitation
 */
export function getHourlyRate(precipitationPerHour: number): number {
  return precipitationPerHour; // Already hourly, no conversion needed
}

/**
 * Calculate total precipitation from weather data array
 * Sums hourly precipitation values
 */
export function getTotalPrecipitation(weatherData: WeatherData[]): number {
  return weatherData.reduce((sum, data) => sum + data.precipitation, 0);
}

/**
 * Calculate precipitation over last N hours
 * Filters data within time window and sums hourly precipitation values
 */
export function getPrecipitationInWindow(
  weatherData: WeatherData[],
  hoursAgo: number
): number {
  const cutoffTime = new Date(Date.now() - hoursAgo * 60 * 60 * 1000);

  return weatherData
    .filter(data => new Date(data.timestamp) >= cutoffTime)
    .reduce((sum, data) => sum + data.precipitation, 0);
}

/**
 * Get precipitation intensity category based on hourly rate
 * Thresholds based on standard meteorological definitions for hourly rainfall
 */
export function getIntensity(precipitationPerHour: number): 'none' | 'light' | 'moderate' | 'heavy' {
  if (precipitationPerHour === 0) return 'none';
  if (precipitationPerHour <= 0.01) return 'light';  // Trace to 0.01 in/hr
  if (precipitationPerHour <= 0.1) return 'light';   // 0.01-0.1 in/hr
  if (precipitationPerHour <= 0.3) return 'moderate'; // 0.1-0.3 in/hr
  return 'heavy';                                     // > 0.3 in/hr
}

/**
 * Detect if multiple recent periods had measurable precipitation
 * Checks for persistent rainfall pattern across multiple hourly readings
 */
export function hasPersistentPrecipitation(
  recentWeather: WeatherData[],
  threshold = 0.01
): boolean {
  if (!recentWeather || recentWeather.length === 0) return false;

  const periodsWithPrecip = recentWeather.filter(w => w.precipitation > threshold).length;
  return periodsWithPrecip >= 2; // 2+ periods = persistent
}
