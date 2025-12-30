import { WeatherData, WeatherCondition, ConditionLevel } from '../../../types/weather';
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
   * Combines precipitation, temperature, wind, and humidity assessments
   */
  static calculateCondition(
    weather: WeatherData,
    recentWeather?: WeatherData[]
  ): WeatherCondition {
    const reasons: string[] = [];
    let level: ConditionLevel = 'good';

    // Helper to downgrade condition level
    const downgradeCondition = (newLevel: ConditionLevel) => {
      if (newLevel === 'bad' || level === 'good') {
        level = newLevel;
      }
    };

    // 1. Assess Precipitation
    const precipCondition = PrecipitationAnalyzer.assessCondition(weather, recentWeather);
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

    // 4. Assess Humidity (only marginal, not bad)
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
}
