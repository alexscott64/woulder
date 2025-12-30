import * as windCalc from '../calculations/wind';

/**
 * WindAnalyzer
 *
 * Assesses climbing conditions based on wind.
 * Uses pure calculation utilities from calculations/wind
 */
export class WindAnalyzer {
  // Re-export constants for backward compatibility
  static readonly CALM_MAX = windCalc.CALM_MAX;
  static readonly LIGHT_MAX = windCalc.LIGHT_MAX;
  static readonly MODERATE_MAX = windCalc.MODERATE_MAX;
  static readonly STRONG_MAX = windCalc.STRONG_MAX;

  /**
   * Get wind category
   * Delegates to pure calculation
   */
  static getCategory(windSpeedMph: number): 'calm' | 'light' | 'moderate' | 'strong' | 'dangerous' {
    return windCalc.getCategory(windSpeedMph);
  }

  /**
   * Check if wind aids evaporation/drying
   * Delegates to pure calculation
   */
  static aidsDrying(windSpeedMph: number): boolean {
    return windCalc.aidsDrying(windSpeedMph);
  }

  /**
   * Convert wind direction degrees to compass bearing
   * Delegates to pure calculation
   */
  static degreesToCompass(degrees: number): string {
    return windCalc.degreesToCompass(degrees);
  }

  /**
   * Convert wind direction degrees to cardinal direction name
   * Delegates to pure calculation
   */
  static degreesToCardinal(degrees: number): string {
    return windCalc.degreesToCardinal(degrees);
  }

  /**
   * Convert mph to km/h
   * Delegates to pure calculation
   */
  static mphToKmh(mph: number): number {
    return windCalc.mphToKmh(mph);
  }

  /**
   * Convert km/h to mph
   * Delegates to pure calculation
   */
  static kmhToMph(kmh: number): number {
    return windCalc.kmhToMph(kmh);
  }

  /**
   * Convert mph to knots
   * Delegates to pure calculation
   */
  static mphToKnots(mph: number): number {
    return windCalc.mphToKnots(mph);
  }

  /**
   * Get Beaufort scale number (0-12)
   * Delegates to pure calculation
   */
  static getBeaufortScale(windSpeedMph: number): {
    scale: number;
    description: string;
  } {
    return windCalc.getBeaufortScale(windSpeedMph);
  }

  /**
   * Check if wind is too strong for safe climbing
   * Delegates to pure calculation
   */
  static isSafeForClimbing(windSpeedMph: number): boolean {
    return windCalc.isSafeForClimbing(windSpeedMph);
  }

  /**
   * Get wind color for UI display
   * PRESENTATION LOGIC: Should be moved to components/weather/weatherDisplay
   * @deprecated Import from components/weather/weatherDisplay instead
   */
  static getColor(windSpeedMph: number): string {
    const category = this.getCategory(windSpeedMph);

    switch (category) {
      case 'calm':
      case 'light':
        return 'text-green-600 dark:text-green-400';
      case 'moderate':
        return 'text-yellow-600 dark:text-yellow-400';
      case 'strong':
      case 'dangerous':
        return 'text-red-600 dark:text-red-400';
    }
  }

  /**
   * Assess wind condition for climbing
   * BUSINESS LOGIC: Determines if wind is good/marginal/bad for climbing
   */
  static assessCondition(windSpeedMph: number): {
    level: 'good' | 'marginal' | 'bad';
    reason: string | null;
  } {
    const category = this.getCategory(windSpeedMph);
    const speed = Math.round(windSpeedMph);

    switch (category) {
      case 'calm':
      case 'light':
        return { level: 'good', reason: null };

      case 'moderate':
        return {
          level: 'marginal',
          reason: `Moderate winds (${speed}mph)`
        };

      case 'strong':
        return {
          level: 'bad',
          reason: `High winds (${speed}mph)`
        };

      case 'dangerous':
        return {
          level: 'bad',
          reason: `Dangerous winds (${speed}mph)`
        };
    }
  }
}
