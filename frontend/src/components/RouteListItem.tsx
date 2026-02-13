import { ExternalLink, ChevronDown, ChevronUp, Droplets, Sun, Clock, MapPin, TreePine, Compass, User } from 'lucide-react';
import { RouteActivitySummary, ClimbHistoryEntry, BoulderDryingStatus } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';
import { useRecentTicksForRoute, useBoulderDryingStatus } from '../hooks/useClimbActivity';
import { DryingForecastTimeline } from './DryingForecastTimeline';

interface RouteListItemProps {
  route: RouteActivitySummary;
  isExpanded: boolean;
  onToggleExpand: () => void;
  dryingStatus?: BoulderDryingStatus; // Optional: can be provided from parent batch fetch
  useBatchMode?: boolean; // If true, never fetch individually (wait for batch)
}

// Helper function to clean comments
const cleanComment = (comment: string | undefined): string => {
  if (!comment) return '';

  // Decode HTML entities
  const decoded = comment
    .replace(/&middot;/g, '')
    .replace(/&amp;/g, '&')
    .replace(/&quot;/g, '"')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>');

  // Remove leading middot character (·) and other prefixes
  let cleaned = decoded
    .replace(/^·\s*/g, '')
    .replace(/^Sent!\s*/i, '');

  // Remove leading emojis
  cleaned = cleaned.replace(/^[\u{1F300}-\u{1F9FF}]\s*/u, '');
  cleaned = cleaned.replace(/^[\u{1F300}-\u{1F9FF}\s]+/u, '');

  return cleaned.trim();
};

export function RouteListItem({ route, isExpanded, onToggleExpand, dryingStatus: propDryingStatus, useBatchMode = false }: RouteListItemProps) {
  const { data: recentTicks, isLoading: isLoadingTicks } = useRecentTicksForRoute(route.mp_route_id);

  // CRITICAL: Never fetch individually when in batch mode
  // This prevents stale cached data from individual queries
  const shouldFetchIndividually = !useBatchMode && propDryingStatus === undefined;
  const { data: fetchedDryingStatus } = useBoulderDryingStatus(shouldFetchIndividually ? route.mp_route_id : null);

  // Use prop data if in batch mode, otherwise fall back to individual fetch
  const dryingStatus = useBatchMode ? propDryingStatus : (propDryingStatus || fetchedDryingStatus);

  // Debug logging for forecast data
  if (dryingStatus && isExpanded) {
    console.log(`[FORECAST DEBUG] Route ${route.mp_route_id}:`, {
      hasForecast: !!dryingStatus.forecast,
      forecastLength: dryingStatus.forecast?.length || 0,
      forecast: dryingStatus.forecast
    });
  }

  // Check if rain is forecasted in the next 48 hours
  const hasUpcomingRain = () => {
    if (!dryingStatus?.forecast) return false;
    const now = new Date();
    const next48h = new Date(now.getTime() + 48 * 60 * 60 * 1000);

    return dryingStatus.forecast.some(period => {
      const start = new Date(period.start_time);
      return !period.is_dry && start <= next48h && (period.rain_amount || 0) > 0.01;
    });
  };

  // Get compact dry status badge with skeleton loader
  const getDryStatusBadge = () => {
    if (!dryingStatus) {
      // Skeleton loader with fixed dimensions to prevent layout shift
      return (
        <div className="flex items-center gap-1 px-2 py-0.5 rounded bg-gray-200 dark:bg-gray-700 animate-pulse" style={{ width: '60px', height: '24px' }}>
          <div className="w-3 h-3 rounded-full bg-gray-300 dark:bg-gray-600" />
          <div className="flex-1 h-2 bg-gray-300 dark:bg-gray-600 rounded" />
        </div>
      );
    }

    if (!dryingStatus.is_wet) {
      const upcoming = hasUpcomingRain();
      return (
        <div className={`flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium ${
          upcoming
            ? 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-400'
            : 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400'
        }`}>
          <Sun className="w-3 h-3" />
          <span>{upcoming ? 'Dry (rain soon)' : 'Dry'}</span>
        </div>
      );
    }

    const hours = Math.round(dryingStatus.hours_until_dry);

    if (dryingStatus.status === 'critical') {
      return (
        <div className="flex items-center gap-1 px-2 py-0.5 rounded bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 text-xs font-medium">
          <Droplets className="w-3 h-3" />
          <span>Critical</span>
        </div>
      );
    }

    if (hours < 24) {
      return (
        <div className="flex items-center gap-1 px-2 py-0.5 rounded bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-400 text-xs font-medium">
          <Clock className="w-3 h-3" />
          <span>{hours}h</span>
        </div>
      );
    }

    const days = Math.floor(hours / 24);
    return (
      <div className="flex items-center gap-1 px-2 py-0.5 rounded bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400 text-xs font-medium">
        <Droplets className="w-3 h-3" />
        <span>{days}d</span>
      </div>
    );
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors">
      {/* Main Route Info - Always visible */}
      <button
        onClick={onToggleExpand}
        className="w-full p-3 text-left"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            {/* Route Name & Rating */}
            <div className="flex items-baseline gap-2 mb-1.5">
              <h3 className="text-base font-bold text-gray-900 dark:text-gray-100 truncate">
                {route.name}
              </h3>
              <span className="text-sm font-semibold text-blue-600 dark:text-blue-400 flex-shrink-0">
                {route.rating}
              </span>
            </div>

            {/* Dry Status + Activity Summary */}
            <div className="flex items-center gap-2 flex-wrap">
              {getDryStatusBadge()}

              <div className="text-xs text-gray-600 dark:text-gray-400">
                {formatDaysAgo(route.days_since_climb)}
              </div>
            </div>
          </div>

          {/* Right Side Icons */}
          <div className="flex items-center gap-1 flex-shrink-0">
            <a
              href={`https://www.mountainproject.com/route/${route.mp_route_id}`}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
              className="p-1.5 text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors rounded"
              title="View on Mountain Project"
            >
              <ExternalLink className="w-4 h-4" />
            </a>
            {isExpanded ? (
              <ChevronUp className="w-5 h-5 text-gray-400" />
            ) : (
              <ChevronDown className="w-5 h-5 text-gray-400" />
            )}
          </div>
        </div>
      </button>

      {/* Expanded Content - Organized Sections */}
      {isExpanded && (
        <div className="border-t border-gray-200 dark:border-gray-700">
          {/* Boulder Details Section */}
          {dryingStatus && (
            <div className="p-3 bg-gray-50 dark:bg-gray-900/50">
              <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                Boulder Conditions
              </h4>

              {/* Status Message */}
              <div className="mb-3 text-sm text-gray-600 dark:text-gray-400">
                {dryingStatus.message}
              </div>

              {/* Details Grid */}
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div className="flex items-center gap-1.5 text-gray-600 dark:text-gray-400">
                  <MapPin className="w-3.5 h-3.5" />
                  <span>{dryingStatus.latitude.toFixed(4)}, {dryingStatus.longitude.toFixed(4)}</span>
                </div>
                <div className="flex items-center gap-1.5 text-gray-600 dark:text-gray-400">
                  <Compass className="w-3.5 h-3.5" />
                  <span>{dryingStatus.aspect} facing</span>
                </div>
                <div className="flex items-center gap-1.5 text-gray-600 dark:text-gray-400">
                  <TreePine className="w-3.5 h-3.5" />
                  <span>{Math.round(dryingStatus.tree_coverage_percent)}% tree cover</span>
                </div>
                <div className="flex items-center gap-1.5 text-gray-600 dark:text-gray-400">
                  <Sun className="w-3.5 h-3.5" />
                  <span>{Math.round(dryingStatus.sun_exposure_hours)}h sun (6d)</span>
                </div>
              </div>

              {/* Last Rain */}
              {dryingStatus.last_rain_timestamp && (
                <div className="mt-2 text-xs text-gray-500 dark:text-gray-500">
                  Last rain: {new Date(dryingStatus.last_rain_timestamp).toLocaleDateString()} ({Math.round((Date.now() - new Date(dryingStatus.last_rain_timestamp).getTime()) / (1000 * 60 * 60 * 24))}d ago)
                </div>
              )}

              {/* Confidence Score */}
              <div className="mt-2 flex items-center gap-1 text-xs">
                <span className="text-gray-500 dark:text-gray-500">Confidence:</span>
                <div className="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                  <div
                    className={`h-full rounded-full ${
                      dryingStatus.confidence_score >= 80 ? 'bg-green-500' :
                      dryingStatus.confidence_score >= 60 ? 'bg-yellow-500' :
                      'bg-orange-500'
                    }`}
                    style={{ width: `${dryingStatus.confidence_score}%` }}
                  />
                </div>
                <span className="text-gray-600 dark:text-gray-400">{dryingStatus.confidence_score}%</span>
              </div>
            </div>
          )}

          {/* 6-Day Forecast */}
          {dryingStatus?.forecast && dryingStatus.forecast.length > 0 && (
            <div className="p-3 border-t border-gray-200 dark:border-gray-700">
              <DryingForecastTimeline forecast={dryingStatus.forecast} />
            </div>
          )}

          {/* Recent Ascents */}
          <div className="p-3 border-t border-gray-200 dark:border-gray-700">
            <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              Recent Ascents
            </h4>
            {isLoadingTicks ? (
              <div className="text-center text-sm text-gray-500 dark:text-gray-500 py-2">
                Loading...
              </div>
            ) : recentTicks && recentTicks.length > 0 ? (
              <div className="space-y-2">
                {recentTicks.slice(0, 5).map((tick: ClimbHistoryEntry, index: number) => {
                  const tickComment = cleanComment(tick.comment);
                  return (
                    <div
                      key={index}
                      className="text-sm border-l-2 border-blue-500 pl-2"
                    >
                      <div className="flex items-baseline gap-2">
                        <User className="w-3 h-3 text-gray-400 flex-shrink-0 mt-0.5" />
                        <span className="font-medium text-gray-900 dark:text-gray-100">
                          {tick.climbed_by}
                        </span>
                        {tick.style && (
                          <span className="text-xs text-gray-500 dark:text-gray-500">
                            {tick.style}
                          </span>
                        )}
                        <span className="text-xs text-gray-500 dark:text-gray-500 ml-auto">
                          {formatDaysAgo(tick.days_since_climb)}
                        </span>
                      </div>
                      {tickComment && (
                        <p className="text-xs text-gray-600 dark:text-gray-400 mt-0.5 italic">
                          "{tickComment}"
                        </p>
                      )}
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="text-center text-sm text-gray-500 dark:text-gray-500 py-2">
                No recent ascents found
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
