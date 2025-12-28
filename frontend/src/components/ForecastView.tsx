import { WeatherData, DailySunTimes } from '../types/weather';
import { getWeatherCondition, getConditionColor, getWeatherIconUrl } from '../utils/weatherConditions';
import { calculateSnowAccumulation, getSnowDepthColor } from '../utils/snowAccumulation';
import { getTempColor, getPrecipColor } from '../utils/climbingConditions';
import { calculateDayPestConditions, getPestLevelColor, getPestLevelText, PestLevel } from '../utils/pestConditions';
import { format } from 'date-fns';
import { Droplet, Wind, Snowflake, Sunrise, Sunset, Sun, Bug } from 'lucide-react';

interface ForecastViewProps {
  hourlyData: WeatherData[];
  currentWeather?: WeatherData;
  historicalData?: WeatherData[];
  elevationFt?: number; // Elevation in feet for temperature adjustment
  dailySunTimes?: DailySunTimes[]; // Sunrise/sunset for each day
}

interface DayForecast {
  date: Date;
  dayName: string;
  high: number;
  low: number;
  avgPrecip: number;
  avgWind: number;
  icon: string;
  description: string;
  condition: 'good' | 'marginal' | 'bad';
  hours: WeatherData[];
  hasSnow: boolean;
  hasRain: boolean;
  snowDepth: number; // Estimated snow depth in inches
  sunrise?: string;  // Sunrise time for this day
  sunset?: string;   // Sunset time for this day
  daylightHours?: number; // Total daylight hours
  effectiveSunHours?: number; // Estimated direct sun hours (accounting for clouds)
  pestLevel?: PestLevel; // Worst pest activity level for this day
  mosquitoLevel?: PestLevel;
  outdoorPestLevel?: PestLevel;
}

// Helper to format sun time from ISO string to "7:54 AM" format
function formatSunTime(isoTime: string | undefined): string {
  if (!isoTime) return '--';
  try {
    const date = new Date(isoTime);
    return format(date, 'h:mm a');
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

export function ForecastView({ hourlyData, currentWeather, historicalData, elevationFt = 0, dailySunTimes }: ForecastViewProps) {
  // Group hourly data by day
  const dailyForecasts: DayForecast[] = [];

  // Create a map of sun times by date for quick lookup
  const sunTimesByDate = new Map<string, DailySunTimes>();
  if (dailySunTimes) {
    dailySunTimes.forEach(st => sunTimesByDate.set(st.date, st));
  }

  // Use local timezone to get today's date
  const now = new Date();
  const localToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const todayStr = format(localToday, 'yyyy-MM-dd');

  // Include current weather in the data if provided
  const allData = currentWeather ? [currentWeather, ...hourlyData] : hourlyData;

  // Calculate snow accumulation across all data with elevation adjustment
  const snowDepthByDay = calculateSnowAccumulation(historicalData || [], allData, elevationFt);

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

  // Group deduplicated data by day
  const days = new Map<string, WeatherData[]>();
  Array.from(deduplicatedMap.values()).forEach(data => {
    const timestamp = new Date(data.timestamp);
    const dateOnly = new Date(timestamp.getFullYear(), timestamp.getMonth(), timestamp.getDate());
    const dateKey = format(dateOnly, 'yyyy-MM-dd');

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
    const avgWind = hours.reduce((sum, h) => sum + h.wind_speed, 0) / hours.length;

    // Use most common weather icon
    const iconCounts = new Map<string, number>();
    hours.forEach(h => {
      iconCounts.set(h.icon, (iconCounts.get(h.icon) || 0) + 1);
    });
    const icon = Array.from(iconCounts.entries()).sort((a, b) => b[1] - a[1])[0][0];
    const description = hours[Math.floor(hours.length / 2)].description;

    // Determine overall condition for the day
    const conditions = hours.map(h => getWeatherCondition(h).level);
    let condition: 'good' | 'marginal' | 'bad' = 'good';
    if (conditions.some(c => c === 'bad')) condition = 'bad';
    else if (conditions.some(c => c === 'marginal')) condition = 'marginal';

    // Determine if there's snow or rain
    const hasSnow = hours.some(h => h.temperature <= 32 && h.precipitation > 0);
    const hasRain = hours.some(h => h.temperature > 32 && h.precipitation > 0);

    // Check if this date is today
    const isToday = dateKey === todayStr;

    // Get snow depth for this day (end of day depth)
    const snowDepth = snowDepthByDay.get(dateKey) || 0;

    // Get sun times for this day
    const daySunTimes = sunTimesByDate.get(dateKey);
    const sunrise = daySunTimes?.sunrise;
    const sunset = daySunTimes?.sunset;
    const daylightHours = calculateDaylightHours(sunrise, sunset);
    const effectiveSunHours = calculateEffectiveSunHours(hours, sunrise, sunset);

    // Calculate pest conditions for this day
    // Use cumulative rainfall from previous days in the forecast
    const previousDaysRainfall = dailyForecasts.reduce((sum, d) => sum + d.avgPrecip, 0);
    const dayPestConditions = calculateDayPestConditions(hours, date, previousDaysRainfall);

    dailyForecasts.push({
      date,
      dayName: isToday ? 'Today' : format(date, 'EEE'),
      high,
      low,
      avgPrecip: totalPrecip,
      avgWind,
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
      pestLevel: dayPestConditions.worstLevel,
      mosquitoLevel: dayPestConditions.mosquitoLevel,
      outdoorPestLevel: dayPestConditions.outdoorPestLevel,
    });
  }

  return (
    <div className="space-y-6">
      {/* 7-Day Overview */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-7 gap-4">
        {dailyForecasts.map((day, index) => {
          const conditionColor = getConditionColor(day.condition);

          return (
            <div
              key={index}
              className="bg-white rounded-lg border-2 border-gray-200 p-4 text-center hover:shadow-md transition-shadow"
            >
              {/* Day name with condition indicator */}
              <div className="flex items-center justify-center gap-2 mb-2">
                <span className="text-sm font-bold text-gray-900">{day.dayName}</span>
                <div className={`w-2 h-2 rounded-full ${conditionColor}`} />
              </div>

              {/* Date */}
              <div className="text-xs text-gray-500 mb-2">
                {format(day.date, 'MMM d')}
              </div>

              {/* Weather icon */}
              <img
                src={getWeatherIconUrl(day.icon)}
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
              <div className="mt-2 pt-2 border-t border-gray-100 space-y-1">
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

                {day.avgWind > 10 && (
                  <div className="flex items-center justify-center gap-1 text-xs text-gray-600">
                    <Wind className="w-3 h-3" />
                    <span>{Math.round(day.avgWind)} mph</span>
                  </div>
                )}
              </div>

              {/* Sunrise/Sunset and Sun Hours */}
              {day.sunrise && day.sunset && (
                <div className="mt-2 pt-2 border-t border-gray-100 space-y-1">
                  <div className="flex items-center justify-center gap-2 text-xs text-gray-600">
                    <div className="flex items-center gap-0.5">
                      <Sunrise className="w-3 h-3 text-orange-400" />
                      <span>{formatSunTime(day.sunrise)}</span>
                    </div>
                    <div className="flex items-center gap-0.5">
                      <Sunset className="w-3 h-3 text-orange-600" />
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

              {/* Pest Activity */}
              {day.pestLevel && (
                <div
                  className="mt-2 pt-2 border-t border-gray-100"
                  title={`Mosquitoes: ${getPestLevelText(day.mosquitoLevel || 'low')}, Outdoor Pests: ${getPestLevelText(day.outdoorPestLevel || 'low')}`}
                >
                  <div className={`flex items-center justify-center gap-1 text-xs ${getPestLevelColor(day.pestLevel)}`}>
                    <Bug className="w-3 h-3" />
                    <span className="font-medium">{getPestLevelText(day.pestLevel)} Bugs</span>
                  </div>
                </div>
              )}

              {/* Description */}
              <div className="text-xs text-gray-600 mt-2 capitalize truncate">
                {day.description}
              </div>
            </div>
          );
        })}
      </div>

      {/* Hourly Details (scrollable) */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <h3 className="text-lg font-bold text-gray-900 mb-4">Hourly Forecast</h3>
        <div className="overflow-x-auto">
          {/* Day labels row */}
          <div className="flex gap-6 mb-2">
            {hourlyData.slice(0, 48).map((hour, index) => {
              const date = new Date(hour.timestamp);
              const prevDate = index > 0 ? new Date(hourlyData[index - 1].timestamp) : null;
              const showDayLabel = !prevDate || format(date, 'yyyy-MM-dd') !== format(prevDate, 'yyyy-MM-dd');

              return (
                <div key={`day-${index}`} className="flex-shrink-0 w-20 text-center">
                  {showDayLabel ? (
                    <div className="text-xs font-bold text-gray-900 pb-1 border-b border-gray-300">
                      {format(date, 'EEEE')}
                    </div>
                  ) : (
                    <div className="h-5"></div>
                  )}
                </div>
              );
            })}
          </div>

          {/* Hourly data row */}
          <div className="flex gap-6 pb-2">
            {hourlyData.slice(0, 48).map((hour, index) => {
              const date = new Date(hour.timestamp);
              // Check if this hour matches the current weather timestamp
              const isCurrentHour = currentWeather &&
                Math.abs(new Date(hour.timestamp).getTime() - new Date(currentWeather.timestamp).getTime()) < 60 * 60 * 1000;

              return (
                <div
                  key={index}
                  className={`flex-shrink-0 w-20 text-center ${isCurrentHour ? 'bg-blue-50 rounded-lg p-2 -m-2' : ''}`}
                >
                  {/* Time */}
                  <div className={`text-xs font-medium mb-1 ${isCurrentHour ? 'text-blue-700 font-bold' : 'text-gray-700'}`}>
                    {isCurrentHour ? 'Now' : format(date, 'ha')}
                  </div>

                  {/* Icon */}
                  <img
                    src={getWeatherIconUrl(hour.icon)}
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
                  <div className="flex items-center justify-center gap-1 text-xs text-gray-600">
                    <Wind className="w-3 h-3" />
                    <span>{Math.round(hour.wind_speed)}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
