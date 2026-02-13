import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { heatMapApi } from '../../services/api';
import { Calendar, Activity, Loader2, AlertCircle, TrendingUp, Users } from 'lucide-react';

export function HeatMapPage() {
  const [dateRange, setDateRange] = useState({
    start: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000), // Last 90 days
    end: new Date(),
  });
  const [minActivity, setMinActivity] = useState(5);

  // Fetch heat map data
  const { data, isLoading, error } = useQuery({
    queryKey: ['heatMap', dateRange, minActivity],
    queryFn: () => heatMapApi.getHeatMapActivity({
      startDate: dateRange.start,
      endDate: dateRange.end,
      minActivity,
      limit: 100,
    }),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const handlePreset = (days: number) => {
    const end = new Date();
    const start = new Date(Date.now() - days * 24 * 60 * 60 * 1000);
    setDateRange({ start, end });
  };

  // Sort points by activity score
  const sortedPoints = data?.points.sort((a, b) => b.activity_score - a.activity_score) || [];

  return (
    <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Climbing Activity Across North America
        </h2>
        <p className="text-gray-600 dark:text-gray-400">
          Discover where climbers are actively climbing based on Mountain Project tick data
        </p>
      </div>

      {/* Filters */}
      <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4 mb-6">
        <div className="flex flex-wrap gap-4 items-center">
          {/* Date Range Presets */}
          <div className="flex items-center gap-2">
            <Calendar className="w-5 h-5 text-gray-500" />
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Time Period:</span>
            {[
              { label: 'Last Week', days: 7 },
              { label: 'Last Month', days: 30 },
              { label: 'Last 3 Months', days: 90 },
              { label: 'Last Year', days: 365 },
            ].map((preset) => {
              const isActive = Math.abs(dateRange.end.getTime() - dateRange.start.getTime()) / (1000 * 60 * 60 * 24) === preset.days;
              return (
                <button
                  key={preset.label}
                  onClick={() => handlePreset(preset.days)}
                  className={`px-3 py-1.5 text-xs sm:text-sm rounded-lg transition-colors ${
                    isActive
                      ? 'bg-blue-600 text-white'
                      : 'border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  {preset.label}
                </button>
              );
            })}
          </div>

          {/* Activity Threshold */}
          <div className="flex items-center gap-2 ml-auto">
            <Activity className="w-5 h-5 text-gray-500" />
            <label className="text-xs sm:text-sm text-gray-700 dark:text-gray-300">
              Min. Activity:
            </label>
            <input
              type="range"
              min="1"
              max="50"
              value={minActivity}
              onChange={(e) => setMinActivity(Number(e.target.value))}
              className="w-24 sm:w-32"
            />
            <span className="text-sm font-medium text-gray-900 dark:text-white w-8">
              {minActivity}
            </span>
          </div>
        </div>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <Loader2 className="w-12 h-12 animate-spin text-blue-600 dark:text-blue-400 mx-auto mb-4" />
            <p className="text-gray-700 dark:text-gray-300 font-medium">Loading activity data...</p>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6">
          <div className="flex items-center gap-3">
            <AlertCircle className="w-6 h-6 text-red-600 dark:text-red-400" />
            <div>
              <p className="text-red-900 dark:text-red-200 font-medium mb-1">
                Failed to load activity data
              </p>
              <p className="text-red-700 dark:text-red-300 text-sm">
                {error instanceof Error ? error.message : 'Please try again later'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Activity Grid */}
      {data && sortedPoints.length > 0 && (
        <>
          {/* Summary Stats */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-2">
                <TrendingUp className="w-5 h-5" />
                <span className="text-sm">Total Areas</span>
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {data.count}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-2">
                <Activity className="w-5 h-5" />
                <span className="text-sm">Total Ticks</span>
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {sortedPoints.reduce((sum, p) => sum + p.total_ticks, 0).toLocaleString()}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-2">
                <Users className="w-5 h-5" />
                <span className="text-sm">Total Climbers</span>
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {sortedPoints.reduce((sum, p) => sum + p.unique_climbers, 0).toLocaleString()}
              </div>
            </div>
          </div>

          {/* Areas List */}
          <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-bold text-gray-900 dark:text-white">
                Most Active Areas
              </h3>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                Showing top {sortedPoints.length} areas with recent climbing activity
              </p>
            </div>

            <div className="divide-y divide-gray-200 dark:divide-gray-700">
              {sortedPoints.slice(0, 50).map((point, index) => {
                const daysSince = Math.floor(
                  (new Date().getTime() - new Date(point.last_activity).getTime()) / (1000 * 60 * 60 * 24)
                );
                
                // Color based on recency
                const getRecencyColor = () => {
                  if (daysSince <= 7) return 'bg-red-500';
                  if (daysSince <= 30) return 'bg-orange-500';
                  if (daysSince <= 90) return 'bg-yellow-500';
                  return 'bg-blue-500';
                };

                return (
                  <div
                    key={point.mp_area_id}
                    className="px-6 py-4 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors cursor-pointer"
                  >
                    <div className="flex items-start justify-between gap-4">
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-2">
                          <span className="text-sm font-medium text-gray-500 dark:text-gray-400 w-8">
                            #{index + 1}
                          </span>
                          <div>
                            <h4 className="text-base font-semibold text-gray-900 dark:text-white">
                              {point.name}
                            </h4>
                            <div className="flex items-center gap-2 mt-1 text-sm text-gray-600 dark:text-gray-400">
                              <span className={`w-2 h-2 rounded-full ${getRecencyColor()}`} />
                              <span>
                                {daysSince === 0 ? 'Today' : 
                                 daysSince === 1 ? 'Yesterday' : 
                                 `${daysSince} days ago`}
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>

                      <div className="grid grid-cols-3 gap-4 text-right">
                        <div>
                          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Activity Score</div>
                          <div className="text-lg font-bold text-blue-600 dark:text-blue-400">
                            {point.activity_score}
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Total Ticks</div>
                          <div className="text-lg font-bold text-gray-900 dark:text-white">
                            {point.total_ticks}
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Climbers</div>
                          <div className="text-lg font-bold text-gray-900 dark:text-white">
                            {point.unique_climbers}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </>
      )}

      {/* Empty State */}
      {!isLoading && !error && sortedPoints.length === 0 && (
        <div className="text-center py-12">
          <Activity className="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
          <p className="text-gray-700 dark:text-gray-300 font-medium mb-2">No activity found</p>
          <p className="text-gray-600 dark:text-gray-400 text-sm">
            Try adjusting your filters or time period
          </p>
        </div>
      )}
    </main>
  );
}
