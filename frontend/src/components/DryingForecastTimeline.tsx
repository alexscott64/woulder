import React from 'react';
import { DryingForecastPeriod } from '../types/weather';
import { Droplets, Sun, Clock } from 'lucide-react';

interface DryingForecastTimelineProps {
  forecast: DryingForecastPeriod[];
  className?: string;
}

export const DryingForecastTimeline: React.FC<DryingForecastTimelineProps> = ({
  forecast,
  className = '',
}) => {
  if (!forecast || forecast.length === 0) {
    return null;
  }

  // Calculate total time span for percentage calculations
  const startTime = new Date(forecast[0].start_time);
  const lastPeriod = forecast[forecast.length - 1];
  const endTime = lastPeriod.end_time
    ? new Date(lastPeriod.end_time)
    : new Date(startTime.getTime() + 6 * 24 * 60 * 60 * 1000); // 6 days default

  const totalDuration = endTime.getTime() - startTime.getTime();

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'dry':
        return 'bg-green-500';
      case 'drying':
        return 'bg-yellow-500';
      case 'wet':
        return 'bg-red-500';
      default:
        return 'bg-gray-400';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'dry':
        return <Sun className="w-3 h-3" />;
      case 'drying':
        return <Clock className="w-3 h-3" />;
      case 'wet':
        return <Droplets className="w-3 h-3" />;
      default:
        return null;
    }
  };

  const formatDate = (dateStr: string): string => {
    const date = new Date(dateStr);
    const today = new Date();
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);

    if (date.toDateString() === today.toDateString()) {
      return 'Now';
    } else if (date.toDateString() === tomorrow.toDateString()) {
      return 'Tomorrow';
    } else {
      return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
    }
  };

  const formatDuration = (hours: number): string => {
    if (hours < 24) {
      return `${Math.round(hours)}h`;
    }
    const days = Math.floor(hours / 24);
    const remainingHours = Math.round(hours % 24);
    return remainingHours > 0 ? `${days}d ${remainingHours}h` : `${days}d`;
  };

  return (
    <div className={`space-y-3 ${className}`}>
      <div className="flex items-center justify-between text-xs text-gray-600 dark:text-gray-400">
        <span className="font-medium">6-Day Forecast</span>
        <div className="flex gap-3">
          <div className="flex items-center gap-1">
            <Sun className="w-3 h-3 text-green-500" />
            <span>Dry</span>
          </div>
          <div className="flex items-center gap-1">
            <Clock className="w-3 h-3 text-yellow-500" />
            <span>Drying</span>
          </div>
          <div className="flex items-center gap-1">
            <Droplets className="w-3 h-3 text-red-500" />
            <span>Wet</span>
          </div>
        </div>
      </div>

      {/* Timeline bar */}
      <div className="relative h-8 bg-gray-200 dark:bg-gray-700 rounded-lg overflow-hidden">
        {forecast.map((period, index) => {
          const periodStart = new Date(period.start_time);
          const periodEnd = period.end_time
            ? new Date(period.end_time)
            : endTime;

          const startPercent = ((periodStart.getTime() - startTime.getTime()) / totalDuration) * 100;
          const width = ((periodEnd.getTime() - periodStart.getTime()) / totalDuration) * 100;

          return (
            <div
              key={index}
              className={`absolute h-full ${getStatusColor(period.status)} transition-all duration-300 group cursor-pointer`}
              style={{
                left: `${startPercent}%`,
                width: `${width}%`,
              }}
              title={`${period.status.charAt(0).toUpperCase() + period.status.slice(1)}: ${formatDate(period.start_time)}${period.end_time ? ` - ${formatDate(period.end_time)}` : ''}`}
            >
              <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                {getStatusIcon(period.status)}
              </div>
            </div>
          );
        })}
      </div>

      {/* Period details */}
      <div className="space-y-2">
        {forecast.slice(0, 3).map((period, index) => (
          <div
            key={index}
            className="flex items-center justify-between text-xs p-2 bg-gray-50 dark:bg-gray-800 rounded"
          >
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${getStatusColor(period.status)}`} />
              <span className="font-medium">{formatDate(period.start_time)}</span>
              <span className="text-gray-500 dark:text-gray-400">
                {period.status.charAt(0).toUpperCase() + period.status.slice(1)}
              </span>
            </div>
            <div className="flex items-center gap-3 text-gray-600 dark:text-gray-400">
              {period.hours_until_dry !== undefined && period.hours_until_dry > 0 && (
                <span className="flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  {formatDuration(period.hours_until_dry)}
                </span>
              )}
              {period.rain_amount !== undefined && period.rain_amount > 0 && (
                <span className="flex items-center gap-1">
                  <Droplets className="w-3 h-3" />
                  {period.rain_amount.toFixed(2)}"
                </span>
              )}
            </div>
          </div>
        ))}
        {forecast.length > 3 && (
          <div className="text-xs text-center text-gray-500 dark:text-gray-400">
            + {forecast.length - 3} more periods
          </div>
        )}
      </div>
    </div>
  );
};
