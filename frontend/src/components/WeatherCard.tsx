import { WeatherForecast } from '../types/weather';
import { RiverData } from '../types/river';
import { API_BASE_URL } from '../services/api';
import { WindAnalyzer } from '../utils/weather/analyzers';
import { getConditionColor, getConditionBadgeStyles, getConditionLabel, getWeatherIconUrl } from './weather/weatherDisplay';
import { format } from 'date-fns';
import { formatInTimeZone } from 'date-fns-tz';
import { Cloud, Droplet, Droplets, Wind, Snowflake, ChevronDown, ChevronUp, ChevronRight, Sunrise, Sunset } from 'lucide-react';
import { useState, useEffect, useMemo } from 'react';
import { ConditionsModal } from './ConditionsModal';

interface WeatherCardProps {
  forecast: WeatherForecast;
  isExpanded: boolean;
  onToggleExpand: (expanded: boolean, todayConditionLevel?: 'good' | 'marginal' | 'bad' | 'do_not_climb') => void;
}

// Format sun time from ISO string (e.g., "2025-12-27T07:54") to "7:54 AM"
function formatSunTime(isoTime: string | undefined): string {
  if (!isoTime) return '--';
  try {
    const date = new Date(isoTime);
    return format(date, 'h:mm a');
  } catch {
    return '--';
  }
}

export function WeatherCard({ forecast, isExpanded, onToggleExpand }: WeatherCardProps) {
  const { location, current, hourly, historical, sunrise, sunset, rock_drying_status, snow_depth_inches, today_condition, rain_last_48h, rain_next_48h, pest_conditions } = forecast;

  // Use backend-calculated condition (backend always provides this)
  const todayCondition = useMemo(() => {
    // Backend should always provide the condition
    if (today_condition) {
      return today_condition;
    }

    // This should never happen - log error and return safe default
    console.error('Backend did not provide today_condition - this is a bug');
    return { level: 'good' as const, reasons: ['Weather data unavailable'] };
  }, [today_condition]);

  const conditionColor = getConditionColor(todayCondition.level);
  const conditionBadge = getConditionBadgeStyles(todayCondition.level);
  const conditionLabel = getConditionLabel(todayCondition.level);

  // River crossing state
  const [riverData, setRiverData] = useState<RiverData[]>([]);
  const [hasRivers, setHasRivers] = useState(false);

  // Comprehensive conditions modal state
  const [showConditionsModal, setShowConditionsModal] = useState(false);

  // Fetch river data when component mounts
  useEffect(() => {
    const fetchRiverData = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/rivers/location/${location.id}`);
        if (response.ok) {
          const data = await response.json();
          if (data.rivers && data.rivers.length > 0) {
            setRiverData(data.rivers);
            setHasRivers(true);
          }
        }
      } catch {
        // Silently fail - river data is optional
      }
    };

    fetchRiverData();
  }, [location.id]);

  const handleConditionsClick = async () => {
    // Fetch river data if we have rivers but haven't loaded the data yet
    if (hasRivers && riverData.length === 0) {
      try {
        const response = await fetch(`${API_BASE_URL}/rivers/location/${location.id}`);
        if (response.ok) {
          const data = await response.json();
          setRiverData(data.rivers || []);
        }
      } catch (error) {
        console.error('Error fetching river data:', error);
      }
    }
    setShowConditionsModal(true);
  };

  // Count total conditions
  const conditionsCount = [rock_drying_status, hasRivers, pest_conditions].filter(Boolean).length;
  const hasConditions = conditionsCount > 0;

  // Safely handle potentially null/undefined arrays
  const safeHistorical = historical || [];
  const safeHourly = hourly || [];

  // Combine and deduplicate by timestamp (historical and hourly can overlap)
  const allDataMap = new Map<string, typeof safeHistorical[0]>();
  [...safeHistorical, ...safeHourly].forEach(d => {
    allDataMap.set(d.timestamp.toString(), d);
  });
  const allData = Array.from(allDataMap.values()).sort((a, b) =>
    new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  // Use backend-calculated rain values if available, otherwise calculate on frontend
  const now = new Date();
  let rainLast48hValue: number;
  let rainNext48hValue: number;
  let past48h: typeof allData;
  let next48h: typeof safeHourly;

  if (rain_last_48h !== undefined && rain_next_48h !== undefined) {
    // Use backend values
    rainLast48hValue = rain_last_48h;
    rainNext48hValue = rain_next_48h / 2; // Backend gives total, we show average per day

    // Still need to get the data points for snow/rain detection
    past48h = allData.filter(d => {
      const time = new Date(d.timestamp).getTime();
      const fortyEightHoursAgo = now.getTime() - 48 * 60 * 60 * 1000;
      return time >= fortyEightHoursAgo && time <= now.getTime();
    });
    next48h = safeHourly.filter(d => {
      const time = new Date(d.timestamp).getTime();
      const fortyEightHoursFromNow = now.getTime() + 48 * 60 * 60 * 1000;
      return time > now.getTime() && time <= fortyEightHoursFromNow;
    });
  } else {
    // Fallback: Calculate on frontend
    console.warn('Backend did not provide rain values, falling back to frontend calculation');
    past48h = allData.filter(d => {
      const time = new Date(d.timestamp).getTime();
      const fortyEightHoursAgo = now.getTime() - 48 * 60 * 60 * 1000;
      return time >= fortyEightHoursAgo && time <= now.getTime();
    });
    rainLast48hValue = past48h.reduce((sum, d) => sum + d.precipitation, 0);

    next48h = safeHourly.filter(d => {
      const time = new Date(d.timestamp).getTime();
      const fortyEightHoursFromNow = now.getTime() + 48 * 60 * 60 * 1000;
      return time > now.getTime() && time <= fortyEightHoursFromNow;
    });
    const totalRainNext48h = next48h.reduce((sum, d) => sum + d.precipitation, 0);
    rainNext48hValue = totalRainNext48h / 2; // Average per day
  }

  const rainLast48h = rainLast48hValue;
  const rainNext48h = rainNext48hValue;

  // Determine precipitation type for last 48h
  const hasSnowLast48h = past48h.some(d => d.temperature <= 32 && d.precipitation > 0);
  const hasRainLast48h = past48h.some(d => d.temperature > 32 && d.precipitation > 0);

  // Determine precipitation type for next 48h
  const hasSnowNext48h = next48h.some(d => d.temperature <= 32 && d.precipitation > 0);
  const hasRainNext48h = next48h.some(d => d.temperature > 32 && d.precipitation > 0);

  // Determine if rain is bad (for last 48h total, for next 48h avg per day)
  const rainLast48hBad = rainLast48h > 2; // Total over 2 days
  const rainNext48hBad = rainNext48h > 1; // Average per day

  return (
    <div className={`bg-white dark:bg-gray-800 shadow-md hover:shadow-lg transition-all duration-200 border border-gray-200 dark:border-gray-700 ${
      isExpanded
        ? 'rounded-t-xl rounded-b-none border-b-0'
        : 'rounded-xl'
    }`}>
      {/* Main Card Content */}
      <div className="p-4 sm:p-6">
        {/* Header */}
        <div className="mb-4">
          {/* Title */}
          <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white mb-2">{location.name}</h2>

          {/* Date and Condition row */}
          <div className="flex items-center justify-between gap-3">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              {formatInTimeZone(current.timestamp, 'America/Los_Angeles', 'MMM d, h:mm a')}
            </p>

            {/* Today's Condition - Compact button */}
            <button
              onClick={handleConditionsClick}
              className="group inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-750 hover:border-gray-300 dark:hover:border-gray-600 transition-all"
              title="View detailed conditions"
            >
              {/* Status with dot */}
              <div className="flex items-center gap-1">
                <div className={`w-1.5 h-1.5 rounded-full ${conditionColor}`} />
                <span className={`text-xs font-semibold ${conditionBadge.text}`}>
                  {conditionLabel}
                </span>
              </div>

              {/* Count + chevron */}
              {hasConditions && (
                <span className="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {conditionsCount}
                </span>
              )}
              <ChevronRight className="w-3 h-3 text-gray-400 dark:text-gray-500 group-hover:translate-x-0.5 transition-transform" />
            </button>
          </div>
        </div>

        {/* Current Weather */}
        <div className="flex items-center gap-4 mb-6">
          <img
            src={getWeatherIconUrl(current.icon)}
            alt={current.description}
            className="w-20 h-20"
          />
          <div className="flex-1">
            <div className="text-4xl font-bold text-gray-900 dark:text-white">
              {Math.round(current.temperature)}°F
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400 capitalize mt-1">
              {current.description}
            </div>
          </div>
          {/* Sunrise/Sunset */}
          {(sunrise || sunset) && (
            <div className="flex flex-col gap-1 text-xs text-gray-600 dark:text-gray-400">
              <div className="flex items-center gap-1">
                <Sunrise className="w-4 h-4 text-orange-400" />
                <span>{formatSunTime(sunrise)}</span>
              </div>
              <div className="flex items-center gap-1">
                <Sunset className="w-4 h-4 text-orange-600 dark:text-orange-500" />
                <span>{formatSunTime(sunset)}</span>
              </div>
            </div>
          )}
        </div>

        {/* Rain Alert */}
        {(rainLast48hBad || rainNext48hBad) && (
          <div className="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3">
            <div className="flex items-center gap-2">
              <Droplet className="w-4 h-4 text-red-600 dark:text-red-400" />
              <span className="text-sm font-semibold text-red-900 dark:text-red-200">
                {rainLast48hBad && `${rainLast48h.toFixed(2)}" last 48h`}
                {rainLast48hBad && rainNext48hBad && ' • '}
                {rainNext48hBad && `${rainNext48h.toFixed(2)}" next 48h`}
              </span>
            </div>
          </div>
        )}

        {/* Metrics Grid - 3 columns */}
        <div className="grid grid-cols-3 gap-3 mb-4">
          {/* Last 48h Rain/Snow */}
          <div className="flex flex-col items-center text-center">
            {hasSnowLast48h && hasRainLast48h ? (
              <div className="relative w-5 h-5 mb-1">
                <Snowflake className={`w-3 h-3 absolute top-0 left-0 ${rainLast48hBad ? 'text-red-500' : 'text-blue-400'}`} />
                <Droplet className={`w-3 h-3 absolute bottom-0 right-0 ${rainLast48hBad ? 'text-red-500' : 'text-blue-500'}`} />
              </div>
            ) : hasSnowLast48h ? (
              <Snowflake className={`w-5 h-5 mb-1 ${rainLast48hBad ? 'text-red-500' : 'text-blue-400'}`} />
            ) : (
              <Droplet className={`w-5 h-5 mb-1 ${rainLast48hBad ? 'text-red-500' : 'text-blue-500'}`} />
            )}
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Last 48h</div>
            <div className={`text-sm font-semibold ${rainLast48hBad ? 'text-red-900 dark:text-red-300' : 'text-gray-900 dark:text-white'}`}>
              {rainLast48h.toFixed(2)}"
            </div>
          </div>

          {/* Next 48h Rain/Snow */}
          <div className="flex flex-col items-center text-center">
            {hasSnowNext48h && hasRainNext48h ? (
              <div className="relative w-5 h-5 mb-1">
                <Snowflake className={`w-3 h-3 absolute top-0 left-0 ${rainNext48hBad ? 'text-red-500' : 'text-blue-400'}`} />
                <Droplet className={`w-3 h-3 absolute bottom-0 right-0 ${rainNext48hBad ? 'text-red-500' : 'text-blue-500'}`} />
              </div>
            ) : hasSnowNext48h ? (
              <Snowflake className={`w-5 h-5 mb-1 ${rainNext48hBad ? 'text-red-500' : 'text-blue-400'}`} />
            ) : (
              <Droplet className={`w-5 h-5 mb-1 ${rainNext48hBad ? 'text-red-500' : 'text-blue-400'}`} />
            )}
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Next 48h</div>
            <div className={`text-sm font-semibold ${rainNext48hBad ? 'text-red-900 dark:text-red-300' : 'text-gray-900 dark:text-white'}`}>
              {rainNext48h.toFixed(2)}"
            </div>
          </div>

          {/* Snow on Ground */}
          <div className="flex flex-col items-center text-center">
            <Snowflake className={`w-5 h-5 mb-1 ${snow_depth_inches && snow_depth_inches > 0.5 ? 'text-red-500 fill-current' : 'text-blue-400'}`} />
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">On Ground</div>
            <div className={`text-sm font-semibold ${snow_depth_inches && snow_depth_inches > 0.5 ? 'text-red-900 dark:text-red-300' : 'text-gray-900 dark:text-white'}`}>
              {snow_depth_inches && snow_depth_inches > 0.1 ? `${snow_depth_inches.toFixed(1)}"` : '0"'}
            </div>
          </div>

          {/* Wind */}
          <div className="flex flex-col items-center text-center">
            <Wind className="w-5 h-5 mb-1 text-gray-600 dark:text-gray-400" />
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Wind</div>
            <div className="text-sm font-semibold text-gray-900 dark:text-white">
              {Math.round(current.wind_speed)} {WindAnalyzer.degreesToCompass(current.wind_direction)}
            </div>
          </div>

          {/* Humidity */}
          <div className="flex flex-col items-center text-center">
            <Droplets className="w-5 h-5 mb-1 text-cyan-500 dark:text-cyan-400" />
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Humidity</div>
            <div className="text-sm font-semibold text-gray-900 dark:text-white">
              {current.humidity}%
            </div>
          </div>

          {/* Cloud Cover */}
          <div className="flex flex-col items-center text-center">
            <Cloud className="w-5 h-5 mb-1 text-gray-500 dark:text-gray-400" />
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Clouds</div>
            <div className="text-sm font-semibold text-gray-900 dark:text-white">
              {current.cloud_cover}%
            </div>
          </div>
        </div>
      </div>

      {/* Expandable Forecast Section */}
      <button
        onClick={() => onToggleExpand(!isExpanded, todayCondition.level)}
        className={`w-full px-6 py-3 border-t border-gray-200 dark:border-gray-700 flex items-center justify-center gap-2 text-sm font-medium transition-colors ${
          isExpanded
            ? `${conditionColor.replace('bg-', 'bg-opacity-20 bg-')} text-gray-900 dark:text-white border-b-0`
            : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
        }`}
      >
        {isExpanded ? (
          <>
            <ChevronUp className="w-4 h-4" />
            Hide Forecast
          </>
        ) : (
          <>
            <ChevronDown className="w-4 h-4" />
            Show 6-Day Forecast
          </>
        )}
      </button>

      {/* Comprehensive Conditions Modal */}
      {showConditionsModal && (
        <ConditionsModal
          locationName={location.name}
          rockStatus={rock_drying_status}
          pestConditions={pest_conditions}
          riverData={riverData.length > 0 ? riverData : undefined}
          todayCondition={todayCondition}
          onClose={() => setShowConditionsModal(false)}
        />
      )}
    </div>
  );
}
