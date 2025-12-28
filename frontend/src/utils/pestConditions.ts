/**
 * Pest and Mosquito Activity Calculator
 *
 * Based on entomological research on insect activity patterns:
 * - Temperature thresholds for insect metabolism
 * - Humidity requirements for mosquito survival
 * - Precipitation patterns for breeding sites
 * - Wind conditions affecting flight capability
 * - Seasonal factors
 */

import { WeatherData } from '../types/weather';

export type PestLevel = 'low' | 'moderate' | 'high' | 'very_high' | 'extreme';

export interface PestConditions {
  mosquitoLevel: PestLevel;
  mosquitoScore: number; // 0-100
  outdoorPestLevel: PestLevel;
  outdoorPestScore: number; // 0-100
  factors: string[];
}

/**
 * Calculate mosquito activity level based on weather conditions
 *
 * Key factors from research:
 * - Optimal temp: 70-85°F (active 50-95°F, dormant below 50°F)
 * - Humidity > 50% preferred (they dehydrate easily)
 * - Standing water from rain 7-14 days ago = breeding sites
 * - Low wind (< 10 mph) allows flight
 * - Season matters: peak activity late spring through early fall
 */
function calculateMosquitoScore(
  currentTemp: number,
  currentHumidity: number,
  currentWind: number,
  recentRainfall: number, // Total inches in last 7 days
  month: number // 1-12
): { score: number; factors: string[] } {
  let score = 0;
  const factors: string[] = [];

  // Temperature is a GATING factor - mosquitoes are dormant below 50°F
  // This is the most important factor and should hard-limit the score
  if (currentTemp < 50) {
    // Too cold - mosquitoes are dormant, return very low score regardless of other factors
    factors.push('Too cold for mosquitoes (below 50°F)');
    return { score: Math.min(5, Math.round(currentTemp / 10)), factors };
  }

  // Temperature factor (0-30 points)
  // Mosquitoes are most active 70-85°F
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
  // Mosquitoes thrive in high humidity
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
  // Rain creates breeding sites (standing water)
  // Peak breeding 7-14 days after rain
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
  // Low wind allows mosquitoes to fly
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
  // Peak mosquito season: May-September in northern hemisphere
  const seasonalMultiplier = getSeasonalMosquitoFactor(month);
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
 * Seasonal multiplier for mosquito activity
 * Based on typical northern hemisphere patterns (WA state latitude ~47-49°N)
 */
function getSeasonalMosquitoFactor(month: number): number {
  const factors: Record<number, number> = {
    1: 0,     // January - dormant
    2: 0,     // February - dormant
    3: 0.1,   // March - emerging
    4: 0.3,   // April - increasing
    5: 0.6,   // May - active
    6: 0.9,   // June - high activity
    7: 1.0,   // July - peak
    8: 1.0,   // August - peak
    9: 0.7,   // September - declining
    10: 0.3,  // October - late season
    11: 0.1,  // November - dying off
    12: 0,    // December - dormant
  };
  return factors[month] || 0;
}

/**
 * Calculate general outdoor pest activity
 * Includes: flies, gnats, wasps, bees, ants, spiders, etc.
 *
 * Key factors:
 * - Temperature above 50°F activates most insects
 * - Optimal activity 70-90°F
 * - Humidity affects different species differently
 * - Recent weather changes can drive pests to seek shelter
 */
function calculateOutdoorPestScore(
  currentTemp: number,
  currentHumidity: number,
  recentRainfall: number,
  month: number
): { score: number; factors: string[] } {
  let score = 0;
  const factors: string[] = [];

  // Temperature is a GATING factor - most insects are dormant below 50°F
  if (currentTemp < 50) {
    // Too cold - insects are dormant, return very low score regardless of other factors
    factors.push('Too cold for most insects (below 50°F)');
    return { score: Math.min(8, Math.round(currentTemp / 6)), factors };
  }

  // Temperature factor (0-40 points)
  // Most insects are cold-blooded - activity increases with temp
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
  // Moderate to high humidity is generally favorable
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
  // Rain can increase some pests (mosquitoes, gnats) and decrease others temporarily
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
  const seasonalMultiplier = getSeasonalPestFactor(month);
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
 * Seasonal multiplier for general pest activity
 */
function getSeasonalPestFactor(month: number): number {
  const factors: Record<number, number> = {
    1: 0.1,   // January - minimal
    2: 0.1,   // February - minimal
    3: 0.3,   // March - emerging
    4: 0.5,   // April - increasing
    5: 0.8,   // May - active
    6: 1.0,   // June - high
    7: 1.0,   // July - peak
    8: 1.0,   // August - peak
    9: 0.8,   // September - still high
    10: 0.5,  // October - declining
    11: 0.2,  // November - low
    12: 0.1,  // December - minimal
  };
  return factors[month] || 0;
}

/**
 * Convert score to pest level
 */
function scoreToLevel(score: number): PestLevel {
  if (score >= 80) return 'extreme';
  if (score >= 60) return 'very_high';
  if (score >= 40) return 'high';
  if (score >= 20) return 'moderate';
  return 'low';
}

/**
 * Get color for pest level
 */
export function getPestLevelColor(level: PestLevel): string {
  switch (level) {
    case 'extreme': return 'text-red-600';
    case 'very_high': return 'text-orange-500';
    case 'high': return 'text-yellow-600';
    case 'moderate': return 'text-yellow-500';
    case 'low': return 'text-green-600';
  }
}

/**
 * Get background color for pest level
 */
export function getPestLevelBgColor(level: PestLevel): string {
  switch (level) {
    case 'extreme': return 'bg-red-500';
    case 'very_high': return 'bg-orange-500';
    case 'high': return 'bg-yellow-500';
    case 'moderate': return 'bg-yellow-400';
    case 'low': return 'bg-green-500';
  }
}

/**
 * Get display text for pest level
 */
export function getPestLevelText(level: PestLevel): string {
  switch (level) {
    case 'extreme': return 'Extreme';
    case 'very_high': return 'Very High';
    case 'high': return 'High';
    case 'moderate': return 'Moderate';
    case 'low': return 'Low';
  }
}

/**
 * Main function to calculate pest conditions
 *
 * @param current - Current weather data
 * @param historical - Historical weather data (last 7 days)
 */
export function calculatePestConditions(
  current: WeatherData,
  historical: WeatherData[]
): PestConditions {
  const now = new Date();
  const month = now.getMonth() + 1; // 1-12

  // Calculate total rainfall in the last 7 days
  const sevenDaysAgo = now.getTime() - 7 * 24 * 60 * 60 * 1000;
  const recentHistorical = historical.filter(
    h => new Date(h.timestamp).getTime() >= sevenDaysAgo
  );
  const recentRainfall = recentHistorical.reduce(
    (sum, h) => sum + h.precipitation,
    0
  );

  // Calculate mosquito conditions
  const mosquitoResult = calculateMosquitoScore(
    current.temperature,
    current.humidity,
    current.wind_speed,
    recentRainfall,
    month
  );

  // Calculate outdoor pest conditions
  const pestResult = calculateOutdoorPestScore(
    current.temperature,
    current.humidity,
    recentRainfall,
    month
  );

  // Combine unique factors, prioritizing most relevant
  const allFactors = [...new Set([...mosquitoResult.factors, ...pestResult.factors])];
  const topFactors = allFactors.slice(0, 4); // Limit to 4 factors

  return {
    mosquitoLevel: scoreToLevel(mosquitoResult.score),
    mosquitoScore: Math.round(mosquitoResult.score),
    outdoorPestLevel: scoreToLevel(pestResult.score),
    outdoorPestScore: Math.round(pestResult.score),
    factors: topFactors,
  };
}
