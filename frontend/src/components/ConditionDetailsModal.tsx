import { X } from 'lucide-react';

import { ConditionLevel, RockTemperatureStatus } from '../types/weather';
import { SendWindowDayView } from './weather/SendWindowDayView';

interface ConditionDetailsModalProps {
  locationName: string;
  conditionLevel: ConditionLevel;
  conditionLabel: string;
  reasons: string[];
  onClose: () => void;
  /**
   * Surface-temperature/friction status for the area. When provided
   * alongside `targetDate`, a "Send Windows" section is rendered below
   * the Contributing Factors list with a single-day Gantt strip + DayCard
   * for that date. Reuses the same data source that powers the Rock Temp
   * tab in `ConditionsModal`, so no extra API calls are made.
   */
  rockTempStatus?: RockTemperatureStatus;
  /** Target day in local YYYY-MM-DD form (matches `DailyRockTemp.local_date`). */
  targetDate?: string;
}

export function ConditionDetailsModal({
  locationName,
  conditionLevel,
  conditionLabel,
  reasons,
  onClose,
  rockTempStatus,
  targetDate,
}: ConditionDetailsModalProps) {
  const getBadgeColor = (level: ConditionLevel) => {
    switch (level) {
      case 'good':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 border-green-300 dark:border-green-700';
      case 'marginal':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 border-yellow-300 dark:border-yellow-700';
      case 'bad':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200 border-red-300 dark:border-red-700';
      case 'do_not_climb':
        return 'bg-red-200 text-red-900 dark:bg-red-900 dark:text-red-100 border-red-500 dark:border-red-600';
    }
  };

  // Only show the Send Windows section if we have both a status source
  // and a target date. The wrapper itself handles "no windows" gracefully
  // via DayCard, so we don't need to pre-filter here.
  const showSendWindows = !!rockTempStatus && !!targetDate;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 dark:bg-opacity-70 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
              Climbing Conditions
            </h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{locationName}</p>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition-colors"
            aria-label="Close"
          >
            <X size={24} />
          </button>
        </div>

        {/* Overall Condition Badge */}
        <div className="p-6">
          <div className="flex items-center justify-center mb-6">
            <div className={`px-6 py-3 rounded-full border-2 font-semibold text-lg ${getBadgeColor(conditionLevel)}`}>
              {conditionLabel}
            </div>
          </div>

          {/* Contributing Factors */}
          <div>
            <h3 className="text-base font-semibold text-gray-900 dark:text-white mb-3">
              Contributing Factors:
            </h3>
            <ul className="space-y-2">
              {reasons.map((reason, index) => (
                <li
                  key={index}
                  className="flex items-start gap-2 text-sm text-gray-700 dark:text-gray-300"
                >
                  <span className="text-gray-400 dark:text-gray-600 mt-0.5">•</span>
                  <span>{reason}</span>
                </li>
              ))}
            </ul>
          </div>

          {/* Send Windows — single-day Gantt strip + DayCard for this day,
              wired to the same surface-temperature/friction data that
              powers the Rock Temp tab. */}
          {showSendWindows && (
            <div className="mt-6 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
              <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">
                Send Windows
              </h4>
              <SendWindowDayView status={rockTempStatus!} date={targetDate!} />
            </div>
          )}

          {/* Guidelines */}
          <div className="mt-6 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
            <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-2">
              Condition Guidelines
            </h4>
            <div className="space-y-2 text-xs text-gray-600 dark:text-gray-400">
              <div>
                <span className="font-medium text-green-700 dark:text-green-400">Good:</span> Dry, low winds (&lt;12 mph), ideal temps (41-65°F), normal humidity (&lt;85%)
              </div>
              <div>
                <span className="font-medium text-yellow-700 dark:text-yellow-400">Fair:</span> Light rain (0.05-0.1"), moderate winds (12-20 mph), marginal temps (30-40°F or 66-79°F), high humidity (&gt;85%)
              </div>
              <div>
                <span className="font-medium text-red-700 dark:text-red-400">Poor:</span> Heavy rain (&gt;0.1"), high winds (&gt;20 mph), extreme temps (&lt;30°F or &gt;79°F)
              </div>
              <div>
                <span className="font-medium text-red-900 dark:text-red-300">Do Not Climb:</span> Wet-sensitive rock (sandstone, arkose, graywacke) is currently wet and climbing will cause permanent damage
              </div>
            </div>
          </div>
        </div>

        {/* Close Button */}
        <div className="flex justify-end gap-3 p-6 pt-0">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
