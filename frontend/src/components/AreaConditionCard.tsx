import React from 'react';
import { AreaDryingStats } from '../types/weather';
import { Droplets, Sun, Clock } from 'lucide-react';

interface AreaConditionCardProps {
  areaName: string;
  stats?: AreaDryingStats;
  isLoading?: boolean;
}

export const AreaConditionCard: React.FC<AreaConditionCardProps> = ({
  stats,
  isLoading = false,
}) => {
  // Skeleton loader with fixed dimensions
  if (isLoading || !stats) {
    return (
      <div className="rounded-lg p-3 bg-gray-200 dark:bg-gray-700 animate-pulse" style={{ minHeight: '80px' }}>
        <div className="space-y-2">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-8 bg-gray-300 dark:bg-gray-600 rounded" />
              <div className="w-32 h-4 bg-gray-300 dark:bg-gray-600 rounded" />
            </div>
            <div className="flex items-center gap-4">
              <div className="w-8 h-4 bg-gray-300 dark:bg-gray-600 rounded" />
              <div className="w-8 h-4 bg-gray-300 dark:bg-gray-600 rounded" />
              <div className="w-8 h-4 bg-gray-300 dark:bg-gray-600 rounded" />
            </div>
          </div>
          <div className="w-24 h-3 bg-gray-300 dark:bg-gray-600 rounded" />
        </div>
      </div>
    );
  }
  const getStatusColor = (): string => {
    if (stats.percent_dry >= 80) return 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300';
    if (stats.percent_dry >= 50) return 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-300';
    return 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300';
  };

  // Format last rain timestamp
  const formatLastRain = (timestamp: string): string => {
    try {
      const date = new Date(timestamp);
      if (isNaN(date.getTime())) return 'Unknown';

      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

      if (diffDays === 0) return 'Today';
      if (diffDays === 1) return 'Yesterday';
      if (diffDays < 7) return `${diffDays}d ago`;
      return date.toLocaleDateString();
    } catch {
      return 'Unknown';
    }
  };

  return (
    <div className={`rounded-lg p-3 ${getStatusColor()}`}>
      <div className="space-y-2">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div className="flex items-center gap-3">
            <div className="text-2xl font-bold">
              {Math.round(stats.percent_dry)}%
            </div>
            <div className="text-sm font-medium">
              of {stats.total_routes} route{stats.total_routes !== 1 ? 's' : ''} dry
            </div>
          </div>

          <div className="flex items-center gap-4 text-sm">
            {stats.dry_count > 0 && (
              <div className="flex items-center gap-1">
                <Sun className="w-3.5 h-3.5" />
                <span className="font-medium">{stats.dry_count}</span>
              </div>
            )}
            {stats.drying_count > 0 && (
              <div className="flex items-center gap-1">
                <Clock className="w-3.5 h-3.5" />
                <span className="font-medium">{stats.drying_count}</span>
              </div>
            )}
            {stats.wet_count > 0 && (
              <div className="flex items-center gap-1">
                <Droplets className="w-3.5 h-3.5" />
                <span className="font-medium">{stats.wet_count}</span>
              </div>
            )}
          </div>
        </div>

        {/* Last Rain */}
        {stats.last_rain_timestamp && (
          <div className="text-xs opacity-75">
            Last rain: {formatLastRain(stats.last_rain_timestamp)}
          </div>
        )}
      </div>
    </div>
  );
};
