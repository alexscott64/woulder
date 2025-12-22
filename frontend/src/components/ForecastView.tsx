import { WeatherData } from '../types/weather';
import { getWeatherCondition, getConditionColor, getWeatherIconUrl } from '../utils/weatherConditions';
import { format, isSameDay } from 'date-fns';
import { Droplet, Wind, Snowflake } from 'lucide-react';

interface ForecastViewProps {
  hourlyData: WeatherData[];
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
}

export function ForecastView({ hourlyData }: ForecastViewProps) {
  // Group hourly data by day
  const dailyForecasts: DayForecast[] = [];
  const days = new Map<string, WeatherData[]>();
  const now = new Date();
  const today = format(now, 'yyyy-MM-dd');

  // Group by day (include current hour and all future data)
  hourlyData.forEach(data => {
    const date = new Date(data.timestamp);
    const dateKey = format(date, 'yyyy-MM-dd');

    // Include today and all future dates
    if (dateKey >= today) {
      if (!days.has(dateKey)) {
        days.set(dateKey, []);
      }
      days.get(dateKey)!.push(data);
    }
  });

  // Sort days by date
  const sortedDays = Array.from(days.entries()).sort((a, b) => a[0].localeCompare(b[0]));

  // Calculate daily summaries (limit to 7 days)
  let dayCount = 0;
  for (const [dateKey, hours] of sortedDays) {
    if (dayCount >= 7) break;

    const date = new Date(dateKey);
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

    dailyForecasts.push({
      date,
      dayName: dayCount === 0 ? 'Today' : format(date, 'EEE'),
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
    });

    dayCount++;
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
                <div className="text-lg font-bold text-gray-900">
                  {Math.round(day.high)}°
                </div>
              </div>
              <div className="flex items-center justify-center gap-2">
                <div className="text-xs text-blue-600 font-medium">L</div>
                <div className="text-sm text-gray-500">
                  {Math.round(day.low)}°
                </div>
              </div>

              {/* Quick stats */}
              <div className="mt-2 pt-2 border-t border-gray-100 space-y-1">
                {/* Precipitation - always show */}
                <div className="flex items-center justify-center gap-1 text-xs text-blue-600">
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
                  <span>{day.avgPrecip.toFixed(2)}"</span>
                </div>

                {day.avgWind > 10 && (
                  <div className="flex items-center justify-center gap-1 text-xs text-gray-600">
                    <Wind className="w-3 h-3" />
                    <span>{Math.round(day.avgWind)} mph</span>
                  </div>
                )}
              </div>

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
              const condition = getWeatherCondition(hour);
              const conditionColor = getConditionColor(condition.level);
              const date = new Date(hour.timestamp);

              return (
                <div
                  key={index}
                  className="flex-shrink-0 w-20 text-center"
                >
                  {/* Time */}
                  <div className="text-xs font-medium text-gray-700 mb-1">
                    {format(date, 'ha')}
                  </div>

                  {/* Icon */}
                  <img
                    src={getWeatherIconUrl(hour.icon)}
                    alt={hour.description}
                    className="w-10 h-10 mx-auto"
                  />

                  {/* Temperature */}
                  <div className="text-sm font-bold text-gray-900 mb-1">
                    {Math.round(hour.temperature)}°
                  </div>

                  {/* Condition indicator */}
                  <div className={`w-2 h-2 rounded-full ${conditionColor} mx-auto mb-2`} />

                  {/* Precipitation (average per hour for 3h period) - always show */}
                  <div className="flex items-center justify-center gap-1 text-xs text-blue-600 mb-1">
                    {hour.temperature <= 32 && hour.precipitation > 0 ? (
                      <Snowflake className="w-3 h-3" />
                    ) : (
                      <Droplet className="w-3 h-3" />
                    )}
                    <span>{(hour.precipitation / 3).toFixed(2)}"</span>
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
