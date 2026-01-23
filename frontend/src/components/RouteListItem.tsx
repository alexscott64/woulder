import { ExternalLink, ChevronDown, ChevronUp, Loader2 } from 'lucide-react';
import { RouteActivitySummary, ClimbHistoryEntry } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';
import { useRecentTicksForRoute, useBoulderDryingStatus } from '../hooks/useClimbActivity';
import { BoulderDryingBadge } from './BoulderDryingBadge';
import { DryingForecastTimeline } from './DryingForecastTimeline';

interface RouteListItemProps {
  route: RouteActivitySummary;
  isExpanded: boolean;
  onToggleExpand: () => void;
}

// Helper function to clean comments (reused from RecentActivityModal)
function cleanComment(comment: string | undefined): string | null {
  if (!comment) return null;

  const normalized = comment.replace(/[\s\t]+/g, '');
  const patternsToFilter = [
    '·Send.',
    '·Attempt.',
    '·Flash.',
    'Send.',
    'Attempt.',
    'Flash.'
  ];

  for (const pattern of patternsToFilter) {
    const normalizedPattern = pattern.replace(/[\s\t]+/g, '');
    if (normalized === normalizedPattern || normalized === `&middot;${normalizedPattern}`) {
      return null;
    }
  }

  let cleaned = comment
    .replace(/^[\s·•]+/, '')
    .replace(/^&middot;[\s]+/, '')
    .trim();

  return cleaned || null;
}

// Tick entry component for consistent rendering
function TickEntry({ tick, showDivider = true }: { tick: ClimbHistoryEntry; showDivider?: boolean }) {
  return (
    <div className={`py-1.5 ${showDivider ? 'border-t border-gray-100 dark:border-gray-700' : ''}`}>
      <div className="flex items-start gap-2 text-xs">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5 flex-wrap">
            <span className="font-medium text-gray-900 dark:text-gray-100">
              {tick.climbed_by || 'Anonymous'}
            </span>
            <span className="text-gray-400 dark:text-gray-600">•</span>
            <span className="text-blue-600 dark:text-blue-400 font-medium">
              {tick.style}
            </span>
            <span className="text-gray-400 dark:text-gray-600">•</span>
            <span className="text-gray-600 dark:text-gray-400">
              {formatDaysAgo(tick.days_since_climb)}
            </span>
          </div>
          {cleanComment(tick.comment) && (
            <p className="mt-0.5 text-xs text-gray-600 dark:text-gray-400 italic">
              "{cleanComment(tick.comment)}"
            </p>
          )}
        </div>
      </div>
    </div>
  );
}

export function RouteListItem({ route, isExpanded, onToggleExpand }: RouteListItemProps) {
  const mostRecentTick = route.most_recent_tick;
  const hasBeenClimbed = !!mostRecentTick;

  // Fetch recent ticks when expanded
  const { data: recentTicks, isLoading } = useRecentTicksForRoute(
    isExpanded ? route.mp_route_id : null,
    5
  );

  // Fetch boulder drying status immediately for compact badge
  const { data: dryingStatus, isLoading: isDryingLoading } = useBoulderDryingStatus(
    route.mp_route_id
  );

  return (
    <div className="bg-white dark:bg-gray-800 rounded-md border border-gray-200 dark:border-gray-700 p-2.5">
      {/* Route Header */}
      <div className="flex items-start gap-2">
        <div className="flex-1 min-w-0">
          {/* Route Name and Rating */}
          <div className="flex items-center gap-1.5 flex-wrap">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
              {route.name}
            </h3>
            <span className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-mono font-semibold bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300">
              {route.rating}
            </span>

            {/* Compact Boulder Drying Badge */}
            {dryingStatus && !isDryingLoading && (
              <BoulderDryingBadge status={dryingStatus} compact={true} />
            )}

            <a
              href={`https://www.mountainproject.com/route/${route.mp_route_id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
              onClick={(e) => e.stopPropagation()}
            >
              <ExternalLink className="w-3.5 h-3.5" />
            </a>
          </div>

          {/* Most Recent Tick or No Activity */}
          {hasBeenClimbed ? (
            <div className="mt-1.5">
              <div className="text-xs text-gray-500 dark:text-gray-500 mb-0.5">
                Most recent:
              </div>
              <TickEntry tick={mostRecentTick} showDivider={false} />
            </div>
          ) : (
            <div className="mt-1.5 text-xs text-gray-500 dark:text-gray-500 italic">
              No recorded ascents
            </div>
          )}
        </div>
      </div>

      {/* Expand/Collapse Button - Only show if route has been climbed */}
      {hasBeenClimbed && (
        <button
          onClick={onToggleExpand}
          className="mt-3 w-full flex items-center justify-center gap-1.5 px-3 py-2 text-xs sm:text-sm text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors"
        >
          {isExpanded ? (
            <>
              <ChevronUp className="w-4 h-4" />
              <span>Show less</span>
            </>
          ) : (
            <>
              <ChevronDown className="w-4 h-4" />
              <span>Show last 5 ascents</span>
            </>
          )}
        </button>
      )}

      {/* Expanded Content */}
      {isExpanded && (
        <div className="mt-3 space-y-3">
          {/* Boulder Drying Status */}
          {isDryingLoading ? (
            <div className="flex items-center justify-center py-4">
              <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />
            </div>
          ) : dryingStatus && (
            <>
              <BoulderDryingBadge status={dryingStatus} />

              {/* 6-Day Drying Forecast */}
              {dryingStatus.forecast && dryingStatus.forecast.length > 0 && (
                <div className="pt-3 border-t border-gray-200 dark:border-gray-700">
                  <DryingForecastTimeline forecast={dryingStatus.forecast} />
                </div>
              )}
            </>
          )}

          {/* Recent Ticks */}
          <div className="border-t border-gray-200 dark:border-gray-700">
            {isLoading ? (
              <div className="flex items-center justify-center py-4">
                <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />
              </div>
            ) : recentTicks && recentTicks.length > 0 ? (
              <div className="space-y-0">
                {recentTicks.map((tick, index) => (
                  <TickEntry key={index} tick={tick} />
                ))}
              </div>
            ) : (
              <div className="py-3 text-xs text-gray-500 dark:text-gray-500 text-center italic">
                No recent ascents found
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
