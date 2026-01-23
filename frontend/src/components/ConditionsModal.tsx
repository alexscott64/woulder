import { X, Info, Stone, Waves, Bug, Droplet, AlertTriangle, AlertCircle, TrendingUp, Clock, CalendarCheck } from 'lucide-react';
import { format } from 'date-fns';
import { RockDryingStatus, WeatherCondition, PestConditions, PestLevel } from '../types/weather';
import { RiverData } from '../types/river';
import { useState } from 'react';
import { formatDryTime } from '../utils/weather/formatters';
import { getConditionBadgeStyles, getConditionColor } from './weather/weatherDisplay';

interface ConditionsModalProps {
  locationName: string;
  rockStatus?: RockDryingStatus;
  pestConditions?: PestConditions;
  riverData?: RiverData[];
  todayCondition: WeatherCondition;
  onClose: () => void;
}

type TabType = 'today' | 'rock' | 'rivers' | 'pests';

export function ConditionsModal({
  locationName,
  rockStatus,
  pestConditions,
  riverData,
  todayCondition,
  onClose
}: ConditionsModalProps) {
  // Determine which tabs are available and set initial tab
  const availableTabs: TabType[] = ['today']; // Always show today
  if (rockStatus) availableTabs.push('rock');
  if (riverData && riverData.length > 0) availableTabs.push('rivers');
  if (pestConditions) availableTabs.push('pests');

  const [activeTab, setActiveTab] = useState<TabType>('today');

  const getTabIcon = (tab: TabType) => {
    switch (tab) {
      case 'today':
        return <CalendarCheck className="w-3.5 h-3.5" />;
      case 'rock':
        return <Stone className="w-3.5 h-3.5" />;
      case 'rivers':
        return <Waves className="w-3.5 h-3.5" />;
      case 'pests':
        return <Bug className="w-3.5 h-3.5" />;
    }
  };

  const getRockStatusColor = (status: string) => {
    switch (status) {
      case 'critical':
        return 'text-red-600 dark:text-red-400';
      case 'poor':
        return 'text-red-500 dark:text-red-300'; // Changed from orange to red
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
        return 'bg-red-50 dark:bg-red-900/20'; // Changed from orange to red (lighter shade)
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
      // Handle empty or zero time values
      if (!timestamp || timestamp === '0001-01-01T00:00:00Z' || timestamp.startsWith('0001-')) {
        return 'No recent rain';
      }

      const date = new Date(timestamp);

      // Check if date is invalid or before year 2000 (likely a zero value)
      if (isNaN(date.getTime()) || date.getFullYear() < 2000) {
        return 'No recent rain';
      }

      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffHours = Math.floor(diffMs / (1000 * 60 * 60));

      if (diffHours < 1) return 'Less than 1 hour ago';
      if (diffHours === 1) return '1 hour ago';
      if (diffHours < 24) return `${diffHours} hours ago`;

      const diffDays = Math.floor(diffHours / 24);
      if (diffDays === 1) return '1 day ago';
      if (diffDays > 365) return 'No recent rain';
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

    if (tab === 'today') {
      // Use condition level
      switch (todayCondition.level) {
        case 'bad':
          dotColor = 'bg-red-500';
          break;
        case 'marginal':
          dotColor = 'bg-yellow-500';
          break;
        case 'good':
          dotColor = 'bg-green-500';
          break;
      }
    } else if (tab === 'rock' && rockStatus) {
      switch (rockStatus.status) {
        case 'critical':
          dotColor = 'bg-red-500';
          break;
        case 'poor':
          dotColor = 'bg-red-500';
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
      const worstLevel = pestConditions.mosquito_score > pestConditions.outdoor_pest_score
        ? pestConditions.mosquito_level
        : pestConditions.outdoor_pest_level;

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

  const getTabLabel = (tab: TabType) => {
    switch (tab) {
      case 'today': return 'Today';
      case 'rock': return 'Rock';
      case 'rivers': return 'Rivers';
      case 'pests': return 'Pests';
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-2 sm:p-4">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-xl max-w-2xl w-full max-h-[96vh] overflow-hidden flex flex-col">
        {/* Compact Header */}
        <div className="flex items-center justify-between px-3 sm:px-4 py-2 sm:py-3 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-white dark:from-gray-900 dark:to-gray-800">
          <div className="min-w-0 flex-1 pr-2">
            <h2 className="text-sm sm:text-base font-bold text-gray-900 dark:text-white truncate">Conditions</h2>
            <p className="text-[10px] sm:text-xs text-gray-600 dark:text-gray-400 truncate">{locationName}</p>
          </div>
          <button
            onClick={onClose}
            className="flex-shrink-0 p-1 sm:p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full transition-colors"
          >
            <X className="w-4 h-4 sm:w-5 sm:h-5 text-gray-500 dark:text-gray-400" />
          </button>
        </div>

        {/* Compact Tab Navigation - Fixed Grid Layout */}
        <div className="border-b border-gray-200 dark:border-gray-700 bg-gradient-to-b from-gray-50 to-white dark:from-gray-800 dark:to-gray-900 px-2 sm:px-3 py-1.5">
          <div className={`grid gap-1 ${availableTabs.length === 4 ? 'grid-cols-4' : availableTabs.length === 3 ? 'grid-cols-3' : availableTabs.length === 2 ? 'grid-cols-2' : 'grid-cols-1'}`}>
            {availableTabs.map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                title={getTabLabel(tab)}
                className={`relative flex flex-col sm:flex-row items-center justify-center gap-0.5 sm:gap-1.5 px-1.5 sm:px-2 py-1.5 rounded-lg text-[10px] sm:text-xs font-medium transition-all ${
                  activeTab === tab
                    ? 'bg-blue-500 text-white shadow-md scale-105'
                    : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-600'
                }`}
              >
                <div className="flex items-center justify-center">
                  {getTabIcon(tab)}
                </div>
                <span className="leading-tight">{getTabLabel(tab)}</span>
                {/* Status Dot - positioned absolutely on mobile, normally on desktop */}
                <div className={`absolute -top-0.5 -right-0.5 sm:static sm:ml-auto w-2 h-2 sm:w-1.5 sm:h-1.5 rounded-full ${getTabStatusDot(tab)} ${activeTab === tab ? 'ring-2 ring-white' : ''}`} />
              </button>
            ))}
          </div>
        </div>

        {/* Content - Compact Padding */}
        <div className="flex-1 overflow-y-auto p-2 sm:p-4">
          {/* Today's Conditions Tab */}
          {activeTab === 'today' && (
            <div className="space-y-2 sm:space-y-3">
              {/* Condition Summary Card */}
              <div className={`rounded-lg sm:rounded-xl p-3 sm:p-4 border-2 ${getConditionBadgeStyles(todayCondition.level).border} ${getConditionBadgeStyles(todayCondition.level).bg}`}>
                <div className="flex items-center gap-2 sm:gap-3 mb-2 sm:mb-3">
                  <div className={`w-3 h-3 sm:w-4 sm:h-4 rounded-full ${getConditionColor(todayCondition.level)}`} />
                  <h3 className={`text-sm sm:text-lg font-bold ${getConditionBadgeStyles(todayCondition.level).text}`}>
                    {todayCondition.level === 'good' ? 'Good Conditions' : todayCondition.level === 'marginal' ? 'Fair Conditions' : 'Poor Conditions'}
                  </h3>
                </div>

                {/* Contributing Factors */}
                <div className="space-y-1 sm:space-y-2">
                  <h4 className="text-xs sm:text-sm font-semibold text-gray-700 dark:text-gray-300">Contributing Factors:</h4>
                  <ul className="space-y-1">
                    {todayCondition.reasons.map((reason, index) => (
                      <li key={index} className="flex items-start gap-1.5 sm:gap-2 text-xs sm:text-sm text-gray-700 dark:text-gray-300">
                        <span className="text-gray-400 mt-0.5">•</span>
                        <span>{reason}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>

              {/* Info Note */}
              <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-2 sm:p-3">
                <div className="flex items-start gap-1.5 sm:gap-2">
                  <Info className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
                  <div className="text-[11px] sm:text-sm text-blue-900 dark:text-blue-200">
                    <p className="font-medium mb-0.5 sm:mb-1">About Today's Conditions</p>
                    <p className="text-blue-800 dark:text-blue-300">
                      This rating considers all forecasted conditions throughout the climbing day (9am-8pm Pacific),
                      weighted by the worst conditions you'll encounter. Check the other tabs for specific details.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Rock Conditions Tab */}
          {activeTab === 'rock' && rockStatus && (
            <div className="space-y-2 sm:space-y-3">
              {/* Status Card */}
              <div className={`rounded-lg p-3 sm:p-4 ${getRockStatusBgColor(rockStatus.status)}`}>
                <div className="flex items-center justify-between mb-1.5 sm:mb-2">
                  <span className="text-xs sm:text-sm font-semibold text-gray-900 dark:text-white">Current Status</span>
                  <span className={`text-xs sm:text-sm font-bold ${getRockStatusColor(rockStatus.status)}`}>
                    {getRockStatusText(rockStatus.status)}
                  </span>
                </div>
                <p className="text-xs sm:text-sm text-gray-700 dark:text-gray-300">
                  {rockStatus.message}
                </p>
              </div>

              {/* Rock Type Category */}
              <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3 sm:p-4">
                <div className="mb-1.5">
                  <span className="text-xs sm:text-sm font-semibold text-gray-700 dark:text-gray-300">Rock Type Category</span>
                </div>
                <p className="text-sm sm:text-lg font-bold text-gray-900 dark:text-white">
                  {rockStatus.primary_group_name}
                </p>
                <p className="text-[10px] sm:text-xs text-gray-500 dark:text-gray-400 mt-1">
                  Specific types: {rockStatus.rock_types.join(', ')}
                </p>
              </div>

              {/* Drying Timeline */}
              {rockStatus.is_wet && rockStatus.hours_until_dry > 0 && (
                <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3 sm:p-4">
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="text-xs sm:text-sm font-semibold text-gray-700 dark:text-gray-300">Estimated Dry Time</span>
                    <span className="text-sm sm:text-lg font-bold text-gray-900 dark:text-white">
                      {formatDryTime(rockStatus.hours_until_dry)}
                    </span>
                  </div>
                  {rockStatus.last_rain_timestamp && (
                    <p className="text-[10px] sm:text-xs text-gray-500 dark:text-gray-400">
                      Last rain: {formatLastRain(rockStatus.last_rain_timestamp)}
                    </p>
                  )}
                </div>
              )}

              {/* Wet-Sensitive Warning */}
              {rockStatus.is_wet_sensitive && rockStatus.is_wet && (
                <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-2 sm:p-3">
                  <div className="flex items-start gap-1.5 sm:gap-2">
                    <Info className="w-4 h-4 sm:w-5 sm:h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
                    <div>
                      <p className="text-xs sm:text-sm font-semibold text-red-900 dark:text-red-200 mb-0.5 sm:mb-1">
                        Wet-Sensitive Rock Warning
                      </p>
                      <p className="text-[10px] sm:text-xs text-red-800 dark:text-red-300">
                        {rockStatus.primary_group_name} is permanently damaged when climbed wet. Climbing on wet rock can break holds and ruin routes. Please wait until completely dry.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Info Box */}
              <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-2 sm:p-3">
                <div className="flex items-start gap-1.5 sm:gap-2">
                  <Info className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="text-[10px] sm:text-xs text-blue-800 dark:text-blue-200">
                      <strong>How it's calculated:</strong> Drying time is estimated based on rock type porosity, recent precipitation, temperature, humidity, wind, cloud cover, and sun exposure.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* River Crossings Tab */}
          {activeTab === 'rivers' && riverData && (
            <div className="space-y-2 sm:space-y-3">
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
                      return <Droplet className="w-4 h-4 sm:w-5 sm:h-5 text-green-600 dark:text-green-400" />;
                    case 'caution':
                      return <AlertTriangle className="w-4 h-4 sm:w-5 sm:h-5 text-yellow-600 dark:text-yellow-400" />;
                    case 'unsafe':
                      return <AlertCircle className="w-4 h-4 sm:w-5 sm:h-5 text-red-600 dark:text-red-400" />;
                    default:
                      return <Droplet className="w-4 h-4 sm:w-5 sm:h-5 text-gray-600 dark:text-gray-400" />;
                  }
                };

                return (
                  <div
                    key={index}
                    className={`border-2 rounded-lg p-2.5 sm:p-4 ${getStatusColor(river.status)}`}
                  >
                    {/* River Name & Description */}
                    <div className="flex items-start gap-1.5 sm:gap-2 mb-2 sm:mb-3">
                      {getStatusIcon(river.status)}
                      <div className="flex-1 min-w-0">
                        <h3 className="font-bold text-sm sm:text-base truncate">{river.river.river_name}</h3>
                        {river.river.description && (
                          <p className="text-xs sm:text-sm opacity-75 mt-0.5 line-clamp-2">{river.river.description}</p>
                        )}
                      </div>
                    </div>

                    {/* Status Badge */}
                    <div className="mb-2 sm:mb-3">
                      <div className="inline-block px-2 sm:px-3 py-0.5 sm:py-1 bg-white dark:bg-gray-800 bg-opacity-70 rounded-lg font-bold text-[10px] sm:text-xs uppercase">
                        {river.status}
                      </div>
                    </div>

                    {/* Status Message */}
                    <div className="mb-2 sm:mb-3 p-2 sm:p-2.5 bg-white dark:bg-gray-800 bg-opacity-50 rounded-lg">
                      <p className="text-xs sm:text-sm font-medium">{river.status_message}</p>
                    </div>

                    {/* Metrics Grid */}
                    <div className="grid grid-cols-2 gap-1.5 sm:gap-3 mb-2 sm:mb-3">
                      {/* Current Flow */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-1.5 sm:p-2.5">
                        <div className="text-[10px] sm:text-xs font-medium opacity-75 mb-0.5 sm:mb-1">Current Flow</div>
                        <div className="flex items-baseline gap-0.5 sm:gap-1">
                          <span className="text-sm sm:text-xl font-bold">{Math.round(river.flow_cfs)}</span>
                          <span className="text-[9px] sm:text-xs">CFS</span>
                        </div>
                      </div>

                      {/* Gauge Height */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-1.5 sm:p-2.5">
                        <div className="text-[10px] sm:text-xs font-medium opacity-75 mb-0.5 sm:mb-1">Gauge Height</div>
                        <div className="flex items-baseline gap-0.5 sm:gap-1">
                          <span className="text-sm sm:text-xl font-bold">{river.gauge_height_ft.toFixed(2)}</span>
                          <span className="text-[9px] sm:text-xs">ft</span>
                        </div>
                      </div>

                      {/* Safe Threshold */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-1.5 sm:p-2.5">
                        <div className="text-[10px] sm:text-xs font-medium opacity-75 mb-0.5 sm:mb-1">Safe Threshold</div>
                        <div className="flex items-baseline gap-0.5 sm:gap-1">
                          <span className="text-sm sm:text-lg font-bold">{river.river.safe_crossing_cfs}</span>
                          <span className="text-[9px] sm:text-xs">CFS</span>
                        </div>
                      </div>

                      {/* Percent of Safe */}
                      <div className="bg-white dark:bg-gray-800 bg-opacity-40 rounded-lg p-1.5 sm:p-2.5">
                        <div className="text-[10px] sm:text-xs font-medium opacity-75 mb-0.5 sm:mb-1 flex items-center gap-0.5 sm:gap-1">
                          <TrendingUp className="w-2.5 h-2.5 sm:w-3 sm:h-3" />
                          <span>% of Safe</span>
                        </div>
                        <div className="flex items-baseline gap-0.5 sm:gap-1">
                          <span className="text-sm sm:text-xl font-bold">{Math.round(river.percent_of_safe)}%</span>
                        </div>
                      </div>
                    </div>

                    {/* Timestamp */}
                    <div className="pt-1.5 sm:pt-2 border-t border-current border-opacity-20">
                      <div className="flex items-center gap-0.5 sm:gap-1 text-[10px] sm:text-xs opacity-75 mb-1">
                        <Clock className="w-2.5 h-2.5 sm:w-3 sm:h-3" />
                        <span>Last updated: {format(new Date(river.timestamp), 'MMM d, h:mm a')}</span>
                      </div>
                      {/* USGS Link */}
                      <a
                        href={`https://waterdata.usgs.gov/monitoring-location/${river.river.gauge_id}/`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-[10px] sm:text-xs underline opacity-75 hover:opacity-100 transition-opacity"
                      >
                        View on USGS Water Data →
                      </a>
                    </div>
                  </div>
                );
              })}

              {/* Safety Warning */}
              <div className="mt-2 sm:mt-3 p-2 sm:p-3 bg-gray-100 dark:bg-gray-900 rounded-lg">
                <p className="text-[10px] sm:text-xs text-gray-600 dark:text-gray-400">
                  <strong>Safety Note:</strong> River conditions can change rapidly. Always assess conditions on-site before crossing.
                </p>
              </div>
            </div>
          )}

          {/* Pest Activity Tab */}
          {activeTab === 'pests' && pestConditions && (
            <div className="space-y-2 sm:space-y-3">
              <div className="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-2 sm:p-3">
                <div className="flex items-start gap-1.5 sm:gap-2">
                  <Info className="w-3.5 h-3.5 sm:w-4 sm:h-4 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
                  <p className="text-[10px] sm:text-xs text-blue-800 dark:text-blue-200">
                    Pest activity is based on recent temperature patterns and seasonal timing. Higher activity means more bugs to deal with.
                  </p>
                </div>
              </div>

              {/* Mosquito Activity */}
              <div className={`rounded-lg p-3 sm:p-4 ${getPestLevelBadgeStyles(pestConditions.mosquito_level).bg} border-2 ${getPestLevelBadgeStyles(pestConditions.mosquito_level).border}`}>
                <div className="flex items-center justify-between mb-1.5 sm:mb-2">
                  <span className="text-xs sm:text-sm font-semibold text-gray-900 dark:text-white">Mosquito Activity</span>
                  <span className={`text-xs sm:text-sm font-bold ${getPestLevelBadgeStyles(pestConditions.mosquito_level).text}`}>
                    {pestConditions.mosquito_level.toUpperCase()}
                  </span>
                </div>
                <p className="text-xs sm:text-sm text-gray-700 dark:text-gray-300">
                  Score: {pestConditions.mosquito_score}/100
                </p>
              </div>

              {/* Outdoor Pest Activity */}
              <div className={`rounded-lg p-3 sm:p-4 ${getPestLevelBadgeStyles(pestConditions.outdoor_pest_level).bg} border-2 ${getPestLevelBadgeStyles(pestConditions.outdoor_pest_level).border}`}>
                <div className="flex items-center justify-between mb-1.5 sm:mb-2">
                  <span className="text-xs sm:text-sm font-semibold text-gray-900 dark:text-white">General Outdoor Pests</span>
                  <span className={`text-xs sm:text-sm font-bold ${getPestLevelBadgeStyles(pestConditions.outdoor_pest_level).text}`}>
                    {pestConditions.outdoor_pest_level.toUpperCase()}
                  </span>
                </div>
                <p className="text-xs sm:text-sm text-gray-700 dark:text-gray-300">
                  Score: {pestConditions.outdoor_pest_score}/100
                </p>
              </div>

              {/* Contributing Factors */}
              {pestConditions.factors.length > 0 && (
                <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-2.5 sm:p-4">
                  <h4 className="text-xs sm:text-sm font-semibold text-gray-900 dark:text-white mb-1.5 sm:mb-2">Contributing Factors</h4>
                  <ul className="space-y-0.5 sm:space-y-1">
                    {pestConditions.factors.map((factor, index) => (
                      <li key={index} className="flex items-start gap-1.5 sm:gap-2 text-[10px] sm:text-xs text-gray-700 dark:text-gray-300">
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
