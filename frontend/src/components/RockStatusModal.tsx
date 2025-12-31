import { X, AlertTriangle, Info, Stone } from 'lucide-react';
import { RockDryingStatus } from '../types/weather';

interface RockStatusModalProps {
  rockStatus: RockDryingStatus;
  locationName: string;
  onClose: () => void;
}

export function RockStatusModal({ rockStatus, locationName, onClose }: RockStatusModalProps) {
  const getStatusIcon = () => {
    switch (rockStatus.status) {
      case 'critical':
        return <Stone className="w-5 h-5 text-red-600 dark:text-red-400" strokeWidth={2.5} />;
      case 'poor':
        return <Stone className="w-5 h-5 text-orange-600 dark:text-orange-400" strokeWidth={2.5} />;
      case 'fair':
        return <Stone className="w-5 h-5 text-yellow-600 dark:text-yellow-400" strokeWidth={2.5} />;
      case 'good':
        return <Stone className="w-5 h-5 text-green-600 dark:text-green-400" strokeWidth={2.5} />;
    }
  };

  const getStatusColor = () => {
    switch (rockStatus.status) {
      case 'critical':
        return 'text-red-600 dark:text-red-400';
      case 'poor':
        return 'text-red-600 dark:text-red-400';
      case 'fair':
        return 'text-yellow-600 dark:text-yellow-400';
      case 'good':
        return 'text-green-600 dark:text-green-400';
    }
  };

  const getStatusBgColor = () => {
    switch (rockStatus.status) {
      case 'critical':
        return 'bg-red-100 dark:bg-red-900/30';
      case 'poor':
        return 'bg-red-100 dark:bg-red-900/30';
      case 'fair':
        return 'bg-yellow-100 dark:bg-yellow-900/30';
      case 'good':
        return 'bg-green-100 dark:bg-green-900/30';
    }
  };

  const getStatusText = () => {
    switch (rockStatus.status) {
      case 'critical':
        return 'CRITICAL - DO NOT CLIMB';
      case 'poor':
        return 'POOR - NOT RECOMMENDED';
      case 'fair':
        return 'FAIR - DRYING';
      case 'good':
        return 'GOOD - DRY';
    }
  };

  const formatLastRain = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffHours = Math.floor(diffMs / (1000 * 60 * 60));

      if (diffHours < 1) return 'Less than 1 hour ago';
      if (diffHours === 1) return '1 hour ago';
      if (diffHours < 24) return `${diffHours} hours ago`;

      const diffDays = Math.floor(diffHours / 24);
      if (diffDays === 1) return '1 day ago';
      return `${diffDays} days ago`;
    } catch {
      return 'Unknown';
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-2">
            {getStatusIcon()}
            <h2 className="text-lg font-bold text-gray-900 dark:text-white">Rock Conditions</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full transition-colors"
          >
            <X className="w-5 h-5 text-gray-500 dark:text-gray-400" />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Rock drying status for <span className="font-semibold">{locationName}</span> based on rock type and current weather conditions.
          </p>

          {/* Status Card */}
          <div className={`rounded-lg p-4 ${getStatusBgColor()}`}>
            <div className="flex items-center justify-between mb-2">
              <span className="font-semibold text-gray-900 dark:text-white">Current Status</span>
              <span className={`font-bold ${getStatusColor()}`}>
                {getStatusText()}
              </span>
            </div>
            <p className="text-sm text-gray-700 dark:text-gray-300 mt-2">
              {rockStatus.message}
            </p>
          </div>

          {/* Rock Type Category */}
          <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-4">
            <div className="mb-2">
              <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">Rock Type Category</span>
            </div>
            <p className="text-lg font-bold text-gray-900 dark:text-white">
              {rockStatus.primary_group_name}
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Specific types: {rockStatus.rock_types.join(', ')}
            </p>
          </div>

          {/* Drying Timeline */}
          {rockStatus.is_wet && rockStatus.hours_until_dry > 0 && (
            <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">Estimated Dry Time</span>
                <span className="text-lg font-bold text-gray-900 dark:text-white">
                  {Math.ceil(rockStatus.hours_until_dry)}h
                </span>
              </div>
              {rockStatus.last_rain_timestamp && (
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  Last rain: {formatLastRain(rockStatus.last_rain_timestamp)}
                </p>
              )}
            </div>
          )}

          {/* Wet-Sensitive Warning */}
          {rockStatus.is_wet_sensitive && rockStatus.is_wet && (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3">
              <div className="flex items-start gap-2">
                <AlertTriangle className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-semibold text-red-900 dark:text-red-200 mb-1">
                    Wet-Sensitive Rock Warning
                  </p>
                  <p className="text-xs text-red-800 dark:text-red-300">
                    {rockStatus.primary_group_name} is permanently damaged when climbed wet. Climbing on wet rock can break holds and ruin routes for future climbers. Please wait until completely dry.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Info Box */}
          <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-3">
            <div className="flex items-start gap-2">
              <Info className="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-xs text-blue-800 dark:text-blue-200">
                  <strong>How it's calculated:</strong> Drying time is estimated based on rock type porosity, recent precipitation amount, current temperature, humidity, wind speed, cloud cover, and sun exposure. Wet-sensitive rocks (sandstone, arkose, graywacke) require extra drying time for safety.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onClose}
            className="w-full py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 font-medium rounded-lg transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
