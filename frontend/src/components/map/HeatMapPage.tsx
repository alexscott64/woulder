import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { heatMapApi } from '../../services/api';
import { HeatMapPoint } from '../../types/heatmap';
import { Calendar, Activity, Loader2, AlertCircle, TrendingUp, Users, Map as MapIcon, List } from 'lucide-react';
import { ActivityMapDeckGL } from './ActivityMapDeckGL';
import { AreaDetailDrawer } from './AreaDetailDrawer';
import { ClusterDetailDrawer } from './ClusterDetailDrawer';
import { RouteTypeFilter } from './RouteTypeFilter';

type ViewMode = 'map' | 'list';

export function HeatMapPage() {
  const [dateRange, setDateRange] = useState({
    start: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000), // Last 90 days
    end: new Date(),
  });
  const [minActivity, setMinActivity] = useState(5);
  const [viewMode, setViewMode] = useState<ViewMode>('map');
  const [selectedAreaId, setSelectedAreaId] = useState<number | null>(null);
  const [selectedRouteTypes, setSelectedRouteTypes] = useState<string[]>(['Boulder', 'Sport', 'Trad', 'Ice']);
  const [clusterAreas, setClusterAreas] = useState<HeatMapPoint[]>([]);
  const [showClusterDrawer, setShowClusterDrawer] = useState(false);

  // Fetch heat map data - full mode to get active_routes
  const { data, isLoading, error } = useQuery({
    queryKey: ['heatMap', dateRange, minActivity, selectedRouteTypes],
    queryFn: () => heatMapApi.getHeatMapActivity({
      startDate: dateRange.start,
      endDate: dateRange.end,
      minActivity,
      limit: 10000, // No effective limit - show all areas
      routeTypes: selectedRouteTypes.length > 0 ? selectedRouteTypes : undefined,
      lightweight: false, // Full mode to get active_routes and other stats
    }),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const handlePreset = (days: number) => {
    const end = new Date();
    const start = new Date(Date.now() - days * 24 * 60 * 60 * 1000);
    setDateRange({ start, end });
  };

  // Sort points by activity score
  const sortedPoints: HeatMapPoint[] = data?.points.sort((a, b) => b.activity_score - a.activity_score) || [];

  return (
    <main className="h-screen flex flex-col overflow-hidden">
      <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-4 space-y-4 flex-shrink-0">
        {/* Header */}
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            Climbing Activity Across North America
          </h2>
          <p className="text-gray-600 dark:text-gray-400">
            Discover where climbers are actively climbing based on Mountain Project tick data
          </p>
        </div>

        {/* Filters */}
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <div className="space-y-4">
            {/* Route Type Filter */}
            <RouteTypeFilter
              selectedTypes={selectedRouteTypes}
              onChange={setSelectedRouteTypes}
            />

            {/* Date Range and Activity Filters */}
            <div className="flex flex-wrap gap-4 items-center pt-4 border-t border-gray-200 dark:border-gray-700">
              {/* Date Range Presets */}
              <div className="flex items-center gap-2 flex-wrap">
                <Calendar className="w-5 h-5 text-gray-500" />
                <span className="text-sm font-medium text-gray-700 dark:text-gray-300 hidden sm:inline">Time Period:</span>
                {[
                  { label: 'Week', days: 7 },
                  { label: 'Month', days: 30 },
                  { label: '3 Months', days: 90 },
                  { label: 'Year', days: 365 },
                ].map((preset) => {
                  const isActive = Math.abs(dateRange.end.getTime() - dateRange.start.getTime()) / (1000 * 60 * 60 * 24) === preset.days;
                  return (
                    <button
                      key={preset.label}
                      onClick={() => handlePreset(preset.days)}
                      className={`px-3 py-1.5 text-xs sm:text-sm rounded-lg transition-colors ${
                        isActive
                          ? 'bg-blue-600 text-white'
                          : 'border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300'
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
                <label className="text-xs sm:text-sm text-gray-700 dark:text-gray-300 hidden sm:inline">
                  Min:
                </label>
                <input
                  type="range"
                  min="1"
                  max="50"
                  value={minActivity}
                  onChange={(e) => setMinActivity(Number(e.target.value))}
                  className="w-20 sm:w-32"
                />
                <span className="text-sm font-medium text-gray-900 dark:text-white w-8">
                  {minActivity}
                </span>
              </div>

              {/* View Toggle */}
              <div className="flex items-center gap-1 border border-gray-300 dark:border-gray-600 rounded-lg p-1">
                <button
                  onClick={() => setViewMode('map')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded transition-colors ${
                    viewMode === 'map'
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                  title="Map View"
                >
                  <MapIcon className="w-4 h-4" />
                  <span className="text-sm hidden sm:inline">Map</span>
                </button>
                <button
                  onClick={() => setViewMode('list')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded transition-colors ${
                    viewMode === 'list'
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                  title="List View"
                >
                  <List className="w-4 h-4" />
                  <span className="text-sm hidden sm:inline">List</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Summary Stats - Only show in list view */}
        {viewMode === 'list' && data && sortedPoints.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-2">
                <TrendingUp className="w-5 h-5" />
                <span className="text-sm">Total {data.count === 1 ? 'Area' : 'Areas'}</span>
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {data.count.toLocaleString()}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-2">
                <Activity className="w-5 h-5" />
                <span className="text-sm">Total {sortedPoints.reduce((sum: number, p: HeatMapPoint) => sum + p.total_ticks, 0) === 1 ? 'Tick' : 'Ticks'}</span>
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {sortedPoints.reduce((sum: number, p: HeatMapPoint) => sum + p.total_ticks, 0).toLocaleString()}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-2">
                <Users className="w-5 h-5" />
                <span className="text-sm">Unique {sortedPoints.reduce((sum: number, p: HeatMapPoint) => sum + p.unique_climbers, 0) === 1 ? 'Climber' : 'Climbers'}</span>
              </div>
              <div className="text-3xl font-bold text-gray-900 dark:text-white">
                {sortedPoints.reduce((sum: number, p: HeatMapPoint) => sum + p.unique_climbers, 0).toLocaleString()}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center flex-1">
          <div className="text-center">
            <Loader2 className="w-12 h-12 animate-spin text-blue-600 dark:text-blue-400 mx-auto mb-4" />
            <p className="text-gray-700 dark:text-gray-300 font-medium">Loading activity data...</p>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8">
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
        </div>
      )}

      {/* Map View */}
      {!isLoading && !error && viewMode === 'map' && sortedPoints.length > 0 && (
        <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 pb-4 flex-1">
          <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden h-full relative">
            <ActivityMapDeckGL
              points={sortedPoints}
              onAreaClick={(areaId) => {
                setShowClusterDrawer(false);
                setSelectedAreaId(areaId);
              }}
              selectedAreaId={selectedAreaId}
              onShowCluster={(areas) => {
                setClusterAreas(areas);
                setShowClusterDrawer(true);
              }}
            />
          </div>
          
          {/* Cluster Drawer */}
          <ClusterDetailDrawer
            areas={clusterAreas}
            isOpen={showClusterDrawer && !selectedAreaId}
            onClose={() => {
              setShowClusterDrawer(false);
              setClusterAreas([]);
            }}
            onAreaClick={(areaId: number) => {
              setSelectedAreaId(areaId);
            }}
          />
          
          {/* Area Detail Drawer */}
          <AreaDetailDrawer
            areaId={selectedAreaId}
            dateRange={dateRange}
            isOpen={!!selectedAreaId}
            onClose={() => {
              setSelectedAreaId(null);
              if (!showClusterDrawer) {
                setClusterAreas([]);
              }
            }}
            onBack={clusterAreas.length > 0 ? () => {
              setSelectedAreaId(null);
              setShowClusterDrawer(true);
            } : undefined}
          />
        </div>
      )}

      {/* List View */}
      {!isLoading && !error && viewMode === 'list' && sortedPoints.length > 0 && (
        <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 pb-4 overflow-y-auto">
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
              {sortedPoints.slice(0, 50).map((point: HeatMapPoint, index: number) => {
                const daysSince = Math.floor(
                  (new Date().getTime() - new Date(point.last_activity).getTime()) / (1000 * 60 * 60 * 24)
                );
                
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
                    onClick={() => {
                      setSelectedAreaId(point.mp_area_id);
                      setViewMode('map');
                    }}
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
                          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Score</div>
                          <div className="text-lg font-bold text-blue-600 dark:text-blue-400">
                            {point.activity_score}
                          </div>
                        </div>
                        <div>
                          <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">{point.total_ticks === 1 ? 'Tick' : 'Ticks'}</div>
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
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && sortedPoints.length === 0 && (
        <div className="flex items-center justify-center flex-1">
          <div className="text-center">
            <Activity className="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <p className="text-gray-700 dark:text-gray-300 font-medium mb-2">No activity found</p>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Try adjusting your filters or time period
            </p>
          </div>
        </div>
      )}
    </main>
  );
}
