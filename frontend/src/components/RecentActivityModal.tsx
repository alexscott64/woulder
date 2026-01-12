import { X, Footprints, ExternalLink, Calendar, User, MapPin } from 'lucide-react';
import { format } from 'date-fns';
import { ClimbHistoryEntry } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';

interface RecentActivityModalProps {
  locationName: string;
  climbHistory: ClimbHistoryEntry[];
  onClose: () => void;
}

// Helper function to clean comments
function cleanComment(comment: string | undefined): string | null {
  if (!comment) return null;

  // Remove all spaces and tabs for pattern matching
  const normalized = comment.replace(/[\s\t]+/g, '');

  // Check if it matches the patterns we want to filter
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

  // Clean the comment: remove · and &middot; prefixes and extra whitespace
  let cleaned = comment
    .replace(/^[\s·•]+/, '')
    .replace(/^&middot;[\s]+/, '')
    .trim();

  return cleaned || null;
}

export function RecentActivityModal({
  locationName,
  climbHistory,
  onClose
}: RecentActivityModalProps) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-2 sm:p-4">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header with subtle gradient */}
        <div className="bg-gradient-to-r from-gray-50 via-blue-50/30 to-gray-50 dark:from-gray-900 dark:via-blue-950/30 dark:to-gray-900 px-4 sm:px-5 py-3 sm:py-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5 min-w-0 flex-1">
              <div className="bg-blue-100 dark:bg-blue-900/40 p-1.5 rounded-lg">
                <Footprints className="w-5 h-5 text-blue-600 dark:text-blue-400 flex-shrink-0" />
              </div>
              <div className="min-w-0 flex-1">
                <h2 className="text-base sm:text-lg font-bold text-gray-900 dark:text-white">Recent Activity</h2>
                <p className="text-xs text-gray-500 dark:text-gray-400 truncate">{locationName}</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex-shrink-0 p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors ml-2"
            >
              <X className="w-5 h-5 text-gray-500 dark:text-gray-400" />
            </button>
          </div>
        </div>

        {/* Timeline Content */}
        <div className="flex-1 overflow-y-auto p-3 sm:p-4">
          <div className="space-y-3">
            {climbHistory.map((climb, index) => {
              // Use helper function to clean comment
              const displayComment = cleanComment(climb.comment);

              // Parse date for display
              const climbDate = new Date(climb.climbed_at);
              const dateStr = format(climbDate, 'MMM d, yyyy');

              return (
                <div
                  key={`${climb.mp_route_id}-${climb.climbed_at}`}
                  className={`bg-white dark:bg-gray-900 border rounded-lg p-3 sm:p-4 hover:shadow-md transition-all ${
                    index === 0
                      ? 'border-blue-200 dark:border-blue-800 shadow-sm'
                      : 'border-gray-200 dark:border-gray-700 hover:border-blue-200 dark:hover:border-blue-800'
                  }`}
                >
                  {/* Route name and grade */}
                  <div className="flex items-start justify-between gap-3 mb-2">
                    <a
                      href={`https://www.mountainproject.com/route/${climb.mp_route_id}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-1.5 group flex-1 min-w-0"
                    >
                      <span className="text-sm sm:text-base font-bold text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-500 transition-colors truncate">
                        {climb.route_name}
                      </span>
                      <ExternalLink className="h-3.5 w-3.5 text-gray-400 group-hover:text-blue-600 dark:group-hover:text-blue-500 flex-shrink-0 transition-colors" />
                    </a>
                    <span className="text-xs sm:text-sm font-bold text-blue-700 dark:text-blue-300 bg-blue-50 dark:bg-blue-900/30 px-2 py-0.5 rounded flex-shrink-0">
                      {climb.route_rating}
                    </span>
                  </div>

                  {/* Area */}
                  <a
                    href={`https://www.mountainproject.com/area/${climb.mp_area_id}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-xs text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-500 transition-colors mb-2 group"
                  >
                    <MapPin className="h-3 w-3" />
                    <span>{climb.area_name}</span>
                    <ExternalLink className="h-2.5 w-2.5 opacity-0 group-hover:opacity-100 transition-opacity" />
                  </a>

                  {/* Metadata */}
                  <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-600 dark:text-gray-400">
                    {climb.climbed_by && climb.climbed_by.trim() !== '' && (
                      <>
                        <div className="flex items-center gap-1">
                          <User className="h-3 w-3" />
                          <span>{climb.climbed_by}</span>
                        </div>
                        <span className="text-gray-400">•</span>
                      </>
                    )}
                    <div className="flex items-center gap-1">
                      <Calendar className="h-3 w-3" />
                      <span>{dateStr}</span>
                    </div>
                    <span className="text-gray-400">•</span>
                    <span>{formatDaysAgo(climb.days_since_climb)}</span>
                    {index === 0 && (
                      <>
                        <span className="text-gray-400">•</span>
                        <span className="text-blue-600 dark:text-blue-400 font-semibold">Latest</span>
                      </>
                    )}
                  </div>

                  {/* Comment */}
                  {displayComment && (
                    <div className="mt-2 pt-2 border-t border-blue-100 dark:border-blue-900/50">
                      <p className="text-xs text-gray-700 dark:text-gray-300 italic">
                        &quot;{displayComment}&quot;
                      </p>
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          {/* Empty state (shouldn't happen, but just in case) */}
          {climbHistory.length === 0 && (
            <div className="text-center py-12">
              <Footprints className="w-16 h-16 mx-auto text-gray-300 dark:text-gray-600 mb-4" />
              <p className="text-gray-500 dark:text-gray-400">No recent activity recorded</p>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-4 py-3 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onClose}
            className="w-full py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 font-medium rounded-lg transition-colors text-sm"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
