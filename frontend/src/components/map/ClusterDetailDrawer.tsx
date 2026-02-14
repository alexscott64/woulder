import React from 'react';
import { X, TrendingUp, Users, MapPin } from 'lucide-react';
import { HeatMapPoint } from '../../types/heatmap';
import { formatDistanceToNow } from 'date-fns';

interface ClusterDetailDrawerProps {
  areas: HeatMapPoint[];
  isOpen: boolean;
  onClose: () => void;
  onAreaClick: (areaId: number) => void;
}

export function ClusterDetailDrawer({ areas, isOpen, onClose, onAreaClick }: ClusterDetailDrawerProps) {
  if (!isOpen || areas.length === 0) return null;

  // Calculate aggregate stats
  const totalTicks = areas.reduce((sum, a) => sum + a.total_ticks, 0);
  const totalRoutes = areas.reduce((sum, a) => sum + a.active_routes, 0);
  const totalClimbers = areas.reduce((sum, a) => sum + a.unique_climbers, 0);
  const totalActivityScore = areas.reduce((sum, a) => sum + a.activity_score, 0);

  // Sort areas by activity score
  const sortedAreas = [...areas].sort((a, b) => b.activity_score - a.activity_score);

  return (
    <>
      {/* Overlay */}
      <div
        className="fixed inset-0 bg-black bg-opacity-30 z-40 transition-opacity md:hidden"
        onClick={onClose}
      />

      {/* Drawer */}
      <div className="fixed right-0 top-0 bottom-0 w-full sm:w-96 lg:w-[480px] bg-white dark:bg-gray-800 z-50 shadow-2xl overflow-y-auto transform transition-transform">
        {/* Header */}
        <div className="sticky top-0 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-3 sm:p-4 z-10">
          <div className="flex items-center justify-between gap-3">
            <h2 className="text-lg sm:text-xl font-bold text-gray-900 dark:text-white">
              {areas.length} Climbing {areas.length === 1 ? 'Area' : 'Areas'} in this Region
            </h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors shrink-0"
              title="Close"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="p-3 sm:p-4 space-y-4 sm:space-y-6">
          {/* Aggregate Stats Cards */}
          <div className="grid grid-cols-2 gap-2 sm:gap-3">
            <StatCard
              icon={<TrendingUp className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Total Activity"
              value={totalActivityScore}
            />
            <StatCard
              icon={<TrendingUp className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Total Ticks"
              value={totalTicks}
            />
            <StatCard
              icon={<MapPin className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Active Routes"
              value={totalRoutes}
            />
            <StatCard
              icon={<Users className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Unique Climbers"
              value={totalClimbers}
            />
          </div>

          {/* Areas List */}
          <div>
            <h3 className="text-base font-semibold text-gray-900 dark:text-white mb-3">
              All Areas in this Cluster
            </h3>
            <div className="space-y-2">
              {sortedAreas.map((area, index) => {
                const daysSince = (Date.now() - new Date(area.last_activity).getTime()) / (1000 * 60 * 60 * 24);
                const colorClass = daysSince <= 7 ? 'bg-red-500'
                  : daysSince <= 30 ? 'bg-orange-500'
                  : daysSince <= 90 ? 'bg-yellow-500'
                  : 'bg-blue-500';

                return (
                  <button
                    key={area.mp_area_id}
                    onClick={() => onAreaClick(area.mp_area_id)}
                    className="w-full bg-gray-50 dark:bg-gray-900 rounded-lg p-3 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-left"
                  >
                    <div className="flex justify-between items-start gap-3">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-xs font-medium text-gray-500 dark:text-gray-400">
                            #{index + 1}
                          </span>
                          <div className={`w-2 h-2 rounded-full ${colorClass}`} />
                          <p className="font-semibold text-sm text-gray-900 dark:text-white truncate">
                            {area.name}
                          </p>
                        </div>
                        <div className="grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-gray-600 dark:text-gray-400 mt-2">
                          <div className="flex justify-between">
                            <span>Score:</span>
                            <span className="font-semibold text-gray-900 dark:text-white">{area.activity_score}</span>
                          </div>
                          <div className="flex justify-between">
                            <span>Ticks:</span>
                            <span className="font-semibold text-gray-900 dark:text-white">{area.total_ticks}</span>
                          </div>
                          <div className="flex justify-between">
                            <span>Routes:</span>
                            <span className="font-semibold text-gray-900 dark:text-white">{area.active_routes}</span>
                          </div>
                          <div className="flex justify-between">
                            <span>Climbers:</span>
                            <span className="font-semibold text-gray-900 dark:text-white">{area.unique_climbers}</span>
                          </div>
                        </div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          Last activity: {formatDistanceToNow(new Date(area.last_activity), { addSuffix: true })}
                        </p>
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  return (
    <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-2 sm:p-3">
      <div className="flex items-center justify-center text-gray-500 dark:text-gray-400 mb-1">
        {icon}
      </div>
      <div className="text-xs text-gray-600 dark:text-gray-400 text-center mb-1">{label}</div>
      <div className="text-lg sm:text-xl font-bold text-gray-900 dark:text-white text-center">
        {value.toLocaleString()}
      </div>
    </div>
  );
}
