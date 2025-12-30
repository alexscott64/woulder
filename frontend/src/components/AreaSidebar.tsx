import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { MapPin, Globe, X, Check } from 'lucide-react';
import { weatherApi } from '../services/api';
import { useSettings } from '../contexts/SettingsContext';

export default function AreaSelector() {
  const { settings, setSelectedArea } = useSettings();
  const [isOpen, setIsOpen] = useState(false);

  // Fetch areas with location counts
  const { data: areasData, isLoading } = useQuery({
    queryKey: ['areas'],
    queryFn: weatherApi.getAreas,
  });

  const areas = areasData?.areas || [];
  const totalLocations = areas.reduce((sum, area) => sum + area.location_count, 0);

  const handleAreaClick = (areaId: number | null) => {
    setSelectedArea(areaId);
    setIsOpen(false);
  };

  const selectedAreaName = settings.selectedAreaId
    ? areas.find(a => a.id === settings.selectedAreaId)?.name
    : 'All Areas';

  return (
    <>
      {/* Compact Selector Button - designed to go in header */}
      <button
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-2 px-3 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
        title="Select climbing area"
      >
        {settings.selectedAreaId === null ? (
          <Globe className="w-5 h-5 text-blue-600 dark:text-blue-400" />
        ) : (
          <MapPin className="w-5 h-5 text-blue-600 dark:text-blue-400" />
        )}
        <div className="hidden md:flex flex-col items-start">
          <span className="text-xs text-gray-500 dark:text-gray-400">Area</span>
          <span className="text-sm font-medium leading-none">
            {isLoading ? '...' : selectedAreaName}
          </span>
        </div>
        {!isLoading && (
          <span className="px-1.5 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-xs font-semibold rounded">
            {settings.selectedAreaId === null
              ? totalLocations
              : areas.find(a => a.id === settings.selectedAreaId)?.location_count || 0}
          </span>
        )}
      </button>

      {/* Slide-out Panel / Modal */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-40 animate-fadeIn"
            onClick={() => setIsOpen(false)}
          />

          {/* Panel */}
          <div className="fixed inset-y-0 right-0 w-full sm:w-96 bg-white dark:bg-gray-800 shadow-2xl z-50 animate-slideInRight flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <div>
                <h2 className="text-lg font-bold text-gray-900 dark:text-white">Select Area</h2>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
                  Choose a climbing region
                </p>
              </div>
              <button
                onClick={() => setIsOpen(false)}
                className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                aria-label="Close"
              >
                <X className="w-5 h-5 text-gray-500 dark:text-gray-400" />
              </button>
            </div>

            {/* Area List */}
            <div className="flex-1 overflow-y-auto">
              {isLoading ? (
                <div className="p-4 space-y-3">
                  {[1, 2, 3, 4].map((i) => (
                    <div
                      key={i}
                      className="h-16 bg-gray-100 dark:bg-gray-700 rounded-lg animate-pulse"
                    />
                  ))}
                </div>
              ) : (
                <div className="p-4 space-y-2">
                  {/* All Areas Option */}
                  <button
                    onClick={() => handleAreaClick(null)}
                    className={`w-full flex items-center gap-4 p-4 rounded-lg transition-all ${
                      settings.selectedAreaId === null
                        ? 'bg-blue-50 dark:bg-blue-900/20 border-2 border-blue-500'
                        : 'bg-gray-50 dark:bg-gray-700/50 border-2 border-transparent hover:border-gray-300 dark:hover:border-gray-600'
                    }`}
                  >
                    <div className={`p-2 rounded-lg ${
                      settings.selectedAreaId === null
                        ? 'bg-blue-500'
                        : 'bg-gray-200 dark:bg-gray-600'
                    }`}>
                      <Globe className={`w-5 h-5 ${
                        settings.selectedAreaId === null
                          ? 'text-white'
                          : 'text-gray-600 dark:text-gray-300'
                      }`} />
                    </div>
                    <div className="flex-1 text-left">
                      <div className={`font-semibold ${
                        settings.selectedAreaId === null
                          ? 'text-blue-900 dark:text-blue-100'
                          : 'text-gray-900 dark:text-white'
                      }`}>
                        All Areas
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">
                        {totalLocations} locations across all regions
                      </div>
                    </div>
                    {settings.selectedAreaId === null && (
                      <Check className="w-5 h-5 text-blue-600 dark:text-blue-400 flex-shrink-0" />
                    )}
                  </button>

                  {/* Individual Areas */}
                  {areas.map((area) => (
                    <button
                      key={area.id}
                      onClick={() => handleAreaClick(area.id)}
                      className={`w-full flex items-center gap-4 p-4 rounded-lg transition-all ${
                        settings.selectedAreaId === area.id
                          ? 'bg-blue-50 dark:bg-blue-900/20 border-2 border-blue-500'
                          : 'bg-gray-50 dark:bg-gray-700/50 border-2 border-transparent hover:border-gray-300 dark:hover:border-gray-600'
                      }`}
                    >
                      <div className={`p-2 rounded-lg ${
                        settings.selectedAreaId === area.id
                          ? 'bg-blue-500'
                          : 'bg-gray-200 dark:bg-gray-600'
                      }`}>
                        <MapPin className={`w-5 h-5 ${
                          settings.selectedAreaId === area.id
                            ? 'text-white'
                            : 'text-gray-600 dark:text-gray-300'
                        }`} />
                      </div>
                      <div className="flex-1 text-left">
                        <div className={`font-semibold ${
                          settings.selectedAreaId === area.id
                            ? 'text-blue-900 dark:text-blue-100'
                            : 'text-gray-900 dark:text-white'
                        }`}>
                          {area.name}
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          {area.location_count} location{area.location_count !== 1 ? 's' : ''}
                          {area.region && ` • ${area.region}`}
                        </div>
                      </div>
                      {settings.selectedAreaId === area.id && (
                        <Check className="w-5 h-5 text-blue-600 dark:text-blue-400 flex-shrink-0" />
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Footer (optional - for future features) */}
            <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50">
              <p className="text-xs text-gray-500 dark:text-gray-400 text-center">
                More areas coming soon
              </p>
            </div>
          </div>
        </>
      )}

      <style>{`
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        @keyframes slideInRight {
          from { transform: translateX(100%); }
          to { transform: translateX(0); }
        }
        .animate-fadeIn {
          animation: fadeIn 0.2s ease-out;
        }
        .animate-slideInRight {
          animation: slideInRight 0.3s ease-out;
        }
      `}</style>
    </>
  );
}
