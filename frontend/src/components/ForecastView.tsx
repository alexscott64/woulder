import { WeatherData, DailySunTimes, WeatherCondition, RockDryingStatus } from '../types/weather';
import { TemperatureAnalyzer, ConditionCalculator } from '../utils/weather/analyzers';
import { getConditionColor, getConditionBadgeStyles, getConditionLabel, getWeatherIconUrl, getSnowDepthColor } from './weather/weatherDisplay';
import { format } from 'date-fns';
import { formatInTimeZone } from 'date-fns-tz';
import { Droplets, Droplet, Wind, Snowflake, Sunrise, Sunset, Sun, Cloud } from 'lucide-react';
import { useState, useEffect, useRef } from 'react';
import { ConditionDetailsModal } from './ConditionDetailsModal';

// Get wind color based on climbing conditions
// Good: <12 mph (green), Moderate: 12-20 mph (yellow), High: >20 mph (red)
function getWindColor(windSpeed: number): string {
  if (windSpeed <= 12) return 'text-green-600 dark:text-green-400';
  if (windSpeed <= 20) return 'text-yellow-600 dark:text-yellow-400';
  return 'text-red-600 dark:text-red-400';
}

// Get temperature color for climbing comfort (using TemperatureAnalyzer)
function getTempColor(temp: number): string {
  return TemperatureAnalyzer.getColor(temp);
}

// Get precipitation color for climbing conditions (using PrecipitationAnalyzer)
function getPrecipColor(precip: number): string {
  if (precip < 0.01) return 'text-green-600 dark:text-green-400'; // None/trace
  if (precip <= 0.04) return 'text-yellow-600 dark:text-yellow-400'; // Fair (0.01-0.04)
  return 'text-red-600 dark:text-red-400'; // Poor (> 0.04)
}

interface ForecastViewProps {
  locationId: number;
  hourlyData: WeatherData[];
  currentWeather?: WeatherData;
  historicalData?: WeatherData[];
  elevationFt?: number; // Elevation in feet for temperature adjustment
  dailySunTimes?: DailySunTimes[]; // Sunrise/sunset for each day
  dailySnowDepth?: Record<string, number>; // Backend-calculated daily snow depth forecast
  todayCondition?: WeatherCondition; // Backend-calculated today's condition
  rockDryingStatus?: RockDryingStatus; // Rock drying status for critical override
}

interface DayForecast {
  date: Date;
  dayName: string;
  high: number;
  low: number;
  avgPrecip: number;
  avgWind: number;
  avgHumidity: number; // Average humidity for the day
  avgCloudCover: number; // Average cloud cover for the day
  icon: string;
  description: string;
  condition: 'good' | 'marginal' | 'bad' | 'do_not_climb';
  hours: WeatherData[];
  hasSnow: boolean;
  hasRain: boolean;
  snowDepth: number; // Estimated snow depth in inches
  sunrise?: string;  // Sunrise time for this day
  sunset?: string;   // Sunset time for this day
  daylightHours?: number; // Total daylight hours
  effectiveSunHours?: number; // Estimated direct sun hours (accounting for clouds)
  conditionReasons?: string[]; // Combined reasons for the day's overall condition
}

// Helper to format sun time from ISO string to "7:54 AM" format in Pacific timezone
function formatSunTime(isoTime: string | undefined): string {
  if (!isoTime) return '--';
  try {
    return formatInTimeZone(isoTime, 'America/Los_Angeles', 'h:mm a');
  } catch {
    return '--';
  }
}

// Helper to calculate daylight hours from sunrise/sunset
function calculateDaylightHours(sunrise: string | undefined, sunset: string | undefined): number | undefined {
  if (!sunrise || !sunset) return undefined;
  try {
    const sunriseDate = new Date(sunrise);
    const sunsetDate = new Date(sunset);
    const diffMs = sunsetDate.getTime() - sunriseDate.getTime();
    return diffMs / (1000 * 60 * 60); // Convert to hours
  } catch {
    return undefined;
  }
}

// Helper to calculate effective sun hours based on cloud cover during daylight
// This estimates how much direct sunlight an area will receive, accounting for clouds
function calculateEffectiveSunHours(
  hours: WeatherData[],
  sunrise: string | undefined,
  sunset: string | undefined
): number | undefined {
  if (!sunrise || !sunset || hours.length === 0) return undefined;

  try {
    const sunriseTime = new Date(sunrise).getTime();
    const sunsetTime = new Date(sunset).getTime();

    let effectiveSunMinutes = 0;

    // Sort hours by timestamp
    const sortedHours = [...hours].sort((a, b) =>
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );

    for (let i = 0; i < sortedHours.length; i++) {
      const hour = sortedHours[i];
      const hourTime = new Date(hour.timestamp).getTime();

      // Determine the time interval this data point represents
      // Use 3 hours since that's our forecast interval, or time to next point
      let intervalMinutes = 180; // 3 hours default
      if (i < sortedHours.length - 1) {
        const nextHourTime = new Date(sortedHours[i + 1].timestamp).getTime();
        intervalMinutes = (nextHourTime - hourTime) / (1000 * 60);
      }

      // Calculate overlap with daylight hours
      const intervalStart = hourTime;
      const intervalEnd = hourTime + intervalMinutes * 60 * 1000;

      // Find the overlap between this interval and daylight
      const overlapStart = Math.max(intervalStart, sunriseTime);
      const overlapEnd = Math.min(intervalEnd, sunsetTime);

      if (overlapEnd > overlapStart) {
        const daylightMinutesInInterval = (overlapEnd - overlapStart) / (1000 * 60);

        // Cloud cover is 0-100%, convert to sun fraction (0% clouds = 100% sun)
        // Use a non-linear curve: thin clouds still let significant light through
        const cloudFraction = hour.cloud_cover / 100;
        // Effective sun: 100% at 0% clouds, ~50% at 50% clouds, ~10% at 100% clouds
        const sunFraction = Math.pow(1 - cloudFraction, 1.5);

        effectiveSunMinutes += daylightMinutesInInterval * sunFraction;
      }
    }

    return effectiveSunMinutes / 60; // Convert to hours
  } catch {
    return undefined;
  }
}

export function ForecastView({ locationId: _locationId, hourlyData, currentWeather, historicalData, elevationFt: _elevationFt = 0, dailySunTimes, dailySnowDepth, todayCondition, rockDryingStatus }: ForecastViewProps) {
  // State for condition details modal
  const [showConditionModal, setShowConditionModal] = useState(false);
  const [selectedDayCondition, setSelectedDayCondition] = useState<{
    dayName: string;
    level: 'good' | 'marginal' | 'bad' | 'do_not_climb';
    reasons: string[];
  } | null>(null);

  // State and ref for floating day label
  const [currentDay, setCurrentDay] = useState<string>('');
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const hourElementsRef = useRef<(HTMLDivElement | null)[]>([]);

  // Filter hourly data to show 3-hour intervals (1am, 4am, 7am, 10am, 1pm, 4pm, 7pm, 10pm)
  const filteredHourlyData = (() => {
    if (!hourlyData || hourlyData.length === 0) return [];

    // Target hours for 3-hour intervals: 1, 4, 7, 10, 13, 16, 19, 22
    const targetHours = [1, 4, 7, 10, 13, 16, 19, 22];

    const filtered: WeatherData[] = [];

    // Get current time in Pacific timezone
    const nowPacific = new Date(new Date().toLocaleString('en-US', { timeZone: 'America/Los_Angeles' }));
    const currentHour = nowPacific.getHours();
    const currentDate = formatInTimeZone(nowPacific, 'America/Los_Angeles', 'yyyy-MM-dd');

    // Find the current or most recent target hour
    // E.g., if it's 4:15pm (hour 16), current target = 16
    // E.g., if it's 5:30pm (hour 17), current target = 16 (most recent)
    let currentTargetHour = targetHours[0];
    for (const targetHour of targetHours) {
      if (currentHour >= targetHour) {
        currentTargetHour = targetHour;
      } else {
        break;
      }
    }

    let foundCurrent = false;
    const addedHours = new Set<string>(); // Track "date-hour" combinations we've added

    // Go through all hourly data and pick target hours
    for (let i = 0; i < hourlyData.length && filtered.length < 48; i++) {
      const timestamp = hourlyData[i].timestamp;

      // Get hour in Pacific timezone
      const hourPacific = parseInt(formatInTimeZone(timestamp, 'America/Los_Angeles', 'H'));
      const datePacific = formatInTimeZone(timestamp, 'America/Los_Angeles', 'yyyy-MM-dd');

      // Only process target hours
      if (!targetHours.includes(hourPacific)) continue;

      // Skip if we've already added this date-hour combination
      const dateHourKey = `${datePacific}-${hourPacific}`;
      if (addedHours.has(dateHourKey)) continue;

      if (!foundCurrent) {
        // Try to find current target hour on today
        if (datePacific === currentDate && hourPacific === currentTargetHour) {
          filtered.push(hourlyData[i]);
          addedHours.add(dateHourKey);
          foundCurrent = true;
          continue;
        }

        // Fallback: If we're on today and this hour is >= current hour, this becomes "Now"
        if (datePacific === currentDate && hourPacific >= currentHour) {
          filtered.push(hourlyData[i]);
          addedHours.add(dateHourKey);
          foundCurrent = true;
          continue;
        }
      } else {
        // After we've found the current hour, include all future target hours
        filtered.push(hourlyData[i]);
        addedHours.add(dateHourKey);
      }
    }

    return filtered;
  })();

  // Update floating day label based on scroll position
  useEffect(() => {
    const handleScroll = () => {
      if (!scrollContainerRef.current) return;

      const container = scrollContainerRef.current;
      const containerLeft = container.getBoundingClientRect().left;

      // Find which hour element is currently at the left edge
      for (let i = 0; i < hourElementsRef.current.length; i++) {
        const element = hourElementsRef.current[i];
        if (!element) continue;

        const rect = element.getBoundingClientRect();
        const relativeLeft = rect.left - containerLeft;

        // If this element is visible at the left edge (within 100px threshold)
        if (relativeLeft >= -50 && relativeLeft <= 100) {
          const hour = filteredHourlyData[i];
          if (hour) {
            const dayName = formatInTimeZone(hour.timestamp, 'America/Los_Angeles', 'EEEE');
            setCurrentDay(dayName);
          }
          break;
        }
      }
    };

    const container = scrollContainerRef.current;
    if (container) {
      // Set initial day
      handleScroll();
      container.addEventListener('scroll', handleScroll);
      return () => container.removeEventListener('scroll', handleScroll);
    }
  }, [filteredHourlyData]);

  // Calculate total rain from last 48 hours (affects climbing conditions)
  // Use ALL data (historical + current/hourly) to match WeatherCard display calculation
  const currentTime = new Date();
  const fortyEightHoursAgo = currentTime.getTime() - 48 * 60 * 60 * 1000;

  // Combine and deduplicate all data by timestamp
  const allDataForRainCalc = new Map<string, WeatherData>();
  const safeHistorical = historicalData || [];
  const safeHourly = hourlyData || [];
  [...safeHistorical, ...safeHourly].forEach(d => {
    allDataForRainCalc.set(d.timestamp.toString(), d);
  });

  // Calculate rain in last 48h
  const historical48hRain = Array.from(allDataForRainCalc.values())
    .filter(d => {
      const time = new Date(d.timestamp).getTime();
      return time >= fortyEightHoursAgo && time <= currentTime.getTime();
    })
    .reduce((sum, d) => sum + d.precipitation, 0);

  // Group hourly data by day
  const dailyForecasts: DayForecast[] = [];

  // Create a map of sun times by date for quick lookup
  const sunTimesByDate = new Map<string, DailySunTimes>();
  if (dailySunTimes) {
    dailySunTimes.forEach(st => sunTimesByDate.set(st.date, st));
  }

  // Use Pacific timezone to get today's date (must match WeatherCard logic)
  const now = new Date();
  const todayStr = formatInTimeZone(now, 'America/Los_Angeles', 'yyyy-MM-dd');

  // Include current weather in the data if provided
  const allData = currentWeather ? [currentWeather, ...hourlyData] : hourlyData;

  // Use backend-calculated snow depth (backend now handles all snow calculations)
  const snowDepthByDay = dailySnowDepth
    ? new Map(Object.entries(dailySnowDepth))
    : new Map<string, number>(); // Empty map if backend data not available

  // Deduplicate all data by timestamp before grouping
  const deduplicatedMap = new Map<string, WeatherData>();

  // First add historical data
  if (historicalData) {
    historicalData.forEach(data => {
      deduplicatedMap.set(data.timestamp.toString(), data);
    });
  }

  // Then add current/hourly data (overwrites any duplicates from historical)
  allData.forEach(data => {
    deduplicatedMap.set(data.timestamp.toString(), data);
  });

  // Group deduplicated data by day (in Pacific timezone)
  const days = new Map<string, WeatherData[]>();
  Array.from(deduplicatedMap.values()).forEach(data => {
    // Group by Pacific timezone date
    const dateKey = formatInTimeZone(data.timestamp, 'America/Los_Angeles', 'yyyy-MM-dd');

    // Include all data from today onwards
    if (dateKey >= todayStr) {
      if (!days.has(dateKey)) {
        days.set(dateKey, []);
      }
      days.get(dateKey)!.push(data);
    }
  });

  // Sort days by date
  const sortedDays = Array.from(days.entries()).sort((a, b) => a[0].localeCompare(b[0]));

  // Calculate daily summaries (show up to 6 days)
  for (let i = 0; i < Math.min(sortedDays.length, 6); i++) {
    const [dateKey, hours] = sortedDays[i];

    const date = new Date(dateKey + 'T00:00:00');
    const temps = hours.map(h => h.temperature);
    const high = Math.max(...temps);
    const low = Math.min(...temps);
    const totalPrecip = hours.reduce((sum, h) => sum + h.precipitation, 0);
    const avgPrecip = totalPrecip / hours.length; // Average precipitation per hour
    const avgWind = hours.reduce((sum, h) => sum + h.wind_speed, 0) / hours.length;
    const avgHumidity = hours.reduce((sum, h) => sum + h.humidity, 0) / hours.length;
    const avgCloudCover = hours.reduce((sum, h) => sum + h.cloud_cover, 0) / hours.length;

    // Use most common weather icon
    const iconCounts = new Map<string, number>();
    hours.forEach(h => {
      iconCounts.set(h.icon, (iconCounts.get(h.icon) || 0) + 1);
    });
    const icon = Array.from(iconCounts.entries()).sort((a, b) => b[1] - a[1])[0][0];
    const description = hours[Math.floor(hours.length / 2)].description;

    // Check if this date is today (we'll use backend condition for today)
    const isToday = dateKey === todayStr;

    // Determine overall condition for the day and collect reasons
    let condition: 'good' | 'marginal' | 'bad' | 'do_not_climb' = 'good';
    const conditionReasons: string[] = [];

    // For today, use backend-calculated condition (backend is authoritative for today)
    if (isToday && todayCondition) {
      // Override to 'do_not_climb' if rock is critical (wet-sensitive like sandstone)
      if (rockDryingStatus?.status === 'critical') {
        condition = 'do_not_climb';
        conditionReasons.push(rockDryingStatus.message);
        conditionReasons.push(...todayCondition.reasons);
      } else {
        condition = todayCondition.level;
        conditionReasons.push(...todayCondition.reasons);
      }
    } else {
      // For future days (2-6), calculate on frontend using simplified logic
      // Weight climbing hours (9am-8pm) more heavily for temperature issues
      // But keep rain/precipitation for all hours since it affects conditions later
      const isClimbingHour = (hour: WeatherData): boolean => {
        // Convert to Pacific timezone to get the correct hour
        const pacificDate = new Date(hour.timestamp).toLocaleString('en-US', { timeZone: 'America/Los_Angeles' });
        const hourOfDay = new Date(pacificDate).getHours();
        return hourOfDay >= 9 && hourOfDay < 20; // 9am to 8pm Pacific
      };

      const hourConditions = hours.map((h, index) => {
        // Pass previous hours as recent weather context for precipitation pattern analysis
        const recentHours = index > 0 ? hours.slice(Math.max(0, index - 2), index) : [];
        return {
          condition: ConditionCalculator.calculateCondition(h, recentHours),
          hour: h,
          isClimbingTime: isClimbingHour(h)
        };
      });

      // Filter conditions: only consider non-climbing hours if they have rain/wind issues
      const relevantConditions = hourConditions.filter(hc => {
        if (hc.isClimbingTime) return true; // Always consider climbing hours

      // For non-climbing hours, only consider if there are rain/wind issues
      const hasRainIssue = hc.condition.reasons.some(r =>
        r.toLowerCase().includes('rain') || r.toLowerCase().includes('precip')
      );
      const hasWindIssue = hc.condition.reasons.some(r =>
        r.toLowerCase().includes('wind')
      );

      return hasRainIssue || hasWindIssue;
    });

    // Find the worst condition and consolidate reasons (show worst value for each factor)
    const badHours = relevantConditions.filter(hc => hc.condition.level === 'bad');
    const marginalHours = relevantConditions.filter(hc => hc.condition.level === 'marginal');

    // Helper function to consolidate reasons by extracting worst values
    const consolidateReasons = (hours: typeof hourConditions) => {
      const factorMap = new Map<string, { reason: string; value: number }>();

      hours.forEach(hc => {
        hc.condition.reasons.forEach(reason => {
          // Skip "Drying slowly" or "recent rain" reasons - we'll add unified 48h calculation
          if (reason.includes('Drying slowly') || reason.includes('recent rain') || reason.includes('Recent rain')) {
            return;
          }

          // Extract factor type and value from reason string
          // Examples: "High humidity (90%)", "Cold (38°F)", "Heavy rain (0.15in/3h)"

          if (reason.includes('humidity')) {
            const match = reason.match(/(\d+)%/);
            if (match) {
              const value = parseInt(match[1]);
              const existing = factorMap.get('humidity');
              if (!existing || value > existing.value) {
                factorMap.set('humidity', { reason, value });
              }
            }
          } else if (reason.includes('cold') || reason.includes('Too cold')) {
            const match = reason.match(/(\d+)°F/);
            if (match) {
              const value = parseInt(match[1]);
              const existing = factorMap.get('cold');
              if (!existing || value < existing.value) { // Lower is worse for cold
                factorMap.set('cold', { reason, value });
              }
            }
          } else if (reason.includes('hot') || reason.includes('Too hot') || reason.includes('Warm')) {
            const match = reason.match(/(\d+)°F/);
            if (match) {
              const value = parseInt(match[1]);
              const existing = factorMap.get('hot');
              if (!existing || value > existing.value) { // Higher is worse for hot
                factorMap.set('hot', { reason, value });
              }
            }
          } else if (reason.includes('wind')) {
            const match = reason.match(/(\d+)mph/);
            if (match) {
              const value = parseInt(match[1]);
              const existing = factorMap.get('wind');
              if (!existing || value > existing.value) {
                factorMap.set('wind', { reason, value });
              }
            }
          } else if (reason.includes('rain') || reason.includes('precip') || reason.includes('drizzle')) {
            const match = reason.match(/([\d.]+)in/);
            if (match) {
              const value = parseFloat(match[1]);
              const existing = factorMap.get('precipitation');
              if (!existing || value > existing.value) {
                factorMap.set('precipitation', { reason, value });
              }
            }
          } else {
            // For reasons without numeric values, just keep unique ones
            const key = reason.toLowerCase().replace(/[^a-z]/g, '');
            factorMap.set(key, { reason, value: 0 });
          }
        });
      });

      return Array.from(factorMap.values()).map(f => f.reason);
    };

    if (badHours.length > 0) {
      condition = 'bad';
      conditionReasons.push(...consolidateReasons(badHours));
      if (badHours.length > 1) {
        conditionReasons.push(`${badHours.length} hours with poor conditions`);
      }
    } else if (marginalHours.length > 0) {
      condition = 'marginal';
      conditionReasons.push(...consolidateReasons(marginalHours));
      if (marginalHours.length > 1) {
        conditionReasons.push(`${marginalHours.length} hours with fair conditions`);
      }
    } else {
      conditionReasons.push('Good climbing conditions all day');
    }
    } // End of else block for frontend calculation

    // Determine if there's snow or rain
    const hasSnow = hours.some(h => h.temperature <= 32 && h.precipitation > 0);
    const hasRain = hours.some(h => h.temperature > 32 && h.precipitation > 0);

    // Factor in historical rain from last 48 hours (affects rock conditions)
    // Heavy recent rain degrades conditions even if forecast is good
    if (isToday && historical48hRain > 0) {
      if (historical48hRain > 0.5) {
        // Heavy rain in last 48h - rock may still be wet/seeping
        if (condition === 'good') condition = 'marginal';
        conditionReasons.push(`Recent heavy rain (${historical48hRain.toFixed(2)}in in last 48h)`);
      } else if (historical48hRain > 0.2) {
        // Moderate rain in last 48h - consider for condition
        if (condition === 'good') {
          conditionReasons.push(`Recent rain (${historical48hRain.toFixed(2)}in in last 48h) - rock may be damp`);
        }
      }
    }

    // Get snow depth for this day (end of day depth)
    const snowDepth = snowDepthByDay.get(dateKey) || 0;

    // Get sun times for this day
    const daySunTimes = sunTimesByDate.get(dateKey);
    const sunrise = daySunTimes?.sunrise;
    const sunset = daySunTimes?.sunset;
    const daylightHours = calculateDaylightHours(sunrise, sunset);
    const effectiveSunHours = calculateEffectiveSunHours(hours, sunrise, sunset);

    dailyForecasts.push({
      date,
      dayName: isToday ? 'Today' : format(date, 'EEE'),
      high,
      low,
      avgPrecip, // Changed from totalPrecip to avgPrecip
      avgWind,
      avgHumidity,
      avgCloudCover,
      icon,
      description,
      condition,
      hours,
      hasSnow,
      hasRain,
      snowDepth,
      sunrise,
      sunset,
      daylightHours,
      effectiveSunHours,
      conditionReasons,
    });
  }

  return (
    <div className="space-y-6">
      {/* 7-Day Overview */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-7 gap-4">
        {dailyForecasts.map((day, index) => {
          const conditionColor = getConditionColor(day.condition);
          const conditionBadge = getConditionBadgeStyles(day.condition);
          const conditionLabel = getConditionLabel(day.condition);

          return (
            <div
              key={index}
              className="bg-white dark:bg-gray-800 rounded-lg border-2 border-gray-200 dark:border-gray-700 p-4 text-center hover:shadow-md transition-shadow"
            >
              {/* Day name */}
              <div className="text-sm font-bold text-gray-900 dark:text-white mb-1">{day.dayName}</div>

              {/* Condition badge */}
              <div className="flex justify-center mb-2">
                <button
                  onClick={() => {
                    setSelectedDayCondition({
                      dayName: day.dayName,
                      level: day.condition,
                      reasons: day.conditionReasons || []
                    });
                    setShowConditionModal(true);
                  }}
                  className={`px-2 py-0.5 rounded-full text-xs font-semibold border ${conditionBadge.bg} ${conditionBadge.text} ${conditionBadge.border} flex items-center gap-1 hover:opacity-80 active:opacity-100 transition-opacity cursor-pointer`}
                  title="Click for condition details"
                >
                  <div className={`w-1.5 h-1.5 rounded-full ${conditionColor}`} />
                  <span>{conditionLabel}</span>
                </button>
              </div>

              {/* Date */}
              <div className="text-xs text-gray-500 dark:text-gray-400 mb-2">
                {formatInTimeZone(day.date, 'America/Los_Angeles', 'MMM d')}
              </div>

              {/* Weather icon */}
              <img
                src={getWeatherIconUrl(day.icon, day.hours[Math.floor(day.hours.length / 2)].timestamp, day.sunrise, day.sunset)}
                alt={day.description}
                className="w-12 h-12 mx-auto"
              />

              {/* Temperature range */}
              <div className="flex items-center justify-center gap-2 mb-1">
                <div className="text-xs text-red-600 font-medium">H</div>
                <div className={`text-lg font-bold ${getTempColor(day.high)}`}>
                  {Math.round(day.high)}°
                </div>
              </div>
              <div className="flex items-center justify-center gap-2">
                <div className="text-xs text-blue-600 font-medium">L</div>
                <div className={`text-sm font-semibold ${getTempColor(day.low)}`}>
                  {Math.round(day.low)}°
                </div>
              </div>

              {/* Quick stats */}
              <div className="mt-2 pt-2 border-t border-gray-100 dark:border-gray-700 space-y-1">
                {/* Precipitation - always show */}
                <div className={`flex items-center justify-center gap-1 text-xs ${getPrecipColor(day.avgPrecip)}`}>
                  {day.hasSnow && day.hasRain ? (
                    <>
                      <Snowflake className="w-3 h-3" />
                      <Droplet className="w-3 h-3" />
                    </>
                  ) : day.hasSnow ? (
                    <Snowflake className="w-3 h-3" />
                  ) : (
                    <Droplet className="w-3 h-3" />
                  )}
                  <span className="font-semibold">{day.avgPrecip.toFixed(2)}"</span>
                </div>

                {/* Snow depth - show if there's snow on ground */}
                {day.snowDepth > 0 && (
                  <div className={`flex items-center justify-center gap-1 text-xs ${getSnowDepthColor(day.snowDepth)}`}>
                    <Snowflake className="w-3 h-3 fill-current" />
                    <span className="font-bold">{day.snowDepth.toFixed(1)}" on ground</span>
                  </div>
                )}

                {/* Wind - always show */}
                <div className={`flex items-center justify-center gap-1 text-xs ${getWindColor(day.avgWind)}`}>
                  <Wind className="w-3 h-3" />
                  <span>{Math.round(day.avgWind)} mph</span>
                </div>

                {/* Humidity */}
                <div className="flex items-center justify-center gap-1 text-xs text-gray-600 dark:text-gray-400">
                  <Droplets className="w-3 h-3" />
                  <span>{Math.round(day.avgHumidity)}%</span>
                </div>

                {/* Cloud Cover */}
                <div className="flex items-center justify-center gap-1 text-xs text-gray-600 dark:text-gray-400">
                  <Cloud className="w-3 h-3" />
                  <span>{Math.round(day.avgCloudCover)}%</span>
                </div>
              </div>

              {/* Sunrise/Sunset and Sun Hours */}
              {day.sunrise && day.sunset && (
                <div className="mt-2 pt-2 border-t border-gray-100 dark:border-gray-700 space-y-1">
                  <div className="flex items-center justify-center gap-2 text-xs text-gray-600 dark:text-gray-400">
                    <div className="flex items-center gap-0.5">
                      <Sunrise className="w-3 h-3 text-orange-400" />
                      <span>{formatSunTime(day.sunrise)}</span>
                    </div>
                    <div className="flex items-center gap-0.5">
                      <Sunset className="w-3 h-3 text-orange-600 dark:text-orange-500" />
                      <span>{formatSunTime(day.sunset)}</span>
                    </div>
                  </div>
                  {day.effectiveSunHours !== undefined && day.daylightHours && (
                    <div className="flex items-center justify-center gap-1 text-xs text-amber-600" title={`${day.daylightHours.toFixed(1)}h total daylight`}>
                      <Sun className="w-3 h-3" />
                      <span className="font-medium">{day.effectiveSunHours.toFixed(1)}h sun</span>
                    </div>
                  )}
                </div>
              )}

              {/* Description */}
              <div className="text-xs text-gray-600 dark:text-gray-400 mt-2 capitalize truncate">
                {day.description}
              </div>
            </div>
          );
        })}
      </div>

      {/* Hourly Details (scrollable) */}
      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-4">Hourly Forecast</h3>
        <div className="relative">
          {/* Sticky day label overlay */}
          {currentDay && (
            <div className="absolute left-0 top-0 z-30 bg-white dark:bg-gray-800 border-l-4 border-blue-500 dark:border-blue-400 pl-3 pr-4 py-1.5 shadow-md pointer-events-none">
              <div className="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider">
                {currentDay}
              </div>
            </div>
          )}

          {/* Scrollable container */}
          <div ref={scrollContainerRef} className="overflow-x-auto">
            {/* Day labels row */}
            <div className="mb-4 pb-2 relative">
              {/* Full-width border line */}
              <div className="absolute bottom-0 left-0 right-0 h-px bg-gray-200 dark:bg-gray-700" style={{width: `calc(${filteredHourlyData.length} * 5rem + ${filteredHourlyData.length - 1} * 1.5rem)`}}></div>

              <div className="flex gap-6">
                {filteredHourlyData.map((hour, index) => {
                  const date = new Date(hour.timestamp);
                  const prevDate = index > 0 ? new Date(filteredHourlyData[index - 1].timestamp) : null;
                  const showDayLabel = !prevDate || format(date, 'yyyy-MM-dd') !== format(prevDate, 'yyyy-MM-dd');

                  return (
                    <div key={`day-${index}`} className="flex-shrink-0 w-20 text-center">
                      {showDayLabel ? (
                        <div className="text-xs font-semibold text-gray-600 dark:text-gray-400 uppercase tracking-wide pb-2">
                          {format(date, 'EEEE')}
                        </div>
                      ) : (
                        <div className="h-6"></div>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Hourly data row */}
            <div className="flex gap-6 pb-2">
              {filteredHourlyData.map((hour, index) => {
              // The first hour in filtered data is "Now" (closest future target hour)
              const isCurrentHour = index === 0;

              // Find the sunrise/sunset for this hour's date
              const dateKey = formatInTimeZone(hour.timestamp, 'America/Los_Angeles', 'yyyy-MM-dd');
              const daySunTimes = dailySunTimes?.find(d => d.date === dateKey);

              return (
                <div
                  key={index}
                  ref={el => { hourElementsRef.current[index] = el; }}
                  className={`flex-shrink-0 w-20 text-center ${isCurrentHour ? 'bg-blue-50 dark:bg-blue-900/30 rounded-lg p-2 -m-2' : ''}`}
                >
                  {/* Time */}
                  <div className={`text-xs font-medium mb-1 ${isCurrentHour ? 'text-blue-700 dark:text-blue-300 font-bold' : 'text-gray-700 dark:text-gray-300'}`}>
                    {isCurrentHour ? 'Now' : formatInTimeZone(hour.timestamp, 'America/Los_Angeles', 'ha')}
                  </div>

                  {/* Icon */}
                  <img
                    src={getWeatherIconUrl(hour.icon, hour.timestamp, daySunTimes?.sunrise, daySunTimes?.sunset)}
                    alt={hour.description}
                    className="w-10 h-10 mx-auto"
                  />

                  {/* Temperature */}
                  <div className={`text-sm font-bold mb-1 ${getTempColor(hour.temperature)}`}>
                    {Math.round(hour.temperature)}°
                  </div>

                  {/* Precipitation (3-hour total) - always show */}
                  <div className={`flex items-center justify-center gap-1 text-xs mb-1 ${getPrecipColor(hour.precipitation)}`}>
                    {hour.temperature <= 32 && hour.precipitation > 0 ? (
                      <Snowflake className="w-3 h-3" />
                    ) : (
                      <Droplet className="w-3 h-3" />
                    )}
                    <span className="font-semibold">{hour.precipitation.toFixed(2)}"</span>
                  </div>

                  {/* Wind */}
                  <div className={`flex items-center justify-center gap-1 text-xs ${getWindColor(hour.wind_speed)}`}>
                    <Wind className="w-3 h-3" />
                    <span>{Math.round(hour.wind_speed)}</span>
                  </div>

                  {/* Humidity */}
                  <div className="flex items-center justify-center gap-1 text-xs text-gray-600 dark:text-gray-400">
                    <Droplets className="w-3 h-3" />
                    <span>{hour.humidity}%</span>
                  </div>

                  {/* Cloud Cover */}
                  <div className="flex items-center justify-center gap-1 text-xs text-gray-600 dark:text-gray-400">
                    <Cloud className="w-3 h-3" />
                    <span>{hour.cloud_cover}%</span>
                  </div>
                </div>
              );
            })}
            </div>
          </div>
        </div>
      </div>

      {/* Condition Details Modal */}
      {showConditionModal && selectedDayCondition && (
        <ConditionDetailsModal
          locationName={`${selectedDayCondition.dayName} Forecast`}
          conditionLevel={selectedDayCondition.level}
          conditionLabel={getConditionLabel(selectedDayCondition.level)}
          reasons={selectedDayCondition.reasons}
          onClose={() => {
            setShowConditionModal(false);
            setSelectedDayCondition(null);
          }}
        />
      )}
    </div>
  );
}
