import * as tempCalc from '../calculations/temperature';

/**
 * TemperatureAnalyzer
 *
 * Assesses climbing conditions based on temperature.
 * Uses pure calculation utilities from calculations/temperature
 */
export class TemperatureAnalyzer {
  // Re-export constants for backward compatibility
  static readonly IDEAL_MIN = tempCalc.IDEAL_MIN;
  static readonly IDEAL_MAX = tempCalc.IDEAL_MAX;
  static readonly COLD_MIN = tempCalc.COLD_MIN;
  static readonly COLD_MAX = tempCalc.COLD_MAX;
  static readonly WARM_MIN = tempCalc.WARM_MIN;
  static readonly WARM_MAX = tempCalc.WARM_MAX;
  static readonly FREEZING = tempCalc.FREEZING;

  /**
   * Get temperature category for climbing
   * Delegates to pure calculation
   */
  static getCategory(tempF: number): 'too-cold' | 'cold' | 'ideal' | 'warm' | 'too-hot' {
    return tempCalc.getCategory(tempF);
  }

  /**
   * Convert Fahrenheit to Celsius
   * Delegates to pure calculation
   */
  static fahrenheitToCelsius(tempF: number): number {
    return tempCalc.fahrenheitToCelsius(tempF);
  }

  /**
   * Convert Celsius to Fahrenheit
   * Delegates to pure calculation
   */
  static celsiusToFahrenheit(tempC: number): number {
    return tempCalc.celsiusToFahrenheit(tempC);
  }

  /**
   * Check if temperature is at or below freezing
   * Delegates to pure calculation
   */
  static isFreezing(tempF: number): boolean {
    return tempCalc.isFreezing(tempF);
  }

  /**
   * Check if temperature supports fast drying (>55°F)
   * Delegates to pure calculation
   */
  static supportsDrying(tempF: number): boolean {
    return tempCalc.supportsDrying(tempF);
  }

  /**
   * Adjust temperature for elevation (lapse rate: ~3.5°F per 1000ft)
   * Delegates to pure calculation
   */
  static adjustForElevation(seaLevelTempF: number, elevationFt: number): number {
    return tempCalc.adjustForElevation(seaLevelTempF, elevationFt);
  }

  /**
   * Get feels-like temperature
   * Delegates to pure calculation
   */
  static calculateFeelsLike(tempF: number, windSpeedMph: number, humidityPercent: number): number {
    return tempCalc.calculateFeelsLike(tempF, windSpeedMph, humidityPercent);
  }

  /**
   * Get temperature color for UI display
   * PRESENTATION LOGIC: Should be moved to components/weather/weatherDisplay
   * @deprecated Import from components/weather/weatherDisplay instead
   */
  static getColor(tempF: number): string {
    const category = this.getCategory(tempF);

    switch (category) {
      case 'ideal':
        return 'text-green-600 dark:text-green-400';
      case 'cold':
      case 'warm':
        return 'text-yellow-600 dark:text-yellow-400';
      case 'too-cold':
      case 'too-hot':
        return 'text-red-600 dark:text-red-400';
    }
  }

  /**
   * Assess temperature condition for climbing
   * BUSINESS LOGIC: Determines if temperature is good/marginal/bad for climbing
   */
  static assessCondition(tempF: number): {
    level: 'good' | 'marginal' | 'bad';
    reason: string | null;
  } {
    const category = this.getCategory(tempF);
    const temp = Math.round(tempF);

    switch (category) {
      case 'too-cold':
        return {
          level: 'bad',
          reason: `Too cold (${temp}°F)`
        };
      case 'cold':
        return {
          level: 'marginal',
          reason: `Cold (${temp}°F)`
        };
      case 'ideal':
        return { level: 'good', reason: null };
      case 'warm':
        return {
          level: 'marginal',
          reason: `Warm (${temp}°F)`
        };
      case 'too-hot':
        return {
          level: 'bad',
          reason: `Too hot (${temp}°F)`
        };
    }
  }
}
