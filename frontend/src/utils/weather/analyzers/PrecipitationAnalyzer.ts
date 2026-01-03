import { WeatherData } from '../../../types/weather';
import * as precipCalc from '../calculations/precipitation';
import * as tempCalc from '../calculations/temperature';
import * as windCalc from '../calculations/wind';

/**
 * PrecipitationAnalyzer
 *
 * Assesses climbing conditions based on precipitation.
 * Uses pure calculation utilities from calculations/precipitation
 */
export class PrecipitationAnalyzer {
  /**
   * Get precipitation rate (already in inches per hour)
   * Delegates to pure calculation
   */
  static getHourlyRate(precipitationPerHour: number): number {
    return precipCalc.getHourlyRate(precipitationPerHour);
  }

  /**
   * Calculate total precipitation over a time window
   * Delegates to pure calculation
   */
  static getTotalPrecipitation(weatherData: WeatherData[]): number {
    return precipCalc.getTotalPrecipitation(weatherData);
  }

  /**
   * Detect persistent precipitation pattern
   * Delegates to pure calculation
   */
  static hasPersistentPrecipitation(recentWeather: WeatherData[], threshold = 0.01): boolean {
    return precipCalc.hasPersistentPrecipitation(recentWeather, threshold);
  }

  /**
   * Get precipitation intensity category
   * Delegates to pure calculation
   */
  static getIntensity(precipitationPerHour: number): 'none' | 'light' | 'moderate' | 'heavy' {
    return precipCalc.getIntensity(precipitationPerHour);
  }

  /**
   * Calculate precipitation over last N hours
   * Delegates to pure calculation
   */
  static getPrecipitationInWindow(
    weatherData: WeatherData[],
    hoursAgo: number
  ): number {
    return precipCalc.getPrecipitationInWindow(weatherData, hoursAgo);
  }

  /**
   * Assess surface drying conditions
   * BUSINESS LOGIC: Determines if conditions are good for drying rocks
   */
  static assessDryingConditions(weather: WeatherData): {
    canDryQuickly: boolean;
    factors: string[];
  } {
    const factors: string[] = [];
    let score = 0;

    // Temperature factor (>55°F is good for drying)
    if (tempCalc.supportsDrying(weather.temperature)) {
      score++;
      factors.push('warm temperature');
    } else {
      factors.push('cool temperature (slow drying)');
    }

    // Cloud cover factor (<50% = sunny)
    if (weather.cloud_cover < 50) {
      score++;
      factors.push('sunny (UV helps dry)');
    } else {
      factors.push('overcast (slow drying)');
    }

    // Wind factor (5-20 mph is ideal)
    if (windCalc.aidsDrying(weather.wind_speed)) {
      score++;
      factors.push('moderate wind');
    } else if (weather.wind_speed <= windCalc.CALM_MAX) {
      factors.push('calm (little evaporation)');
    } else {
      factors.push('very windy');
    }

    return {
      canDryQuickly: score >= 2, // Need at least 2 of 3 factors
      factors
    };
  }

  /**
   * Determine if surfaces are likely still wet from recent rain
   * BUSINESS LOGIC: Climbing-specific wet surface assessment
   */
  static areSurfacesLikelyWet(
    currentPrecip: number,
    recentPrecip: number,
    dryingConditions: boolean
  ): boolean {
    // Currently raining
    if (currentPrecip > 0.01) return true;

    // Recent significant rain + poor drying
    if (recentPrecip > 0.05 && !dryingConditions) return true;

    return false;
  }

  /**
   * Get precipitation condition assessment
   * BUSINESS LOGIC: Determines if precipitation is good/marginal/bad for climbing
   */
  static assessCondition(
    current: WeatherData,
    recentWeather?: WeatherData[]
  ): {
    level: 'good' | 'marginal' | 'bad';
    reason: string | null;
  } {
    const drying = this.assessDryingConditions(current);

    // Poor (bad): > 0.3 in/hr (heavy rain)
    if (current.precipitation > 0.3) {
      return {
        level: 'bad',
        reason: `Heavy rain (${current.precipitation.toFixed(2)}in/hr)`
      };
    }

    // Marginal: 0.1 to 0.3 in/hr (moderate rain)
    if (current.precipitation >= 0.1 && current.precipitation <= 0.3) {
      return {
        level: 'marginal',
        reason: `Moderate rain (${current.precipitation.toFixed(2)}in/hr)`
      };
    }

    // Light rain/drizzle: 0.01 to 0.1 in/hr
    // Consider drying conditions and persistence
    if (current.precipitation >= 0.01 && current.precipitation < 0.1) {
      const recentTotal = recentWeather ? this.getTotalPrecipitation(recentWeather) : 0;
      const isPersistent = recentWeather ? this.hasPersistentPrecipitation(recentWeather) : false;

      // Persistent drizzle over multiple hours
      if (isPersistent) {
        const hours = recentWeather ? recentWeather.length : 0;
        return {
          level: 'marginal',
          reason: `Persistent drizzle (${(recentTotal + current.precipitation).toFixed(2)}in over ${hours}h)`
        };
      }

      // Light rain with poor drying conditions
      if (!drying.canDryQuickly) {
        return {
          level: 'marginal',
          reason: `Light rain, poor drying (${current.precipitation.toFixed(2)}in/hr)`
        };
      }

      // Brief light rain with good drying conditions - can climb!
      return {
        level: 'good',
        reason: `Light rain (${current.precipitation.toFixed(2)}in/hr, drying fast)`
      };
    }

    // No current rain, but check recent rain (last 24 hours only)
    if (recentWeather) {
      // Only look at last 24 hours for "recent" rain assessment
      const last24hTotal = this.getPrecipitationInWindow(recentWeather, 24);

      if (last24hTotal > 0.05 && !drying.canDryQuickly) {
        return {
          level: 'marginal',
          reason: `Drying slowly after rain (${last24hTotal.toFixed(2)}in in last 24h)`
        };
      }
    }

    // No precipitation issues
    return { level: 'good', reason: null };
  }
}
