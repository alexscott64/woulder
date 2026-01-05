import { WeatherData, WeatherCondition } from '../../../types/weather';

/**
 * ConditionCalculator - Minimal frontend condition calculation
 *
 * PRIMARY SOURCE: Backend calculates today's condition (backend/internal/weather/conditions.go)
 *
 * This minimal implementation is ONLY used for:
 * - Future day forecasts (days 2-6) where backend doesn't pre-calculate
 * - The backend is authoritative for "today's" condition
 */
export class ConditionCalculator {
  /**
   * Calculate condition for a single weather data point
   * Used only for future day forecasts - backend handles today's condition
   */
  static calculateCondition(
    weather: WeatherData | null | undefined,
    _recentWeather?: WeatherData[]
  ): WeatherCondition {
    if (!weather) {
      return { level: 'good', reasons: [] };
    }

    let level: 'good' | 'marginal' | 'bad' = 'good';
    const reasons: string[] = [];

    // Precipitation check
    if (weather.precipitation >= 0.05) {
      level = 'bad';
      reasons.push(`Moderate rain (${weather.precipitation.toFixed(2)}in/hr)`);
    } else if (weather.precipitation >= 0.01) {
      level = 'marginal';
      reasons.push(`Light rain (${weather.precipitation.toFixed(2)}in/hr)`);
    }

    // Temperature check
    if (weather.temperature < 40) {
      if (level === 'good') level = 'bad';
      reasons.push(`Too cold (${Math.round(weather.temperature)}°F)`);
    } else if (weather.temperature < 45) {
      if (level === 'good') level = 'marginal';
      reasons.push(`Cold (${Math.round(weather.temperature)}°F)`);
    } else if (weather.temperature > 90) {
      if (level === 'good') level = 'bad';
      reasons.push(`Too hot (${Math.round(weather.temperature)}°F)`);
    } else if (weather.temperature > 85) {
      if (level === 'good') level = 'marginal';
      reasons.push(`Warm (${Math.round(weather.temperature)}°F)`);
    }

    // Wind check
    if (weather.wind_speed > 30) {
      if (level === 'good') level = 'bad';
      reasons.push(`Dangerous winds (${Math.round(weather.wind_speed)}mph)`);
    } else if (weather.wind_speed > 20) {
      if (level === 'good') level = 'marginal';
      reasons.push(`Strong winds (${Math.round(weather.wind_speed)}mph)`);
    } else if (weather.wind_speed > 12) {
      if (level === 'good') level = 'marginal';
      reasons.push(`Moderate winds (${Math.round(weather.wind_speed)}mph)`);
    }

    // Humidity check
    if (weather.humidity >= 85) {
      if (level === 'good') level = 'marginal';
      reasons.push(`High humidity (${weather.humidity}%)`);
    }

    return { level, reasons };
  }
}
