import { X, Info, Stone, Waves, Bug, Droplet, AlertTriangle, AlertCircle, TrendingUp, Clock } from 'lucide-react';
import { format } from 'date-fns';
import { RockDryingStatus } from '../types/weather';
import type { PestConditions } from '../utils/pests/analyzers/PestAnalyzer';
import { RiverData } from '../types/river';
import { useState } from 'react';
import { PestLevel } from '../utils/pests/calculations/pests';

interface ConditionsModalProps {
  locationName: string;
  rockStatus?: RockDryingStatus;
  pestConditions?: PestConditions;
  riverData?: RiverData[];
  onClose: () => void;
}

type TabType = 'rock' | 'rivers' | 'pests';

export function ConditionsModal({
  locationName,
  rockStatus,
  pestConditions,
  riverData,
  onClose
}: ConditionsModalProps) {
  // Determine which tabs are available and set initial tab
  const availableTabs: TabType[] = [];
  if (rockStatus) availableTabs.push('rock');
  if (riverData && riverData.length > 0) availableTabs.push('rivers');
  if (pestConditions) availableTabs.push('pests');

  const [activeTab, setActiveTab] = useState<TabType>(availableTabs[0] || 'rock');

  const getTabIcon = (tab: TabType) => {
    switch (tab) {
      case 'rock':
        return <Stone className="w-4 h-4" />;
      case 'rivers':
        return <Waves className="w-4 h-4" />;
      case 'pests':
        return <Bug className="w-4 h-4" />;
    }
  };

  const getTabLabel = (tab: TabType) => {
    switch (tab) {
      case 'rock':
        return 'Rock Conditions';
      case 'rivers':
        return 'River Crossings';
      case 'pests':
        return 'Pest Activity';
    }
  };

  const getRockStatusColor = (status: string) => {
    switch (status) {
      case 'critical':
        return 'text-red-600 dark:text-red-400';
      case 'poor':
        return 'text-orange-600 dark:text-orange-400';
      case 'fair':
        return 'text-yellow-600 dark:text-yellow-400';
      case 'good':
        return 'text-green-600 dark:text-green-400';
      default:
        return 'text-gray-600 dark:text-gray-400';
    }
  };

  const getRockStatusBgColor = (status: string) => {
    switch (status) {
      case 'critical':
        return 'bg-red-100 dark:bg-red-900/30';
      case 'poor':
        return 'bg-orange-100 dark:bg-orange-900/30';
      case 'fair':
        return 'bg-yellow-100 dark:bg-yellow-900/30';
      case 'good':
        return 'bg-green-100 dark:bg-green-900/30';
      default:
        return 'bg-gray-100 dark:bg-gray-900/30';
    }
  };

  const getRockStatusText = (status: string) => {
    switch (status) {
      case 'critical':
        return 'CRITICAL - DO NOT CLIMB';
      case 'poor':
        return 'POOR - NOT RECOMMENDED';
      case 'fair':
        return 'FAIR - DRYING';
      case 'good':
        return 'GOOD - DRY';
      default:
        return 'UNKNOWN';
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

  const getPestLevelBadgeStyles = (level: PestLevel) => {
    switch (level) {
      case 'extreme':
        return {
          bg: 'bg-red-100 dark:bg-red-900/30',
          text: 'text-red-700 dark:text-red-300',
          border: 'border-red-300 dark:border-red-700'
        };
      case 'very_high':
        return {
          bg: 'bg-orange-100 dark:bg-orange-900/30',
          text: 'text-orange-700 dark:text-orange-300',
          border: 'border-orange-300 dark:border-orange-700'
        };
      case 'high':
        return {
          bg: 'bg-yellow-100 dark:bg-yellow-900/30',
          text: 'text-yellow-700 dark:text-yellow-300',
          border: 'border-yellow-300 dark:border-yellow-700'
        };
      case 'moderate':
        return {
          bg: 'bg-yellow-50 dark:bg-yellow-900/20',
          text: 'text-yellow-600 dark:text-yellow-400',
          border: 'border-yellow-200 dark:border-yellow-600'
        };
      case 'low':
        return {
          bg: 'bg-green-100 dark:bg-green-900/30',
          text: 'text-green-700 dark:text-green-300',
          border: 'border-green-300 dark:border-green-700'
        };
    }
  };

  const getTabStatusDot = (tab: TabType) => {
    let dotColor = 'bg-gray-400';

    if (tab === 'rock' && rockStatus) {
      switch (rockStatus.status) {
        case 'critical':
          dotColor = 'bg-red-500';
          break;
        case 'poor':
          dotColor = 'bg-orange-500';
          break;
        case 'fair':
          dotColor = 'bg-yellow-500';
          break;
        case 'good':
          dotColor = 'bg-green-500';
          break;
      }
    } else if (tab === 'rivers' && riverData) {
      const allSafe = riverData.every(r => r.is_safe);
      const anySafe = riverData.some(r => r.is_safe);
      if (allSafe) {
        dotColor = 'bg-green-500';
      } else if (anySafe) {
        dotColor = 'bg-yellow-500';
      } else {
        dotColor = 'bg-red-500';
      }
    } else if (tab === 'pests' && pestConditions) {
      // Use the worse of mosquito or outdoor pest levels
      const worstLevel = pestConditions.mosquitoScore > pestConditions.outdoorPestScore
        ? pestConditions.mosquitoLevel
        : pestConditions.outdoorPestLevel;

      switch (worstLevel) {
        case 'extreme':
        case 'very_high':
          dotColor = 'bg-red-500';
          break;
        case 'high':
          dotColor = 'bg-orange-500';
          break;
        case 'moderate':
          dotColor = 'bg-yellow-500';
          break;
        case 'low':
          dotColor = 'bg-green-500';
          break;
      }
    }

    return dotColor;
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-2 sm:p-4">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl max-w-2xl w-full max-h-[95vh] sm:max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-3 sm:p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="min-w-0 flex-1 pr-2">
            <h2 className="text-base sm:text-lg font-bold text-gray-900 dark:text-white truncate">Conditions</h2>
            <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 truncate">{locationName}</p>
          </div>
          <button
            onClick={onClose}
            className="flex-shrink-0 p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full transition-colors"
          >
            <X className="w-5 h-5 text-gray-500 dark:text-gray-400" />
          </button>
        </div>

        {/* Mobile-Friendly Tab Navigation */}
        <div className="px-3 sm:px-4 pt-3 pb-2 bg-gray-50 dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
          <div className="flex gap-2 overflow-x-auto scrollbar-hide pb-1">
            {availableTabs.map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`flex items-center gap-2 px-3 sm:px-4 py-2 rounded-full font-medium text-xs sm:text-sm transition-all whitespace-nowrap flex-shrink-0 ${
                  activeTab === tab
                    ? 'bg-blue-500 text-white shadow-md scale-105'
                    : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 shadow-sm'
                }`}
              >
                {getTabIcon(tab)}
                <span className={`w-2 h-2 rounded-full flex-shrink-0 ${
                  activeTab === tab ? 'bg-white' : getTabStatusDot(tab)
                }`} />
                <span className="hidden sm:inline">{getTabLabel(tab)}</span>
                <span className="sm:hidden">
                  {tab === 'rock' ? 'Rock' : tab === 'rivers' ? 'Rivers' : 'Pests'}
                </span>
              </button>
            ))}
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-3 sm:p-4">
          {/* Rock Conditions Tab */}
          {activeTab === 'rock' && rockStatus && (
            <div className="space-y-4">
              {/* Status Card */}
              <div className={`rounded-lg p-4 ${getRockStatusBgColor(rockStatus.status)}`}>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-semibold text-gray-900 dark:text-white">Current Status</span>
                  <span className={`font-bold ${getRockStatusColor(rockStatus.status)}`}>
                    {getRockStatusText(rockStatus.status)}
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
                    <Info className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
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
          )}

          {/* River Crossings Tab */}
          {activeTab === 'rivers' && riverData && (
            <div className="space-y-4">
              {riverData.map((river, index) => {
                const getStatusColor = (status: string) => {
                  const baseStatus = status.replace('estimated ', '');
                  switch (baseStatus) {
                    case 'safe':
                      return 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-900 dark:text-green-100';
                    case 'caution':
                      return 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-900 dark:text-yellow-100';
                    case 'unsafe':
                      return 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-900 dark:text-red-100';
                    default:
                      return 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100';
                  }
                };

                const getStatusIcon = (status: string) => {
                  const baseStatus = status.replace('estimated ', '');
                  switch (baseStatus) {
                    case 'safe':
                      return <Droplet className="w-5 h-5 text-green-600 dark:text-green-400" />;
                    case 'caution':
                      return <AlertTriangle className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />;
                    case 'unsafe':
                      return <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400" />;
                    default:
                      return <Droplet className="w-5 h-5 text-gray-600 dark:text-gray-400" />;
                  }
                };

                return (
                  <div
                    key={index}
                    className={`border-2 rounded-lg p-4 ${getStatusColor(river.status)}`}
                  >
                    {/* River Name & Description */}
                    <div className="flex items-start gap-2 mb-3">
                      {getStatusIcon(river.status)}
                      <div className="flex-1">
                        <h3 className="font-bold text-lg">{river.river.river_name}</h3>
                        {river.river.description && (
                          <p className="text-sm opacity-75 mt-1">{river.river.description}</p>
                        )}
                      </div>
                    </div>

                    {/* Status Badge */}
                    <div className="mb-3">
                      <div className="inline-block px-3 py-1 bg-white dark:bg-gray-800 bg-opacity-70 rounded-lg font-bold text-xs uppercase">
                        {river.status}
                      </div>
                    </div>

                    {/* Status Message */}
                    <div className="mb-3 p-2.5 bg-white dark:bg-gray-800 bg-opacity-50 rounded-lg">
                      <p className="text-sm font-medium">{river.status_message}</p>
                    </div>

                    {/* Metrics Grid */}
                    <div className="grid grid-cols-2 gap-3 mb-3">
                      {/* Current Flow */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-2.5">
                        <div className="text-xs font-medium opacity-75 mb-1">Current Flow</div>
                        <div className="flex items-baseline gap-1">
                          <span className="text-xl font-bold">{Math.round(river.flow_cfs)}</span>
                          <span className="text-xs">CFS</span>
                        </div>
                      </div>

                      {/* Gauge Height */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-2.5">
                        <div className="text-xs font-medium opacity-75 mb-1">Gauge Height</div>
                        <div className="flex items-baseline gap-1">
                          <span className="text-xl font-bold">{river.gauge_height_ft.toFixed(2)}</span>
                          <span className="text-xs">ft</span>
                        </div>
                      </div>

                      {/* Safe Threshold */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-2.5">
                        <div className="text-xs font-medium opacity-75 mb-1">Safe Threshold</div>
                        <div className="flex items-baseline gap-1">
                          <span className="text-lg font-bold">{river.river.safe_crossing_cfs}</span>
                          <span className="text-xs">CFS</span>
                        </div>
                      </div>

                      {/* Percent of Safe */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-2.5">
                        <div className="text-xs font-medium opacity-75 mb-1 flex items-center gap-1">
                          <TrendingUp className="w-3 h-3" />
                          <span>% of Safe</span>
                        </div>
                        <div className="flex items-baseline gap-1">
                          <span className="text-xl font-bold">{Math.round(river.percent_of_safe)}%</span>
                        </div>
                      </div>
                    </div>

                    {/* Timestamp */}
                    <div className="pt-2 border-t border-current border-opacity-20">
                      <div className="flex items-center gap-1 text-xs opacity-75 mb-1.5">
                        <Clock className="w-3 h-3" />
                        <span>Last updated: {format(new Date(river.timestamp), 'MMM d, h:mm a')}</span>
                      </div>
                      {/* USGS Link */}
                      <a
                        href={`https://waterdata.usgs.gov/monitoring-location/${river.river.gauge_id}/`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-xs underline opacity-75 hover:opacity-100 transition-opacity"
                      >
                        View on USGS Water Data →
                      </a>
                    </div>
                  </div>
                );
              })}

              {/* Safety Warning */}
              <div className="mt-4 p-3 bg-gray-100 dark:bg-gray-900 rounded-lg">
                <p className="text-xs text-gray-600 dark:text-gray-400">
                  <strong>Safety Note:</strong> River conditions can change rapidly. Always assess conditions on-site before crossing. These thresholds are estimates and may not account for all factors affecting crossing safety.
                </p>
              </div>
            </div>
          )}

          {/* Pest Activity Tab */}
          {activeTab === 'pests' && pestConditions && (
            <div className="space-y-4">
              <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-3 mb-4">
                <div className="flex items-start gap-2">
                  <Info className="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
                  <p className="text-xs text-blue-800 dark:text-blue-200">
                    Pest activity is based on recent temperature patterns and seasonal timing. Higher activity means more bugs to deal with.
                  </p>
                </div>
              </div>

              {/* Mosquito Activity */}
              <div className={`rounded-lg p-4 ${getPestLevelBadgeStyles(pestConditions.mosquitoLevel).bg} border-2 ${getPestLevelBadgeStyles(pestConditions.mosquitoLevel).border}`}>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-semibold text-gray-900 dark:text-white">Mosquito Activity</span>
                  <span className={`font-bold ${getPestLevelBadgeStyles(pestConditions.mosquitoLevel).text}`}>
                    {pestConditions.mosquitoLevel.toUpperCase()}
                  </span>
                </div>
                <p className="text-sm text-gray-700 dark:text-gray-300 mb-2">
                  Score: {pestConditions.mosquitoScore}/100
                </p>
              </div>

              {/* Outdoor Pest Activity */}
              <div className={`rounded-lg p-4 ${getPestLevelBadgeStyles(pestConditions.outdoorPestLevel).bg} border-2 ${getPestLevelBadgeStyles(pestConditions.outdoorPestLevel).border}`}>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-semibold text-gray-900 dark:text-white">General Outdoor Pests</span>
                  <span className={`font-bold ${getPestLevelBadgeStyles(pestConditions.outdoorPestLevel).text}`}>
                    {pestConditions.outdoorPestLevel.toUpperCase()}
                  </span>
                </div>
                <p className="text-sm text-gray-700 dark:text-gray-300 mb-2">
                  Score: {pestConditions.outdoorPestScore}/100
                </p>
              </div>

              {/* Contributing Factors */}
              {pestConditions.factors.length > 0 && (
                <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-4">
                  <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-2">Contributing Factors</h4>
                  <ul className="space-y-1">
                    {pestConditions.factors.map((factor, index) => (
                      <li key={index} className="flex items-start gap-2 text-xs text-gray-700 dark:text-gray-300">
                        <span className="text-gray-400 dark:text-gray-600 mt-0.5">•</span>
                        <span>{factor}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-3 sm:p-4 border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={onClose}
            className="w-full py-2.5 sm:py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 font-medium rounded-lg transition-colors text-sm sm:text-base"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
