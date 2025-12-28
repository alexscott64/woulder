import { X, Bug, AlertTriangle } from 'lucide-react';
import {
  PestConditions,
  getPestLevelColor,
  getPestLevelBgColor,
  getPestLevelText,
} from '../utils/pestConditions';

interface PestInfoModalProps {
  pestConditions: PestConditions;
  locationName: string;
  onClose: () => void;
}

export function PestInfoModal({ pestConditions, locationName, onClose }: PestInfoModalProps) {
  const { mosquitoLevel, mosquitoScore, outdoorPestLevel, outdoorPestScore, factors } = pestConditions;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <div className="flex items-center gap-2">
            <Bug className="w-5 h-5 text-amber-600" />
            <h2 className="text-lg font-bold text-gray-900">Pest Activity</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-gray-100 rounded-full transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          <p className="text-sm text-gray-600">
            Pest activity levels for <span className="font-semibold">{locationName}</span> based on current weather and recent conditions.
          </p>

          {/* Mosquito Level */}
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <span className="text-lg">🦟</span>
                <span className="font-semibold text-gray-900">Mosquito Activity</span>
              </div>
              <span className={`font-bold ${getPestLevelColor(mosquitoLevel)}`}>
                {getPestLevelText(mosquitoLevel)}
              </span>
            </div>
            {/* Progress bar */}
            <div className="w-full bg-gray-200 rounded-full h-2.5 mb-2">
              <div
                className={`h-2.5 rounded-full ${getPestLevelBgColor(mosquitoLevel)}`}
                style={{ width: `${mosquitoScore}%` }}
              />
            </div>
            <p className="text-xs text-gray-500">
              Score: {mosquitoScore}/100
            </p>
          </div>

          {/* Outdoor Pest Level */}
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <Bug className="w-5 h-5 text-amber-600" />
                <span className="font-semibold text-gray-900">Outdoor Pests</span>
              </div>
              <span className={`font-bold ${getPestLevelColor(outdoorPestLevel)}`}>
                {getPestLevelText(outdoorPestLevel)}
              </span>
            </div>
            {/* Progress bar */}
            <div className="w-full bg-gray-200 rounded-full h-2.5 mb-2">
              <div
                className={`h-2.5 rounded-full ${getPestLevelBgColor(outdoorPestLevel)}`}
                style={{ width: `${outdoorPestScore}%` }}
              />
            </div>
            <p className="text-xs text-gray-500">
              Score: {outdoorPestScore}/100 • Includes flies, gnats, wasps, ants, etc.
            </p>
          </div>

          {/* Factors */}
          {factors.length > 0 && (
            <div className="border-t border-gray-200 pt-4">
              <div className="flex items-center gap-2 mb-2">
                <AlertTriangle className="w-4 h-4 text-amber-500" />
                <span className="text-sm font-semibold text-gray-700">Contributing Factors</span>
              </div>
              <ul className="space-y-1">
                {factors.map((factor, index) => (
                  <li key={index} className="text-sm text-gray-600 flex items-start gap-2">
                    <span className="text-amber-400 mt-1">•</span>
                    {factor}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Info */}
          <div className="bg-blue-50 rounded-lg p-3 mt-4">
            <p className="text-xs text-blue-800">
              <strong>How it's calculated:</strong> Pest activity is estimated based on temperature (insect metabolism),
              humidity (mosquito survival), recent rainfall (breeding sites), wind conditions, and seasonal patterns.
              Mosquito breeding peaks 7-14 days after rainfall.
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-gray-200">
          <button
            onClick={onClose}
            className="w-full py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium rounded-lg transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
