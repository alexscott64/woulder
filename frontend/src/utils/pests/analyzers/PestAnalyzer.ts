import { WeatherData } from '../../../types/weather';
import * as pestCalc from '../calculations/pests';

/**
 * Pest and Mosquito Activity Analyzer
 *
 * Assesses pest conditions based on weather data.
 * Based on entomological research on insect activity patterns.
 */

export interface PestConditions {
  mosquitoLevel: pestCalc.PestLevel;
  mosquitoScore: number; // 0-100
  outdoorPestLevel: pestCalc.PestLevel;
  outdoorPestScore: number; // 0-100
  factors: string[];
}

export interface DayPestConditions {
  mosquitoLevel: pestCalc.PestLevel;
  outdoorPestLevel: pestCalc.PestLevel;
  worstLevel: pestCalc.PestLevel;
}

export class PestAnalyzer {
  /**
   * Calculate mosquito activity level based on weather conditions
   * BUSINESS LOGIC: Determines mosquito threat for outdoor activities
   *
   * Key factors from research:
   * - Optimal temp: 70-85°F (active 50-95°F, dormant below 50°F)
   * - Humidity > 50% preferred (they dehydrate easily)
   * - Standing water from rain 7-14 days ago = breeding sites
   * - Low wind (< 10 mph) allows flight
   * - Season matters: peak activity late spring through early fall
   */
  static calculateMosquitoScore(
    currentTemp: number,
    currentHumidity: number,
    currentWind: number,
    recentRainfall: number,
    month: number
  ): { score: number; factors: string[] } {
    let score = 0;
    const factors: string[] = [];

    // Temperature is a GATING factor - mosquitoes are dormant below 50°F
    if (currentTemp < 50) {
      factors.push('Too cold for mosquitoes (below 50°F)');
      return { score: Math.min(5, Math.round(currentTemp / 10)), factors };
    }

    // Temperature factor (0-30 points)
    if (currentTemp >= 70 && currentTemp <= 85) {
      score += 30;
      factors.push('Optimal mosquito temperature');
    } else if (currentTemp >= 60 && currentTemp < 70) {
      score += 20;
      factors.push('Warm enough for mosquito activity');
    } else if (currentTemp >= 85 && currentTemp <= 95) {
      score += 20;
      factors.push('Hot but mosquitoes still active');
    } else if (currentTemp >= 50 && currentTemp < 60) {
      score += 10;
      factors.push('Cool - reduced mosquito activity');
    } else if (currentTemp > 95) {
      score += 5;
      factors.push('Too hot - mosquitoes seek shade');
    }

    // Humidity factor (0-25 points)
    if (currentHumidity >= 70) {
      score += 25;
      factors.push('High humidity favors mosquitoes');
    } else if (currentHumidity >= 50) {
      score += 15;
      factors.push('Moderate humidity');
    } else if (currentHumidity >= 30) {
      score += 5;
      factors.push('Low humidity limits mosquitoes');
    } else {
      score += 0;
      factors.push('Very dry - mosquitoes dehydrate');
    }

    // Recent rainfall factor (0-25 points)
    if (recentRainfall >= 2) {
      score += 25;
      factors.push('Recent heavy rain created breeding sites');
    } else if (recentRainfall >= 1) {
      score += 20;
      factors.push('Recent rain provides breeding habitat');
    } else if (recentRainfall >= 0.5) {
      score += 12;
      factors.push('Some recent moisture');
    } else if (recentRainfall >= 0.1) {
      score += 5;
      factors.push('Minimal recent rainfall');
    } else {
      score += 0;
      factors.push('Dry conditions limit breeding');
    }

    // Wind factor (0-10 points)
    if (currentWind <= 5) {
      score += 10;
      factors.push('Calm conditions favor mosquitoes');
    } else if (currentWind <= 10) {
      score += 6;
      factors.push('Light wind');
    } else if (currentWind <= 15) {
      score += 3;
      factors.push('Moderate wind limits flight');
    } else {
      score += 0;
      factors.push('Strong wind grounds mosquitoes');
    }

    // Seasonal factor (0-10 points)
    const seasonalMultiplier = pestCalc.getSeasonalMosquitoFactor(month);
    score += seasonalMultiplier * 10;

    if (seasonalMultiplier >= 0.8) {
      factors.push('Peak mosquito season');
    } else if (seasonalMultiplier >= 0.5) {
      factors.push('Active mosquito season');
    } else if (seasonalMultiplier >= 0.2) {
      factors.push('Early/late season');
    } else {
      factors.push('Off-season for mosquitoes');
    }

    return { score: Math.min(100, Math.max(0, score)), factors };
  }

  /**
   * Calculate general outdoor pest activity
   * BUSINESS LOGIC: Determines threat from flies, gnats, wasps, bees, ants, spiders, etc.
   */
  static calculateOutdoorPestScore(
    currentTemp: number,
    currentHumidity: number,
    recentRainfall: number,
    month: number
  ): { score: number; factors: string[] } {
    let score = 0;
    const factors: string[] = [];

    // Temperature is a GATING factor
    if (currentTemp < 50) {
      factors.push('Too cold for most insects (below 50°F)');
      return { score: Math.min(8, Math.round(currentTemp / 6)), factors };
    }

    // Temperature factor (0-40 points)
    if (currentTemp >= 75 && currentTemp <= 90) {
      score += 40;
      factors.push('Optimal temperature for insect activity');
    } else if (currentTemp >= 65 && currentTemp < 75) {
      score += 30;
      factors.push('Warm - good insect activity');
    } else if (currentTemp >= 90 && currentTemp <= 100) {
      score += 30;
      factors.push('Hot - high insect activity');
    } else if (currentTemp >= 55 && currentTemp < 65) {
      score += 15;
      factors.push('Cool - reduced insect activity');
    } else if (currentTemp >= 50 && currentTemp < 55) {
      score += 8;
      factors.push('Near insect activity threshold');
    } else if (currentTemp > 100) {
      score += 15;
      factors.push('Extreme heat - insects seek shade');
    }

    // Humidity factor (0-20 points)
    if (currentHumidity >= 60 && currentHumidity <= 80) {
      score += 20;
      factors.push('Ideal humidity for insects');
    } else if (currentHumidity > 80) {
      score += 15;
      factors.push('High humidity');
    } else if (currentHumidity >= 40 && currentHumidity < 60) {
      score += 12;
      factors.push('Moderate humidity');
    } else {
      score += 5;
      factors.push('Dry conditions');
    }

    // Recent rainfall factor (0-20 points)
    if (recentRainfall >= 1 && recentRainfall < 3) {
      score += 20;
      factors.push('Recent rain increased pest breeding');
    } else if (recentRainfall >= 0.5 && recentRainfall < 1) {
      score += 15;
      factors.push('Some moisture aids pest activity');
    } else if (recentRainfall >= 3) {
      score += 12;
      factors.push('Heavy rain - mixed pest effects');
    } else {
      score += 8;
      factors.push('Dry period');
    }

    // Seasonal factor (0-20 points)
    const seasonalMultiplier = pestCalc.getSeasonalPestFactor(month);
    score += seasonalMultiplier * 20;

    if (seasonalMultiplier >= 0.8) {
      factors.push('Peak pest season');
    } else if (seasonalMultiplier >= 0.5) {
      factors.push('Active pest season');
    } else if (seasonalMultiplier >= 0.2) {
      factors.push('Early/late pest season');
    } else {
      factors.push('Low pest season');
    }

    return { score: Math.min(100, Math.max(0, score)), factors };
  }

  /**
   * Main function to calculate pest conditions
   * BUSINESS LOGIC: Comprehensive pest threat assessment
   */
  static assessConditions(
    current: WeatherData,
    historical: WeatherData[]
  ): PestConditions {
    const now = new Date();
    const month = now.getMonth() + 1; // 1-12

    // Calculate recent rainfall
    const recentRainfall = pestCalc.calculateRecentRainfall(historical, 7);

    // Calculate mosquito conditions
    const mosquitoResult = this.calculateMosquitoScore(
      current.temperature,
      current.humidity,
      current.wind_speed,
      recentRainfall,
      month
    );

    // Calculate outdoor pest conditions
    const pestResult = this.calculateOutdoorPestScore(
      current.temperature,
      current.humidity,
      recentRainfall,
      month
    );

    // Combine unique factors
    const allFactors = [...new Set([...mosquitoResult.factors, ...pestResult.factors])];
    const topFactors = allFactors.slice(0, 4);

    return {
      mosquitoLevel: pestCalc.scoreToLevel(mosquitoResult.score),
      mosquitoScore: Math.round(mosquitoResult.score),
      outdoorPestLevel: pestCalc.scoreToLevel(pestResult.score),
      outdoorPestScore: Math.round(pestResult.score),
      factors: topFactors,
    };
  }

  /**
   * Calculate pest conditions for a forecast day
   * BUSINESS LOGIC: Day-level pest threat assessment for forecasts
   */
  static assessDayConditions(
    dayHours: WeatherData[],
    forecastDate: Date,
    recentRainfall: number = 0
  ): DayPestConditions {
    if (dayHours.length === 0) {
      return {
        mosquitoLevel: 'low',
        outdoorPestLevel: 'low',
        worstLevel: 'low',
      };
    }

    const month = forecastDate.getMonth() + 1;

    // Use peak conditions for assessment
    const highTemp = pestCalc.getMaxValue(dayHours, 'temperature');
    const avgHumidity = pestCalc.getAverageValue(dayHours, 'humidity');
    const avgWind = pestCalc.getAverageValue(dayHours, 'wind_speed');

    // Include day's precipitation
    const dayPrecip = dayHours.reduce((sum, h) => sum + h.precipitation, 0);
    const totalRecentRainfall = recentRainfall + dayPrecip;

    // Calculate conditions
    const mosquitoResult = this.calculateMosquitoScore(
      highTemp,
      avgHumidity,
      avgWind,
      totalRecentRainfall,
      month
    );

    const pestResult = this.calculateOutdoorPestScore(
      highTemp,
      avgHumidity,
      totalRecentRainfall,
      month
    );

    const mosquitoLevel = pestCalc.scoreToLevel(mosquitoResult.score);
    const outdoorPestLevel = pestCalc.scoreToLevel(pestResult.score);

    return {
      mosquitoLevel,
      outdoorPestLevel,
      worstLevel: pestCalc.getWorstLevel(mosquitoLevel, outdoorPestLevel),
    };
  }
}
