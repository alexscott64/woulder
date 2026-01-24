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

  // Consolidate adjacent periods with same status (backend creates too many tiny periods)
  const consolidatedForecast: DryingForecastPeriod[] = [];
  for (const period of forecast) {
    const last = consolidatedForecast[consolidatedForecast.length - 1];
    if (last && last.status === period.status) {
      // Same status - extend the last period
      last.end_time = period.end_time;
      if (period.rain_amount) {
        last.rain_amount = (last.rain_amount || 0) + period.rain_amount;
      }
    } else {
      // Different status - add new period
      consolidatedForecast.push({ ...period });
    }
  }

  // Calculate total time span - ALWAYS use now + 6 days
  const now = new Date();
  const startTime = now;
  const endTime = new Date(now.getTime() + 6 * 24 * 60 * 60 * 1000); // Always 6 days from now
  const totalDuration = endTime.getTime() - startTime.getTime();

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'dry':
        return 'bg-gradient-to-br from-emerald-400 to-green-500';
      case 'drying':
        return 'bg-gradient-to-br from-amber-400 to-yellow-500';
      case 'wet':
        return 'bg-gradient-to-br from-rose-400 to-red-500';
      default:
        return 'bg-gradient-to-br from-gray-300 to-gray-400';
    }
  };

  const getStatusDotColor = (status: string): string => {
    switch (status) {
      case 'dry':
        return 'bg-emerald-500';
      case 'drying':
        return 'bg-amber-500';
      case 'wet':
        return 'bg-rose-500';
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

  // Find wet/drying periods for markers
  const wetPeriods = consolidatedForecast.filter(p => p.status === 'wet' || p.status === 'drying');

  return (
    <div className={`space-y-2 ${className}`}>
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-semibold text-gray-700 dark:text-gray-300">6-Day Forecast</span>
        <div className="flex gap-2 text-[10px]">
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-gradient-to-br from-emerald-400 to-green-500" />
            <span className="text-gray-600 dark:text-gray-400">Dry</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-gradient-to-br from-amber-400 to-yellow-500" />
            <span className="text-gray-600 dark:text-gray-400">Drying</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-2 h-2 rounded-full bg-gradient-to-br from-rose-400 to-red-500" />
            <span className="text-gray-600 dark:text-gray-400">Wet</span>
          </div>
        </div>
      </div>

      {/* Timeline bar */}
      <div className="relative">
        <div className="relative h-8 bg-gradient-to-b from-gray-100 to-gray-200 dark:from-gray-700 dark:to-gray-800 rounded-lg overflow-hidden shadow-inner border border-gray-300 dark:border-gray-600">
          {consolidatedForecast.map((period, index) => {
            const periodStart = new Date(period.start_time);
            const periodEnd = period.end_time
              ? new Date(period.end_time)
              : endTime;

            const startPercent = Math.max(0, ((periodStart.getTime() - startTime.getTime()) / totalDuration) * 100);
            let endPercent = Math.min(100, ((periodEnd.getTime() - startTime.getTime()) / totalDuration) * 100);

            // If this is the last period, extend it to 100% to fill the bar
            const isLastPeriod = index === consolidatedForecast.length - 1;
            if (isLastPeriod && endPercent < 100) {
              endPercent = 100;
            }

            let width = Math.max(0, endPercent - startPercent);

            const colorClass = getStatusColor(period.status);

            // Give ALL periods minimum width for visibility (prevents gray gaps)
            if (width < 1) {
              width = 1; // Minimum 1% width
            }

            // Give wet/drying periods extra visibility
            const isWetOrDrying = period.status === 'wet' || period.status === 'drying';
            if (isWetOrDrying && width < 2) {
              width = 2; // Minimum 2% width for wet/drying
            }

            return (
              <div
                key={index}
                className={`absolute h-full ${colorClass} transition-all duration-200 group cursor-pointer hover:brightness-110`}
                style={{
                  left: `${startPercent}%`,
                  width: `${width}%`,
                }}
                title={`${period.status.charAt(0).toUpperCase() + period.status.slice(1)}: ${formatDate(period.start_time)}${period.end_time ? ` - ${formatDate(period.end_time)}` : ''}`}
              >
                <div className="absolute inset-0 bg-white/10 opacity-0 group-hover:opacity-100 transition-opacity" />
                <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-all duration-150">
                  <div className="bg-white/30 backdrop-blur-sm rounded-full p-1">
                    {getStatusIcon(period.status)}
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Markers only for wet/drying periods - limit to avoid overlap */}
        {wetPeriods.length > 0 && wetPeriods.length <= 5 && (
          <div className="relative mt-1.5 h-5">
            {wetPeriods.slice(0, 3).map((period, index) => {
              const periodStart = new Date(period.start_time);
              const startPercent = ((periodStart.getTime() - startTime.getTime()) / totalDuration) * 100;

              // Skip if too close to start (would overlap with legend)
              if (startPercent < 5) return null;

              return (
                <div
                  key={index}
                  className="absolute flex flex-col items-center -translate-x-1/2"
                  style={{ left: `${Math.max(5, Math.min(95, startPercent))}%` }}
                >
                  <div className="w-px h-2 bg-gray-400 dark:bg-gray-500" />
                  <span className="text-[9px] text-gray-600 dark:text-gray-400 font-medium whitespace-nowrap bg-white dark:bg-gray-800 px-1 rounded">
                    {formatDate(period.start_time)}
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};
