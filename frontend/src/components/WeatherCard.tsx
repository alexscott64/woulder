import { WeatherForecast } from '../types/weather';
import { RiverData } from '../types/river';
import { API_BASE_URL } from '../services/api';
import { ConditionCalculator, WindAnalyzer } from '../utils/weather/analyzers';
import { getConditionColor, getConditionBadgeStyles, getConditionLabel, getWeatherIconUrl } from './weather/weatherDisplay';
import { PestAnalyzer } from '../utils/pests/analyzers';
import type { PestConditions } from '../utils/pests/analyzers/PestAnalyzer';
import { getPestLevelColor } from './pests/pestDisplay';
import { format } from 'date-fns';
import { Cloud, Droplet, Droplets, Wind, Snowflake, ChevronDown, ChevronUp, Waves, Sunrise, Sunset, Bug } from 'lucide-react';
import { useState, useEffect, useMemo } from 'react';
import { RiverInfoModal } from './RiverInfoModal';
import { PestInfoModal } from './PestInfoModal';
import { ConditionDetailsModal } from './ConditionDetailsModal';

interface WeatherCardProps {
  forecast: WeatherForecast;
  isExpanded: boolean;
  onToggleExpand: (expanded: boolean) => void;
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
  const { location, current, hourly, historical, sunrise, sunset } = forecast;

  // Pass recent historical data for precipitation pattern analysis
  const condition = ConditionCalculator.calculateCondition(current, historical);
  const conditionColor = getConditionColor(condition.level);
  const conditionBadge = getConditionBadgeStyles(condition.level);
  const conditionLabel = getConditionLabel(condition.level);

  // River crossing state
  const [showRiverModal, setShowRiverModal] = useState(false);
  const [riverData, setRiverData] = useState<RiverData[]>([]);
  const [loadingRivers, setLoadingRivers] = useState(false);
  const [hasRivers, setHasRivers] = useState(false);

  // Pest info state
  const [showPestModal, setShowPestModal] = useState(false);

  // Condition details state
  const [showConditionModal, setShowConditionModal] = useState(false);

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

  const handleRiverClick = async () => {
    setLoadingRivers(true);
    try {
      const response = await fetch(`${API_BASE_URL}/rivers/location/${location.id}`);
      if (response.ok) {
        const data = await response.json();
        setRiverData(data.rivers);
        setShowRiverModal(true);
      }
    } catch (error) {
      console.error('Error fetching river data:', error);
    } finally {
      setLoadingRivers(false);
    }
  };

  // Safely handle potentially null/undefined arrays
  const safeHistorical = historical || [];
  const safeHourly = hourly || [];

  // Combine and deduplicate by timestamp (historical and hourly can overlap)
  const allDataMap = new Map<string, typeof safeHistorical[0]>();
  [...safeHistorical, ...safeHourly].forEach(d => {
    // Use timestamp as key - later entries (hourly/forecast) will overwrite earlier ones
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

  // Check for snow on ground (use recent historical data)
  const recentData = [...safeHistorical].reverse().slice(0, 8); // Last 24 hours
  const snowInfo = ConditionCalculator.calculateSnowProbability(recentData);

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
          {/* Title row with condition badge */}
          <div className="flex items-center justify-between gap-2 mb-1">
            <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white">{location.name}</h2>
            <button
              onClick={() => setShowConditionModal(true)}
              className={`px-2.5 py-0.5 rounded-full text-xs font-semibold border ${conditionBadge.bg} ${conditionBadge.text} ${conditionBadge.border} flex items-center gap-1.5 flex-shrink-0 hover:opacity-80 active:opacity-100 transition-opacity cursor-pointer`}
              title="Click for condition details"
            >
              <div className={`w-2 h-2 rounded-full ${conditionColor}`} />
              <span>{conditionLabel}</span>
            </button>
          </div>
          {/* Date and info icons row */}
          <div className="flex items-center justify-between">
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {format(new Date(current.timestamp), 'MMM d, h:mm a')}
            </p>
            {/* Info Icons */}
            {(pestConditions || hasRivers) && (
              <div className="flex items-center gap-1">
                {/* Pest Activity Icon */}
                {pestConditions && (
                  <button
                    onClick={() => setShowPestModal(true)}
                    className="relative p-1.5 hover:bg-amber-50 dark:hover:bg-amber-900/30 active:bg-amber-100 dark:active:bg-amber-900/50 rounded-full transition-colors"
                    title="Pest Activity Info"
                  >
                    <Bug className={`w-5 h-5 ${getPestLevelColor(pestConditions.mosquitoLevel)}`} />
                    <div className={`absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-gray-800 ${
                      pestConditions.mosquitoScore >= 60 || pestConditions.outdoorPestScore >= 60 ? 'bg-red-500' :
                      pestConditions.mosquitoScore >= 40 || pestConditions.outdoorPestScore >= 40 ? 'bg-yellow-500' :
                      'bg-green-500'
                    }`} />
                  </button>
                )}
                {/* River Crossing Icon */}
                {hasRivers && (
                  <button
                    onClick={handleRiverClick}
                    disabled={loadingRivers}
                    className="relative p-1.5 hover:bg-blue-50 dark:hover:bg-blue-900/30 active:bg-blue-100 dark:active:bg-blue-900/50 rounded-full transition-colors"
                    title="River Crossing Info"
                  >
                    <Waves className={`w-5 h-5 text-blue-600 dark:text-blue-400 ${loadingRivers ? 'animate-pulse' : ''}`} />
                    {riverData.length > 0 && (
                      <div className={`absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-white dark:border-gray-800 ${
                        riverData.every(r => r.is_safe) ? 'bg-green-500' :
                        riverData.some(r => r.status === 'unsafe') ? 'bg-red-500' :
                        'bg-yellow-500'
                      }`} />
                    )}
                  </button>
                )}
              </div>
            )}
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
            <Snowflake className="w-5 h-5 mb-1 text-blue-400" />
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Snow</div>
            <div className="text-sm font-semibold text-gray-900 dark:text-white">
              {snowInfo.probability}
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

        {/* Condition Reasons */}
        <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
          <div className="text-xs text-gray-600 dark:text-gray-400">
            {condition.reasons.join(' • ')}
          </div>
        </div>
      </div>

      {/* Expandable Forecast Section */}
      <button
        onClick={() => onToggleExpand(!isExpanded)}
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

      {/* River Info Modal */}
      {showRiverModal && (
        <RiverInfoModal
          rivers={riverData}
          locationName={location.name}
          onClose={() => setShowRiverModal(false)}
        />
      )}

      {/* Pest Info Modal */}
      {showPestModal && pestConditions && (
        <PestInfoModal
          pestConditions={pestConditions}
          locationName={location.name}
          onClose={() => setShowPestModal(false)}
        />
      )}

      {showConditionModal && (
        <ConditionDetailsModal
          locationName={location.name}
          conditionLevel={condition.level}
          conditionLabel={conditionLabel}
          reasons={condition.reasons}
          onClose={() => setShowConditionModal(false)}
        />
      )}
    </div>
  );
}
