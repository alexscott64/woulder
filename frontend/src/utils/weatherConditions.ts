import { WeatherData, WeatherCondition, ConditionLevel } from '../types/weather';

/**
 * Determine the weather condition level for climbing based on weather data
 * Similar to toorainy.com's color coding system
 *
 * Precipitation Assessment:
 * - "Precipitation rate" = amount per hour (converted from 3h period)
 * - Brief isolated rain (e.g., 0.06"/hr for 1h) can dry quickly in good conditions
 * - Persistent drizzle (e.g., 0.02"/hr sustained) keeps surfaces wet
 * - Considers recent precipitation history to detect patterns
 */
export function getWeatherCondition(weather: WeatherData, recentWeather?: WeatherData[]): WeatherCondition {
  const reasons: string[] = [];
  let level: ConditionLevel = 'good';

  // Check for recent precipitation pattern (last 3-6 hours)
  let recentPrecipTotal = 0;
  let recentHoursWithPrecip = 0;
  if (recentWeather && recentWeather.length > 0) {
    // Look at last 2 data points (6 hours of 3-hour intervals)
    const recentData = recentWeather.slice(0, 2);
    recentPrecipTotal = recentData.reduce((sum, w) => sum + w.precipitation, 0);
    recentHoursWithPrecip = recentData.filter(w => w.precipitation > 0.01).length;
  }

  // Assess drying conditions (affects how quickly surfaces dry after rain)
  const hasDryingConditions =
    weather.temperature > 55 && // Warm enough to dry
    weather.cloud_cover < 50 && // Sunny (UV helps dry)
    weather.wind_speed > 5 && // Wind helps evaporation
    weather.wind_speed < 20; // But not too windy

  // Precipitation assessment
  if (weather.precipitation > 0.1) {
    // Heavy rain (>0.033"/hr rate or >0.1" in 3h period)
    level = 'bad';
    reasons.push(`Heavy rain (${weather.precipitation.toFixed(2)}in/3h)`);
  } else if (weather.precipitation > 0.05) {
    // Moderate rain (0.017-0.033"/hr rate)
    level = level === 'good' ? 'marginal' : level;
    reasons.push(`Moderate rain (${weather.precipitation.toFixed(2)}in/3h)`);
  } else if (weather.precipitation > 0.01) {
    // Light rain or drizzle (0.003-0.017"/hr rate)
    // This is where pattern matters!

    if (recentHoursWithPrecip >= 2) {
      // Persistent drizzle - keeps surfaces wet even if rate is low
      level = level === 'good' ? 'marginal' : level;
      reasons.push(`Persistent drizzle (${(recentPrecipTotal + weather.precipitation).toFixed(2)}in over 9h)`);
    } else if (!hasDryingConditions) {
      // Light rain without drying conditions = stays wet
      level = level === 'good' ? 'marginal' : level;
      reasons.push(`Light rain, poor drying (${weather.precipitation.toFixed(2)}in/3h)`);
    } else {
      // Brief light rain with good drying = may dry quickly
      // Don't downgrade condition, but note it
      reasons.push(`Brief light rain (${weather.precipitation.toFixed(2)}in/3h, drying fast)`);
    }
  } else if (recentPrecipTotal > 0.05 && !hasDryingConditions) {
    // No current rain, but recent rain + poor drying = still wet
    level = level === 'good' ? 'marginal' : level;
    reasons.push(`Drying slowly after rain (${recentPrecipTotal.toFixed(2)}in recently)`);
  }

  // Check wind speed (mph)
  if (weather.wind_speed > 20) {
    level = 'bad';
    reasons.push(`High winds (${Math.round(weather.wind_speed)}mph)`);
  } else if (weather.wind_speed > 12) {
    level = level === 'good' ? 'marginal' : level;
    reasons.push(`Moderate winds (${Math.round(weather.wind_speed)}mph)`);
  }

  // Check temperature (Fahrenheit)
  // Green: 41-65°F, Yellow: 30-40°F or 66-79°F, Red: <30°F or >79°F
  const temp = weather.temperature;
  if (temp < 30 || temp > 79) {
    level = 'bad';
    if (temp < 30) {
      reasons.push(`Too cold (${Math.round(temp)}°F)`);
    } else {
      reasons.push(`Too hot (${Math.round(temp)}°F)`);
    }
  } else if ((temp >= 30 && temp <= 40) || (temp >= 66 && temp <= 79)) {
    level = level === 'good' ? 'marginal' : level;
    if (temp >= 30 && temp <= 40) {
      reasons.push(`Cold (${Math.round(temp)}°F)`);
    } else {
      reasons.push(`Warm (${Math.round(temp)}°F)`);
    }
  }

  // Check humidity
  if (weather.humidity > 85) {
    level = level === 'good' ? 'marginal' : level;
    reasons.push(`High humidity (${weather.humidity}%)`);
  }

  // If no issues, note good conditions
  if (reasons.length === 0) {
    reasons.push('Good climbing conditions');
  }

  return { level, reasons };
}

/**
 * Get color class based on condition level
 */
export function getConditionColor(level: ConditionLevel): string {
  switch (level) {
    case 'good':
      return 'bg-green-500';
    case 'marginal':
      return 'bg-yellow-500';
    case 'bad':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
}

/**
 * Get text color class based on condition level
 */
export function getConditionTextColor(level: ConditionLevel): string {
  switch (level) {
    case 'good':
      return 'text-green-600';
    case 'marginal':
      return 'text-yellow-600';
    case 'bad':
      return 'text-red-600';
    default:
      return 'text-gray-600';
  }
}

/**
 * Get display label for condition level
 */
export function getConditionLabel(level: ConditionLevel): string {
  switch (level) {
    case 'good':
      return 'Good';
    case 'marginal':
      return 'Fair';
    case 'bad':
      return 'Poor';
    default:
      return 'Unknown';
  }
}

/**
 * Get badge styles for condition level (background + text color)
 */
export function getConditionBadgeStyles(level: ConditionLevel): { bg: string; text: string; border: string } {
  switch (level) {
    case 'good':
      return { bg: 'bg-green-100', text: 'text-green-700', border: 'border-green-300' };
    case 'marginal':
      return { bg: 'bg-yellow-100', text: 'text-yellow-700', border: 'border-yellow-300' };
    case 'bad':
      return { bg: 'bg-red-100', text: 'text-red-700', border: 'border-red-300' };
    default:
      return { bg: 'bg-gray-100', text: 'text-gray-700', border: 'border-gray-300' };
  }
}

/**
 * Get wind direction as compass bearing (N, NE, E, etc.)
 */
export function getWindDirection(degrees: number): string {
  const directions = ['N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW'];
  const index = Math.round(degrees / 45) % 8;
  return directions[index];
}

/**
 * Get OpenWeatherMap icon URL
 */
export function getWeatherIconUrl(iconCode: string): string {
  return `https://openweathermap.org/img/wn/${iconCode}@2x.png`;
}

/**
 * Calculate total precipitation over last 48 hours
 */
export function calculate48HourRain(weatherData: WeatherData[]): number {
  const now = new Date();
  const fortyEightHoursAgo = new Date(now.getTime() - 48 * 60 * 60 * 1000);

  return weatherData
    .filter(data => new Date(data.timestamp) >= fortyEightHoursAgo)
    .reduce((total, data) => total + data.precipitation, 0);
}

/**
 * Determine if there's likely snow on the ground
 * Based on recent precipitation and temperature
 */
export function getSnowProbability(weatherData: WeatherData[]): { hasSnow: boolean; probability: string } {
  const recent = weatherData.slice(0, 8); // Last 24 hours (3-hour intervals)

  // Check for freezing temps and precipitation
  let snowConditions = 0;
  let freezingHours = 0;

  for (const data of recent) {
    if (data.temperature <= 32) {
      freezingHours++;
      if (data.precipitation > 0) {
        snowConditions++;
      }
    }
  }

  // Calculate probability
  if (snowConditions >= 3) {
    return { hasSnow: true, probability: 'High' };
  } else if (snowConditions >= 1 || freezingHours >= 4) {
    return { hasSnow: true, probability: 'Moderate' };
  } else if (freezingHours >= 2) {
    return { hasSnow: false, probability: 'Low' };
  }

  return { hasSnow: false, probability: 'None' };
}
