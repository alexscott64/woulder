import React from 'react';
import { BoulderDryingStatus } from '../types/weather';
import { Droplet, Sun, TreePine, MapPin } from 'lucide-react';

interface BoulderDryingBadgeProps {
  status: BoulderDryingStatus;
  compact?: boolean;
}

export const BoulderDryingBadge: React.FC<BoulderDryingBadgeProps> = ({ status, compact = false }) => {
  // Determine badge color based on status
  const getStatusColor = () => {
    if (!status.is_wet) return 'bg-green-500';

    switch (status.status) {
      case 'critical':
        return 'bg-red-600';
      case 'poor':
        return 'bg-orange-500';
      case 'fair':
        return 'bg-yellow-500';
      case 'good':
        return 'bg-green-500';
      default:
        return 'bg-gray-500';
    }
  };

  // Format hours until dry
  const formatHoursUntilDry = (hours: number): string => {
    if (hours <= 0) return 'Dry';
    if (hours < 1) return '<1h';
    if (hours < 24) return `${Math.round(hours)}h`;
    const days = Math.floor(hours / 24);
    return `${days}d`;
  };

  // Get confidence indicator
  const getConfidenceIndicator = () => {
    if (status.confidence_score >= 90) return '●';
    if (status.confidence_score >= 70) return '◐';
    return '○';
  };

  if (compact) {
    // Compact badge for route lists
    return (
      <div className="flex items-center gap-1 text-xs">
        <div className={`${getStatusColor()} text-white px-2 py-0.5 rounded flex items-center gap-1`}>
          <Droplet className="w-3 h-3" />
          <span className="font-medium">{formatHoursUntilDry(status.hours_until_dry)}</span>
        </div>
        <span className="text-gray-400" title={`Confidence: ${status.confidence_score}%`}>
          {getConfidenceIndicator()}
        </span>
      </div>
    );
  }

  // Detailed badge for expanded views
  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
      <div className="flex items-start justify-between mb-3">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <div className={`${getStatusColor()} text-white px-3 py-1 rounded-md font-medium flex items-center gap-2`}>
              <Droplet className="w-4 h-4" />
              {status.is_wet ? `${formatHoursUntilDry(status.hours_until_dry)} until dry` : 'Dry'}
            </div>
            <span className="text-gray-400 text-sm">
              {status.confidence_score}% confidence
            </span>
          </div>
          <p className="text-sm text-gray-300">{status.message}</p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 text-sm">
        <div className="flex items-center gap-2">
          <Sun className="w-4 h-4 text-yellow-400" />
          <div>
            <div className="text-gray-400">Sun Exposure</div>
            <div className="text-white font-medium">{Math.round(status.sun_exposure_hours)}h over 6 days</div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <TreePine className="w-4 h-4 text-green-400" />
          <div>
            <div className="text-gray-400">Tree Coverage</div>
            <div className="text-white font-medium">{Math.round(status.tree_coverage_percent)}%</div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <MapPin className="w-4 h-4 text-blue-400" />
          <div>
            <div className="text-gray-400">Aspect</div>
            <div className="text-white font-medium">{status.aspect}</div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Droplet className="w-4 h-4 text-gray-400" />
          <div>
            <div className="text-gray-400">Rock Type</div>
            <div className="text-white font-medium">{status.rock_type}</div>
          </div>
        </div>
      </div>

      <div className="mt-3 pt-3 border-t border-gray-700 text-xs text-gray-400">
        <div className="flex items-center justify-between">
          <span>GPS: {status.latitude.toFixed(5)}, {status.longitude.toFixed(5)}</span>
          <span>Last rain: {new Date(status.last_rain_timestamp).toLocaleDateString()}</span>
        </div>
      </div>
    </div>
  );
};

export default BoulderDryingBadge;
