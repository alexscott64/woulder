import React from 'react';
import { AreaDryingStats } from '../types/weather';
import { Mountain, Droplets, Sun, Clock, TreeDeciduous } from 'lucide-react';

interface AreaConditionCardProps {
  areaName: string;
  stats: AreaDryingStats;
  onClick?: () => void;
  className?: string;
}

export const AreaConditionCard: React.FC<AreaConditionCardProps> = ({
  areaName,
  stats,
  onClick,
  className = '',
}) => {
  const getStatusColor = (): string => {
    if (stats.percent_dry >= 80) return 'border-green-500 bg-green-50 dark:bg-green-900/20';
    if (stats.percent_dry >= 50) return 'border-yellow-500 bg-yellow-50 dark:bg-yellow-900/20';
    return 'border-red-500 bg-red-50 dark:bg-red-900/20';
  };

  const getStatusLabel = (): string => {
    if (stats.percent_dry >= 80) return 'Mostly Dry';
    if (stats.percent_dry >= 50) return 'Mixed Conditions';
    return 'Mostly Wet';
  };

  const formatHours = (hours: number): string => {
    if (hours === 0) return 'Dry';
    if (hours < 1) return '<1h';
    if (hours < 24) return `${Math.round(hours)}h`;
    const days = Math.floor(hours / 24);
    const remainingHours = Math.round(hours % 24);
    return remainingHours > 0 ? `${days}d ${remainingHours}h` : `${days}d`;
  };

  return (
    <div
      className={`border-2 rounded-lg p-4 transition-all duration-200 ${getStatusColor()} ${
        onClick ? 'cursor-pointer hover:shadow-md' : ''
      } ${className}`}
      onClick={onClick}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <Mountain className="w-5 h-5 text-gray-700 dark:text-gray-300" />
          <h3 className="font-semibold text-gray-900 dark:text-white">{areaName}</h3>
        </div>
        <div className="flex items-center gap-1 px-2 py-1 bg-white dark:bg-gray-800 rounded text-xs font-medium">
          <span className="text-gray-600 dark:text-gray-400">
            {stats.confidence_score}%
          </span>
        </div>
      </div>

      {/* Status badge */}
      <div className="mb-3">
        <span className="inline-flex items-center px-2 py-1 bg-white dark:bg-gray-800 rounded text-sm font-medium text-gray-700 dark:text-gray-300">
          {getStatusLabel()}
        </span>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 gap-3 mb-3">
        {/* Percent dry */}
        <div className="flex flex-col">
          <span className="text-2xl font-bold text-gray-900 dark:text-white">
            {Math.round(stats.percent_dry)}%
          </span>
          <span className="text-xs text-gray-600 dark:text-gray-400">Dry</span>
        </div>

        {/* Route counts */}
        <div className="flex flex-col">
          <span className="text-2xl font-bold text-gray-900 dark:text-white">
            {stats.total_routes}
          </span>
          <span className="text-xs text-gray-600 dark:text-gray-400">Routes</span>
        </div>
      </div>

      {/* Detailed stats */}
      <div className="grid grid-cols-3 gap-2 pt-3 border-t border-gray-300 dark:border-gray-600">
        <div className="flex flex-col items-center text-center">
          <Sun className="w-4 h-4 text-green-600 dark:text-green-400 mb-1" />
          <span className="text-sm font-semibold text-gray-900 dark:text-white">
            {stats.dry_count}
          </span>
          <span className="text-xs text-gray-600 dark:text-gray-400">Dry</span>
        </div>

        <div className="flex flex-col items-center text-center">
          <Clock className="w-4 h-4 text-yellow-600 dark:text-yellow-400 mb-1" />
          <span className="text-sm font-semibold text-gray-900 dark:text-white">
            {stats.drying_count}
          </span>
          <span className="text-xs text-gray-600 dark:text-gray-400">Drying</span>
        </div>

        <div className="flex flex-col items-center text-center">
          <Droplets className="w-4 h-4 text-red-600 dark:text-red-400 mb-1" />
          <span className="text-sm font-semibold text-gray-900 dark:text-white">
            {stats.wet_count}
          </span>
          <span className="text-xs text-gray-600 dark:text-gray-400">Wet</span>
        </div>
      </div>

      {/* Additional info */}
      {stats.avg_hours_until_dry > 0 && (
        <div className="mt-3 pt-3 border-t border-gray-300 dark:border-gray-600">
          <div className="flex items-center justify-between text-xs">
            <div className="flex items-center gap-1 text-gray-600 dark:text-gray-400">
              <Clock className="w-3 h-3" />
              <span>Avg. dry time</span>
            </div>
            <span className="font-medium text-gray-900 dark:text-white">
              {formatHours(stats.avg_hours_until_dry)}
            </span>
          </div>
        </div>
      )}

      {stats.avg_tree_coverage > 0 && (
        <div className="mt-2">
          <div className="flex items-center justify-between text-xs">
            <div className="flex items-center gap-1 text-gray-600 dark:text-gray-400">
              <TreeDeciduous className="w-3 h-3" />
              <span>Avg. tree coverage</span>
            </div>
            <span className="font-medium text-gray-900 dark:text-white">
              {Math.round(stats.avg_tree_coverage)}%
            </span>
          </div>
        </div>
      )}
    </div>
  );
};
