import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { X, TrendingUp, Users, MapPin, Loader2, ExternalLink, Calendar } from 'lucide-react';
import { heatMapApi } from '../../services/api';
import { formatDistanceToNow } from 'date-fns';

interface AreaDetailDrawerProps {
  areaId: number | null;
  dateRange: { start: Date; end: Date };
  isOpen: boolean;
  onClose: () => void;
}

export function AreaDetailDrawer({ areaId, dateRange, isOpen, onClose }: AreaDetailDrawerProps) {
  const { data: detail, isLoading } = useQuery({
    queryKey: ['areaDetail', areaId, dateRange],
    queryFn: () => heatMapApi.getAreaDetail(areaId!, dateRange),
    enabled: !!areaId && isOpen,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

  if (!isOpen) return null;

  return (
    <>
      {/* Overlay */}
      <div
        className="fixed inset-0 bg-black bg-opacity-30 z-40 transition-opacity"
        onClick={onClose}
      />

      {/* Drawer */}
      <div className="fixed right-0 top-0 bottom-0 w-full md:w-96 lg:w-[480px] bg-white dark:bg-gray-800 z-50 shadow-2xl overflow-y-auto transform transition-transform">
        {/* Header */}
        <div className="sticky top-0 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-4 flex items-center justify-between z-10">
          <h2 className="text-xl font-bold text-gray-900 dark:text-white">
            {detail?.name || 'Loading...'}
          </h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            title="Close"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        {isLoading ? (
          <div className="p-8 text-center">
            <Loader2 className="w-12 h-12 animate-spin text-blue-600 dark:text-blue-400 mx-auto mb-4" />
            <p className="text-gray-600 dark:text-gray-400">Loading area details...</p>
          </div>
        ) : detail ? (
          <div className="p-4 space-y-6">
            {/* Stats Cards */}
            <div className="grid grid-cols-3 gap-3">
              <StatCard
                icon={<TrendingUp className="w-5 h-5" />}
                label="Total Ticks"
                value={detail.total_ticks}
              />
              <StatCard
                icon={<MapPin className="w-5 h-5" />}
                label="Routes"
                value={detail.active_routes}
              />
              <StatCard
                icon={<Users className="w-5 h-5" />}
                label="Climbers"
                value={detail.unique_climbers}
              />
            </div>

            {/* Last Activity */}
            <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3">
              <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400 text-sm mb-1">
                <Calendar className="w-4 h-4" />
                <span>Last Activity</span>
              </div>
              <p className="text-gray-900 dark:text-white font-medium">
                {formatDistanceToNow(new Date(detail.last_activity), { addSuffix: true })}
              </p>
            </div>

            {/* Top Routes */}
            {detail.top_routes && detail.top_routes.length > 0 && (
              <div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  Top Routes
                </h3>
                <div className="space-y-2">
                  {detail.top_routes.slice(0, 10).map((route) => (
                    <div
                      key={route.mp_route_id}
                      className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                    >
                      <div className="flex justify-between items-start mb-1">
                        <div className="flex-1">
                          <p className="font-semibold text-sm text-gray-900 dark:text-white">
                            {route.name}
                          </p>
                          <p className="text-xs text-gray-600 dark:text-gray-400">
                            {route.rating}
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="text-sm font-bold text-blue-600 dark:text-blue-400">
                            {route.tick_count} ticks
                          </p>
                          <p className="text-xs text-gray-500 dark:text-gray-400">
                            {formatDistanceToNow(new Date(route.last_activity), { addSuffix: true })}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Recent Ticks */}
            {detail.recent_ticks && detail.recent_ticks.length > 0 && (
              <div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  Recent Ticks
                </h3>
                <div className="space-y-2">
                  {detail.recent_ticks.slice(0, 10).map((tick, index) => (
                    <div
                      key={`${tick.mp_route_id}-${index}`}
                      className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3"
                    >
                      <div className="flex justify-between items-start mb-1">
                        <div>
                          <p className="font-semibold text-sm text-gray-900 dark:text-white">
                            {tick.route_name}
                          </p>
                          <p className="text-xs text-gray-600 dark:text-gray-400">
                            {tick.rating} · {tick.user_name}
                          </p>
                        </div>
                        <span className="text-xs text-gray-500 dark:text-gray-400">
                          {formatDistanceToNow(new Date(tick.climbed_at), { addSuffix: true })}
                        </span>
                      </div>
                      {tick.comment && (
                        <p className="text-xs text-gray-600 dark:text-gray-400 mt-2 italic">
                          "{tick.comment}"
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Recent Comments */}
            {detail.recent_comments && detail.recent_comments.length > 0 && (
              <div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                  Recent Comments
                </h3>
                <div className="space-y-3">
                  {detail.recent_comments.slice(0, 5).map((comment) => (
                    <div
                      key={comment.id}
                      className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3"
                    >
                      <div className="flex justify-between items-start mb-2">
                        <span className="text-sm font-semibold text-gray-900 dark:text-white">
                          {comment.user_name}
                        </span>
                        <span className="text-xs text-gray-500 dark:text-gray-400">
                          {formatDistanceToNow(new Date(comment.commented_at), { addSuffix: true })}
                        </span>
                      </div>
                      <p className="text-sm text-gray-700 dark:text-gray-300 mb-1">
                        {comment.comment_text}
                      </p>
                      {comment.route_name && (
                        <p className="text-xs text-gray-500 dark:text-gray-400">
                          on {comment.route_name}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* External Link */}
            <a
              href={`https://www.mountainproject.com/area/${areaId}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center justify-center gap-2 w-full py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition-colors"
            >
              <span>View on Mountain Project</span>
              <ExternalLink className="w-4 h-4" />
            </a>
          </div>
        ) : null}
      </div>
    </>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  return (
    <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3">
      <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-1">
        {icon}
      </div>
      <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">{label}</div>
      <div className="text-xl font-bold text-gray-900 dark:text-white">
        {value.toLocaleString()}
      </div>
    </div>
  );
}
