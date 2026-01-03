import { WeatherData, WeatherCondition, ConditionLevel, RockDryingStatus } from '../../../types/weather';
import { PrecipitationAnalyzer } from './PrecipitationAnalyzer';
import { TemperatureAnalyzer } from './TemperatureAnalyzer';
import { WindAnalyzer } from './WindAnalyzer';

/**
 * ConditionCalculator
 *
 * Orchestrates all weather analyzers to provide comprehensive climbing condition assessments.
 * This is the main entry point for weather condition calculations.
 */
export class ConditionCalculator {
  /**
   * Calculate overall climbing conditions
   * Combines precipitation, temperature, wind, humidity, and rock drying status
   */
  static calculateCondition(
    weather: WeatherData,
    recentWeather?: WeatherData[],
    rockStatus?: RockDryingStatus
  ): WeatherCondition {
    const reasons: string[] = [];
    let level: ConditionLevel = 'good';

    // Check rock drying status FIRST - this overrides all other conditions for wet-sensitive rocks
    if (rockStatus && rockStatus.status === 'critical' && rockStatus.is_wet_sensitive) {
      return {
        level: 'do_not_climb',
        reasons: [rockStatus.message]
      };
    }

    // Helper to downgrade condition level
    const downgradeCondition = (newLevel: ConditionLevel) => {
      if (newLevel === 'bad' || level === 'good') {
        level = newLevel;
      }
    };

    // Limit recent weather to last 48 hours for precipitation analysis
    // (historical data can be 7+ days for rock drying, but we only want recent for precip)
    const recentPrecipWeather = recentWeather ? this.getLastNHours(recentWeather, 48) : undefined;

    // 1. Assess Precipitation
    const precipCondition = PrecipitationAnalyzer.assessCondition(weather, recentPrecipWeather);
    if (precipCondition.reason) {
      reasons.push(precipCondition.reason);
      downgradeCondition(precipCondition.level);
    }

    // 2. Assess Wind
    const windCondition = WindAnalyzer.assessCondition(weather.wind_speed);
    if (windCondition.reason) {
      reasons.push(windCondition.reason);
      downgradeCondition(windCondition.level);
    }

    // 3. Assess Temperature
    const tempCondition = TemperatureAnalyzer.assessCondition(weather.temperature);
    if (tempCondition.reason) {
      reasons.push(tempCondition.reason);
      downgradeCondition(tempCondition.level);
    }

    // 4. Assess Humidity
    // High humidity (>85%) affects climbing comfort and grip
    if (weather.humidity > 85) {
      reasons.push(`High humidity (${weather.humidity}%)`);
      if (level === 'good') level = 'marginal';
    }

    // If no issues found, note good conditions
    if (reasons.length === 0) {
      reasons.push('Good climbing conditions');
    }

    return { level, reasons };
  }

  /**
   * @deprecated Use forecast.snow_depth_inches from backend API instead.
   * This client-side calculation is less accurate and doesn't account for
   * proper snow accumulation/melt modeling over time.
   *
   * Calculate snow probability based on temperature and precipitation
   */
  static calculateSnowProbability(weatherData: WeatherData[]): {
    hasSnow: boolean;
    probability: 'None' | 'Low' | 'Moderate' | 'High';
    accumulationInches: number;
  } {
    const recent = weatherData.slice(0, 8); // Last 24 hours (3-hour intervals)

    let snowConditions = 0;
    let freezingHours = 0;
    let snowAccumulation = 0;

    for (const data of recent) {
      const isFreezing = TemperatureAnalyzer.isFreezing(data.temperature);

      if (isFreezing) {
        freezingHours++;

        if (data.precipitation > 0) {
          snowConditions++;
          // Approximate snow accumulation (10:1 ratio for fresh snow)
          snowAccumulation += data.precipitation * 10;
        }
      }
    }

    // Determine probability
    let probability: 'None' | 'Low' | 'Moderate' | 'High';
    let hasSnow: boolean;

    if (snowConditions >= 3) {
      probability = 'High';
      hasSnow = true;
    } else if (snowConditions >= 1 || freezingHours >= 4) {
      probability = 'Moderate';
      hasSnow = true;
    } else if (freezingHours >= 2) {
      probability = 'Low';
      hasSnow = false;
    } else {
      probability = 'None';
      hasSnow = false;
    }

    return {
      hasSnow,
      probability,
      accumulationInches: Math.round(snowAccumulation * 10) / 10
    };
  }

  /**
   * Calculate 48-hour precipitation total
   */
  static calculate48HourRain(weatherData: WeatherData[]): number {
    return PrecipitationAnalyzer.getPrecipitationInWindow(weatherData, 48);
  }

  /**
   * Get weather data from the last N hours
   * Filters historical data to only include recent entries
   */
  private static getLastNHours(weatherData: WeatherData[], hours: number): WeatherData[] {
    if (!weatherData || weatherData.length === 0) return [];

    const now = new Date();
    const cutoffTime = new Date(now.getTime() - hours * 60 * 60 * 1000);

    return weatherData.filter(data => {
      const dataTime = new Date(data.timestamp);
      return dataTime >= cutoffTime;
    });
  }
}
