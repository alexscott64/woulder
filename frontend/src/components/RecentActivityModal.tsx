import { X, Footprints, ExternalLink, Calendar, User, TrendingUp, MapPin } from 'lucide-react';
import { format } from 'date-fns';
import { ClimbHistoryEntry } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';

interface RecentActivityModalProps {
  locationName: string;
  climbHistory: ClimbHistoryEntry[];
  onClose: () => void;
}

export function RecentActivityModal({
  locationName,
  climbHistory,
  onClose
}: RecentActivityModalProps) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-2 sm:p-4">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-2xl max-w-3xl w-full max-h-[96vh] overflow-hidden flex flex-col">
        {/* Header with gradient */}
        <div className="relative bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 px-4 sm:px-6 py-4 sm:py-5">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3 min-w-0 flex-1">
              <div className="bg-white/20 backdrop-blur-sm rounded-full p-2">
                <Footprints className="w-5 h-5 sm:w-6 sm:h-6 text-white" />
              </div>
              <div className="min-w-0 flex-1">
                <h2 className="text-lg sm:text-xl font-bold text-white">Recent Activity</h2>
                <p className="text-xs sm:text-sm text-blue-100 truncate">{locationName}</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex-shrink-0 p-1.5 hover:bg-white/20 rounded-full transition-colors ml-2"
            >
              <X className="w-5 h-5 sm:w-6 sm:h-6 text-white" />
            </button>
          </div>

          {/* Stats bar */}
          <div className="mt-4 flex items-center gap-4 text-xs sm:text-sm text-white/90">
            <div className="flex items-center gap-1.5">
              <TrendingUp className="w-4 h-4" />
              <span className="font-medium">{climbHistory.length} recent {climbHistory.length === 1 ? 'climb' : 'climbs'}</span>
            </div>
            {climbHistory[0] && (
              <div className="flex items-center gap-1.5">
                <Calendar className="w-4 h-4" />
                <span>Latest: {formatDaysAgo(climbHistory[0].days_since_climb)}</span>
              </div>
            )}
          </div>
        </div>

        {/* Timeline Content */}
        <div className="flex-1 overflow-y-auto p-4 sm:p-6">
          <div className="relative">
            {/* Timeline line */}
            <div className="absolute left-[15px] top-8 bottom-8 w-0.5 bg-gradient-to-b from-blue-200 via-indigo-200 to-purple-200 dark:from-blue-800 dark:via-indigo-800 dark:to-purple-800" />

            {/* Timeline items */}
            <div className="space-y-6">
              {climbHistory.map((climb, index) => {
                // Clean comment: remove " · " prefix and filter out generic comments
                const cleanComment = climb.comment
                  ?.replace(/^[\s·•]+/, '')
                  .trim();

                const shouldShowComment = cleanComment &&
                  !['Send.', 'Flash.', 'Attempt.'].includes(cleanComment.replace(/\s+/g, ''));

                // Parse date for display
                const climbDate = new Date(climb.climbed_at);
                const dateStr = format(climbDate, 'MMM d, yyyy');

                return (
                  <div key={`${climb.mp_route_id}-${climb.climbed_at}`} className="relative pl-12">
                    {/* Timeline dot */}
                    <div className={`absolute left-0 w-8 h-8 rounded-full flex items-center justify-center ${
                      index === 0
                        ? 'bg-gradient-to-br from-blue-500 to-indigo-600 shadow-lg shadow-blue-500/50'
                        : 'bg-gradient-to-br from-gray-200 to-gray-300 dark:from-gray-600 dark:to-gray-700'
                    }`}>
                      <Footprints className={`w-4 h-4 ${index === 0 ? 'text-white' : 'text-gray-600 dark:text-gray-400'}`} />
                    </div>

                    {/* Content card */}
                    <div className={`bg-white dark:bg-gray-900 rounded-xl shadow-md hover:shadow-xl transition-all duration-300 overflow-hidden border ${
                      index === 0
                        ? 'border-blue-200 dark:border-blue-800 ring-2 ring-blue-100 dark:ring-blue-900/50'
                        : 'border-gray-200 dark:border-gray-700'
                    }`}>
                      {/* Card header with route name */}
                      <div className={`px-4 sm:px-5 py-3 sm:py-4 ${
                        index === 0
                          ? 'bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/50 dark:to-indigo-950/50'
                          : 'bg-gray-50 dark:bg-gray-800/50'
                      }`}>
                        <div className="flex items-start justify-between gap-3">
                          <a
                            href={`https://www.mountainproject.com/route/${climb.mp_route_id}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-2 group flex-1 min-w-0"
                          >
                            <span className="text-base sm:text-lg font-bold text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors truncate">
                              {climb.route_name}
                            </span>
                            <ExternalLink className="h-4 w-4 text-blue-500 dark:text-blue-400 flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
                          </a>
                          <span className="text-sm sm:text-base font-bold text-indigo-600 dark:text-indigo-400 flex-shrink-0 px-2.5 py-1 bg-white dark:bg-gray-800 rounded-lg shadow-sm">
                            {climb.route_rating}
                          </span>
                        </div>

                        {/* Area link */}
                        <a
                          href={`https://www.mountainproject.com/area/${climb.mp_area_id}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-1.5 mt-2 text-xs sm:text-sm text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors group"
                        >
                          <MapPin className="h-3.5 w-3.5" />
                          <span>{climb.area_name}</span>
                          <ExternalLink className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                        </a>
                      </div>

                      {/* Card body with metadata */}
                      <div className="px-4 sm:px-5 py-3 sm:py-4 space-y-3">
                        {/* Climber and date info */}
                        <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs sm:text-sm">
                          <div className="flex items-center gap-2 text-gray-700 dark:text-gray-300">
                            <User className="h-4 w-4 text-gray-400 dark:text-gray-500" />
                            <span className="font-medium">{climb.climbed_by}</span>
                          </div>
                          <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
                            <Calendar className="h-4 w-4 text-gray-400 dark:text-gray-500" />
                            <span>{dateStr}</span>
                            <span className="text-gray-400 dark:text-gray-600">•</span>
                            <span className="font-medium">{formatDaysAgo(climb.days_since_climb)}</span>
                          </div>
                          {index === 0 && (
                            <div className="ml-auto">
                              <span className="inline-flex items-center px-2.5 py-1 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-full text-[10px] sm:text-xs font-bold shadow-md">
                                LATEST
                              </span>
                            </div>
                          )}
                        </div>

                        {/* Comment (if meaningful) */}
                        {shouldShowComment && (
                          <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-3 border-l-4 border-blue-400 dark:border-blue-600">
                            <p className="text-xs sm:text-sm text-gray-700 dark:text-gray-300 italic leading-relaxed">
                              &quot;{cleanComment}&quot;
                            </p>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
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
        <div className="px-4 sm:px-6 py-3 sm:py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50">
          <div className="flex items-center justify-between gap-3">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              Data from Mountain Project
            </p>
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 font-medium rounded-lg transition-colors text-sm"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
