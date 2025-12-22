import { WeatherData } from '../types/weather';
import { format } from 'date-fns';

/**
 * SWE-Based Temperature-Indexed Snow Model (PNW-friendly)
 *
 * Tracks snow accumulation and melt using Snow Water Equivalent (SWE) and snow density
 * - Snow accumulation with temperature-dependent density and freezing level transition
 * - Rain-on-snow compaction and melt
 * - Temperature-based melt above 34°F
 * - Wind-enhanced melt and sublimation
 * - Humidity-based sublimation
 * - Elevation-adjusted temperatures
 * - Natural compaction / settling
 *
 * @param historicalData - Past weather data (typically 7 days)
 * @param futureData - Current and forecast weather data
 * @param elevationFt - Elevation of location in feet (optional, for temperature adjustment)
 * @returns Map of date keys to snow depth in inches
 */
export function calculateSnowAccumulation(
  historicalData: WeatherData[],
  futureData: WeatherData[],
  elevationFt: number = 0
): Map<string, number> {
  const snowDepthByDay = new Map<string, number>();

  let snowSWE = 0;        // inches of water equivalent
  let snowDensity = 0.12; // fraction (0.08–0.35 typical)

  // Combine all data chronologically
  const allData = [...(historicalData || []), ...futureData].sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  allData.forEach((hour) => {
    // Elevation-adjusted temperature (lapse rate: -3.5°F per 1000 ft)
    const temp = hour.temperature - (elevationFt / 1000) * 3.5;
    const precip = hour.precipitation;
    const windSpeed = hour.wind_speed;
    const humidity = hour.humidity;

    // --- Freezing Level Transition (30-34°F mix zone) ---
    const snowFraction = getSnowFraction(temp);

    if (precip > 0) {
      if (snowFraction > 0) {
        // Snowfall portion
        const snowPrecip = precip * snowFraction;
        snowSWE += snowPrecip;
        const newSnowDensity = getNewSnowDensity(temp);
        snowDensity = blendDensity(snowDensity, snowSWE, snowPrecip, newSnowDensity);
      }

      if (snowFraction < 1 && snowSWE > 0) {
        // Rain-on-snow portion
        const rainPrecip = precip * (1 - snowFraction);
        snowSWE += rainPrecip * 0.7; // most rain infiltrates the pack
        snowDensity = Math.min(0.35, snowDensity + 0.03); // pack compacts

        // Rain energy melt (warmer rain melts more snow)
        const rainTemp = Math.max(temp, 32);
        const rainEnergyMelt = rainPrecip * (rainTemp - 32) * 0.01;
        snowSWE = Math.max(0, snowSWE - rainEnergyMelt);
      }
    }

    // --- Temperature-based melt ---
    if (temp > 34 && snowSWE > 0) {
      const melt = calculateSWEMelt(temp);
      snowSWE = Math.max(0, snowSWE - melt);
    }

    // --- Wind-enhanced melt and sublimation ---
    if (windSpeed > 10 && snowSWE > 0) {
      const windMelt = (windSpeed - 10) * 0.002;
      snowSWE = Math.max(0, snowSWE - windMelt);
    }

    // --- Humidity-based sublimation (dry air = more sublimation) ---
    if (humidity < 60 && snowSWE > 0) {
      const sublimation = (60 - humidity) * 0.0001;
      snowSWE = Math.max(0, snowSWE - sublimation);
    }

    // --- Compaction / settling ---
    if (snowSWE > 0) {
      snowDensity = Math.min(0.4, snowDensity + getCompactionRate(temp));
    }

    // --- Derive depth ---
    const snowDepth = snowSWE > 0 ? snowSWE / snowDensity : 0;

    // --- Store daily snow depth ---
    const dateKey = format(new Date(hour.timestamp), 'yyyy-MM-dd');
    snowDepthByDay.set(dateKey, snowDepth);
  });

  return snowDepthByDay;
}

/**
 * Calculate snow fraction based on temperature (freezing level transition)
 * Returns 1.0 for all snow (temp <= 30°F)
 * Returns 0.0 for all rain (temp >= 34°F)
 * Returns gradient between 30-34°F
 */
function getSnowFraction(temp: number): number {
  if (temp <= 30) return 1.0; // All snow
  if (temp >= 34) return 0.0; // All rain
  // Linear interpolation in transition zone
  return (34 - temp) / 4;
}

/**
 * Estimate density of new snow based on temperature
 */
function getNewSnowDensity(temp: number): number {
  if (temp <= 20) return 0.08; // very fluffy
  if (temp <= 28) return 0.12; // cold
  if (temp <= 32) return 0.18; // near freezing, wet
  return 0.2; // default slightly wet
}

/**
 * Blend new snow density with existing pack
 */
function blendDensity(
  currentDensity: number,
  currentSWE: number,
  newSWE: number,
  newDensity: number
): number {
  return (currentDensity * (currentSWE - newSWE) + newDensity * newSWE) / currentSWE;
}

/**
 * Temperature-driven SWE melt (PNW degree-day approximation)
 */
function calculateSWEMelt(temp: number): number {
  // inches SWE per hour
  return Math.max(0, (temp - 34) * 0.01);
}

/**
 * Compaction rate based on temperature
 */
function getCompactionRate(temp: number): number {
  if (temp < 20) return 0.0003;
  if (temp < 28) return 0.0006;
  if (temp < 32) return 0.0012;
  return 0.0025; // warm, wet snow compacts fast
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
