import { WeatherForecast } from '../types/weather';
import { RiverData } from '../types/river';
import { API_BASE_URL } from '../services/api';
import { ConditionCalculator, WindAnalyzer } from '../utils/weather/analyzers';
import { getConditionColor, getConditionBadgeStyles, getConditionLabel, getWeatherIconUrl } from './weather/weatherDisplay';
import { PestAnalyzer } from '../utils/pests/analyzers';
import type { PestConditions } from '../utils/pests/analyzers/PestAnalyzer';
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
  const { location, current, hourly, historical, sunrise, sunset, rock_drying_status, snow_depth_inches } = forecast;

  // Calculate "Today" condition - EXACT match to ForecastView logic
  const todayCondition = useMemo(() => {
    if (!current || !hourly || hourly.length === 0) {
      return ConditionCalculator.calculateCondition(current, historical, rock_drying_status);
    }

    const now = new Date();
    const todayStr = formatInTimeZone(now, 'America/Los_Angeles', 'yyyy-MM-dd');

    const allData = [current, ...hourly];
    const todayHours = allData.filter(data => {
      const dateKey = formatInTimeZone(data.timestamp, 'America/Los_Angeles', 'yyyy-MM-dd');
      return dateKey === todayStr;
    });

    if (todayHours.length === 0) {
      return ConditionCalculator.calculateCondition(current, historical, rock_drying_status);
    }

    const isClimbingHour = (hour: typeof current): boolean => {
      const pacificDate = new Date(hour.timestamp).toLocaleString('en-US', { timeZone: 'America/Los_Angeles' });
      const hourOfDay = new Date(pacificDate).getHours();
      return hourOfDay >= 9 && hourOfDay < 20;
    };

    const hourConditions = todayHours.map((h, index) => {
      const recentHours = index > 0 ? todayHours.slice(Math.max(0, index - 2), index) : [];
      return {
        condition: ConditionCalculator.calculateCondition(h, recentHours),
        hour: h,
        isClimbingTime: isClimbingHour(h)
      };
    });

    const relevantConditions = hourConditions.filter(hc => {
      if (hc.isClimbingTime) return true;
      const hasRainIssue = hc.condition.reasons.some(r =>
        r.toLowerCase().includes('rain') || r.toLowerCase().includes('precip')
      );
      const hasWindIssue = hc.condition.reasons.some(r =>
        r.toLowerCase().includes('wind')
      );
      return hasRainIssue || hasWindIssue;
    });

    const badHours = relevantConditions.filter(hc => hc.condition.level === 'bad');
    const marginalHours = relevantConditions.filter(hc => hc.condition.level === 'marginal');

    // Consolidate reasons - same as ForecastView
    // Skip "recent rain" reasons from hourly calculations since we'll add unified 48h total below
    const consolidateReasons = (hours: typeof hourConditions) => {
      const factorMap = new Map<string, { reason: string; value: number }>();
      hours.forEach(hc => {
        hc.condition.reasons.forEach(reason => {
          // Skip "Drying slowly" or "recent rain" reasons - we'll add unified 48h calculation
          if (reason.includes('Drying slowly') || reason.includes('recent rain') || reason.includes('Recent rain')) {
            return;
          }

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
              if (!existing || value < existing.value) {
                factorMap.set('cold', { reason, value });
              }
            }
          } else if (reason.includes('hot') || reason.includes('Too hot') || reason.includes('Warm')) {
            const match = reason.match(/(\d+)°F/);
            if (match) {
              const value = parseInt(match[1]);
              const existing = factorMap.get('hot');
              if (!existing || value > existing.value) {
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
            const key = reason.toLowerCase().replace(/[^a-z]/g, '');
            factorMap.set(key, { reason, value: 0 });
          }
        });
      });
      return Array.from(factorMap.values()).map(f => f.reason);
    };

    let level: 'good' | 'marginal' | 'bad' = 'good';
    const reasons: string[] = [];

    if (badHours.length > 0) {
      level = 'bad';
      reasons.push(...consolidateReasons(badHours));
      if (badHours.length > 1) {
        reasons.push(`${badHours.length} hours with poor conditions`);
      }
    } else if (marginalHours.length > 0) {
      level = 'marginal';
      reasons.push(...consolidateReasons(marginalHours));
      if (marginalHours.length > 1) {
        reasons.push(`${marginalHours.length} hours with fair conditions`);
      }
    } else {
      reasons.push('Good climbing conditions all day');
    }

    // Factor in rain from last 48 hours - use ALL data (historical + current/hourly)
    // This matches the display calculation to avoid confusion
    const currentTime = new Date();
    const fortyEightHoursAgo = currentTime.getTime() - 48 * 60 * 60 * 1000;

    // Combine and deduplicate all data by timestamp
    const allDataMap = new Map<string, typeof current>();
    const safeHistorical = historical || [];
    const safeHourly = hourly || [];
    [...safeHistorical, ...safeHourly].forEach(d => {
      allDataMap.set(d.timestamp.toString(), d);
    });

    // Calculate total rain in last 48h from all available data
    const past48hData = Array.from(allDataMap.values()).filter(d => {
      const time = new Date(d.timestamp).getTime();
      return time >= fortyEightHoursAgo && time <= currentTime.getTime();
    });
    const rainLast48h = past48hData.reduce((sum, d) => sum + d.precipitation, 0);

    if (rainLast48h > 0.5) {
      if (level === 'good') level = 'marginal';
      reasons.push(`Recent heavy rain (${rainLast48h.toFixed(2)}in in last 48h)`);
    } else if (rainLast48h > 0.2) {
      reasons.push(`Recent rain (${rainLast48h.toFixed(2)}in in last 48h)`);
    }

    return { level, reasons };
  }, [current, hourly, historical, rock_drying_status]);

  const conditionColor = getConditionColor(todayCondition.level);
  const conditionBadge = getConditionBadgeStyles(todayCondition.level);
  const conditionLabel = getConditionLabel(todayCondition.level);

  // River crossing state
  const [riverData, setRiverData] = useState<RiverData[]>([]);
  const [hasRivers, setHasRivers] = useState(false);

  // Comprehensive conditions modal state
  const [showConditionsModal, setShowConditionsModal] = useState(false);

  // Calculate pest conditions
  const pestConditions: PestConditions | null = useMemo(() => {
    if (!current || !historical || historical.length === 0) return null;
    return PestAnalyzer.assessConditions(current, historical);
  }, [current, historical]);

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
  const conditionsCount = [rock_drying_status, hasRivers, pestConditions].filter(Boolean).length;
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

  // Calculate past 48-hour rain (total)
  const now = new Date();
  const past48h = allData.filter(d => {
    const time = new Date(d.timestamp).getTime();
    const fortyEightHoursAgo = now.getTime() - 48 * 60 * 60 * 1000;
    return time >= fortyEightHoursAgo && time <= now.getTime();
  });
  const rainLast48h = past48h.reduce((sum, d) => sum + d.precipitation, 0);

  // Calculate next 48-hour rain forecast (average per day)
  const next48h = safeHourly.filter(d => {
    const time = new Date(d.timestamp).getTime();
    const fortyEightHoursFromNow = now.getTime() + 48 * 60 * 60 * 1000;
    return time > now.getTime() && time <= fortyEightHoursFromNow;
  });
  const totalRainNext48h = next48h.reduce((sum, d) => sum + d.precipitation, 0);
  const rainNext48h = totalRainNext48h / 2; // Average per day

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
          pestConditions={pestConditions || undefined}
          riverData={riverData.length > 0 ? riverData : undefined}
          todayCondition={todayCondition}
          onClose={() => setShowConditionsModal(false)}
        />
      )}
    </div>
  );
}
