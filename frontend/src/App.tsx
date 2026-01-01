import React, { useState, useEffect } from 'react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { weatherApi } from './services/api';
import { WeatherCard } from './components/WeatherCard';
import { ForecastView } from './components/ForecastView';
import { SettingsModal } from './components/SettingsModal';
import AreaSelector from './components/AreaSidebar';
import { SettingsProvider, useSettings } from './contexts/SettingsContext';
import { ConditionCalculator } from './utils/weather/analyzers';
import { getConditionColor } from './components/weather/weatherDisplay';
import { RefreshCw, WifiOff, ChevronUp, Settings, Github, Heart, Mail } from 'lucide-react';
import { format } from 'date-fns';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes (renamed from cacheTime)
      refetchOnWindowFocus: true,
    },
  },
});

function Dashboard() {
  const { settings } = useSettings();
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [expandedLocationId, setExpandedLocationId] = useState<number | null>(null);
  const [showSettings, setShowSettings] = useState(false);

  // Monitor online/offline status
  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  // Fetch all weather data (filtered by selected area)
  const { data, isLoading, error } = useQuery({
    queryKey: ['allWeather', settings.selectedAreaId],
    queryFn: async () => {
      const response = await weatherApi.getAllWeather(settings.selectedAreaId);
      setLastUpdated(new Date());
      return response;
    },
    refetchInterval: 10 * 60 * 1000, // Refetch every 10 minutes
  });

  // Sort weather data: Skykomish locations first, then Index, then alphabetically
  const sortedWeather = data?.forecasts.sort((a, b) => {
    // Skykomish - Money Creek first
    if (a.location.name === 'Skykomish - Money Creek') return -1;
    if (b.location.name === 'Skykomish - Money Creek') return 1;

    // Skykomish - Paradise second
    if (a.location.name === 'Skykomish - Paradise') return -1;
    if (b.location.name === 'Skykomish - Paradise') return 1;

    // Index third
    if (a.location.name === 'Index') return -1;
    if (b.location.name === 'Index') return 1;

    // Rest alphabetically
    return a.location.name.localeCompare(b.location.name);
  });

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Header */}
      <header className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between gap-2 sm:gap-4">
            <div className="flex items-center gap-2 sm:gap-3">
              <img src="/woulder-logo.svg" alt="woulder logo" className="w-10 h-10 sm:w-12 sm:h-12" />
              <div>
                <h1 className="text-3xl sm:text-4xl text-gray-900 dark:text-white" style={{ fontFamily: "'Righteous', cursive" }}>
                  woulder
                </h1>
              </div>
            </div>

            <div className="flex items-center gap-2 sm:gap-4">
              {/* Offline Status - Only show when offline */}
              {!isOnline && (
                <div className="flex items-center" title="Offline">
                  <WifiOff className="w-5 h-5 text-red-600 dark:text-red-400" />
                </div>
              )}

              {/* Area Selector */}
              <AreaSelector />

              {/* Settings Button */}
              <button
                onClick={() => setShowSettings(true)}
                className="p-2 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                title="Settings"
              >
                <Settings className="w-5 h-5" />
              </button>
            </div>
          </div>

          {/* Last Updated */}
          {lastUpdated && (
            <div className="mt-3 text-xs text-gray-500 dark:text-gray-400">
              Last updated: {format(lastUpdated, 'MMM d, yyyy h:mm:ss a')} {new Intl.DateTimeFormat('en-US', { timeZoneName: 'short' }).formatToParts(lastUpdated).find(part => part.type === 'timeZoneName')?.value}
            </div>
          )}
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {isLoading && (
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <RefreshCw className="w-12 h-12 animate-spin text-blue-600 dark:text-blue-400 mx-auto mb-4" />
              <p className="text-gray-700 dark:text-gray-300 font-medium">Loading weather data...</p>
            </div>
          </div>
        )}

        {error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <p className="text-red-900 dark:text-red-200 font-medium">
              Failed to load weather data. {!isOnline && 'You are currently offline.'}
            </p>
          </div>
        )}

        {sortedWeather && sortedWeather.length > 0 && (
          <>
            {/* Mobile Layout - Single column with inline forecast */}
            <div className="md:hidden space-y-6">
              {/* Mobile Region Header - Clickable Area Selector */}
              <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
                <AreaSelector variant="mobile" />
              </div>
              {sortedWeather.map((forecast) => {
                const isExpanded = expandedLocationId === forecast.location_id;
                const condition = ConditionCalculator.calculateCondition(forecast.current, forecast.historical, forecast.rock_drying_status);
                const conditionColor = getConditionColor(condition.level);

                return (
                  <div key={forecast.location_id} className={isExpanded ? 'shadow-lg rounded-xl' : ''}>
                    <WeatherCard
                      forecast={forecast}
                      isExpanded={isExpanded}
                      onToggleExpand={(expanded) => setExpandedLocationId(expanded ? forecast.location_id : null)}
                    />
                    {/* Expanded forecast - seamlessly connected to card */}
                    {isExpanded && (
                      <div className="bg-white dark:bg-gray-800 rounded-b-xl border border-t-0 border-gray-200 dark:border-gray-700 overflow-hidden">
                        {/* Colored accent bar showing which card this belongs to */}
                        <div className={`h-1 ${conditionColor}`} />
                        <div className="bg-gray-50 dark:bg-gray-900 p-4">
                          <ForecastView
                            hourlyData={forecast.hourly || []}
                            currentWeather={forecast.current}
                            historicalData={forecast.historical || []}
                            elevationFt={forecast.location.elevation_ft || 0}
                            dailySunTimes={forecast.daily_sun_times}
                          />
                        </div>
                        <button
                          onClick={() => setExpandedLocationId(null)}
                          className={`w-full px-6 py-3 border-t border-gray-200 dark:border-gray-700 flex items-center justify-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors ${conditionColor} bg-opacity-10`}
                        >
                          <ChevronUp className="w-4 h-4" />
                          Hide Forecast
                        </button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>

            {/* Desktop Layout - Grid with row-based forecast */}
            <div className="hidden md:block space-y-6">
              {(() => {
                const rows: React.JSX.Element[] = [];
                const expandedIndex = sortedWeather.findIndex(f => f.location_id === expandedLocationId);

                // Group cards into rows of 3 for lg, 2 for md
                for (let i = 0; i < sortedWeather.length; i += 3) {
                  const rowForecasts = sortedWeather.slice(i, i + 3);
                  const rowNumber = i / 3;
                  const expandedInThisRow = expandedIndex >= i && expandedIndex < i + 3;
                  const expandedForecast = expandedInThisRow ? sortedWeather.find(f => f.location_id === expandedLocationId) : null;
                  const expandedCondition = expandedForecast ? ConditionCalculator.calculateCondition(expandedForecast.current, expandedForecast.historical, expandedForecast.rock_drying_status) : null;
                  const expandedConditionColor = expandedCondition ? getConditionColor(expandedCondition.level) : '';
                  // Calculate position of expanded card within row (0, 1, or 2)
                  const expandedPositionInRow = expandedIndex >= 0 ? expandedIndex - i : -1;

                  rows.push(
                    <div key={`row-${rowNumber}`}>
                      {/* Card row */}
                      <div className="grid grid-cols-2 lg:grid-cols-3 gap-6">
                        {rowForecasts.map((forecast) => (
                          <WeatherCard
                            key={forecast.location_id}
                            forecast={forecast}
                            isExpanded={expandedLocationId === forecast.location_id}
                            onToggleExpand={(expanded) => setExpandedLocationId(expanded ? forecast.location_id : null)}
                          />
                        ))}
                      </div>

                      {/* Expanded forecast after this row */}
                      {expandedInThisRow && expandedForecast && (
                        <div className="relative mt-4">
                          {/* Arrow indicator pointing to the active card */}
                          <div
                            className="absolute -top-2 w-4 h-4 bg-white dark:bg-gray-800 border-l border-t border-gray-200 dark:border-gray-700 transform rotate-45 z-10"
                            style={{
                              left: expandedPositionInRow === 0 ? 'calc(16.67% - 8px)' :
                                    expandedPositionInRow === 1 ? 'calc(50% - 8px)' :
                                    'calc(83.33% - 8px)'
                            }}
                          />
                          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
                            {/* Colored accent bar */}
                            <div className={`h-1.5 ${expandedConditionColor}`} />
                            {/* Header with location name */}
                            <div className="px-6 py-4 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
                              <div className="flex items-center gap-3">
                                <div className={`w-3 h-3 rounded-full ${expandedConditionColor}`} />
                                <h3 className="text-lg font-bold text-gray-900 dark:text-white">
                                  {expandedForecast.location.name}
                                </h3>
                              </div>
                              <span className="text-sm text-gray-500 dark:text-gray-400">6-Day Forecast</span>
                            </div>
                            <div className="bg-gray-50 dark:bg-gray-900 p-6">
                              <ForecastView
                                hourlyData={expandedForecast.hourly || []}
                                currentWeather={expandedForecast.current}
                                historicalData={expandedForecast.historical || []}
                                elevationFt={expandedForecast.location.elevation_ft || 0}
                                dailySunTimes={expandedForecast.daily_sun_times}
                              />
                            </div>
                            <button
                              onClick={() => setExpandedLocationId(null)}
                              className={`w-full px-6 py-3 border-t border-gray-200 dark:border-gray-700 flex items-center justify-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors`}
                            >
                              <ChevronUp className="w-4 h-4" />
                              Hide Forecast
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  );
                }

                return rows;
              })()}
            </div>
          </>
        )}

        {!isLoading && !error && (!data || data.forecasts.length === 0) && (
          <div className="text-center py-12">
            <p className="text-gray-700 dark:text-gray-300 font-medium">No weather data available</p>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex flex-col items-center gap-4">
            {/* Tagline */}
            <p className="text-center text-gray-900 dark:text-white font-medium">
              Better forecasts for bouldering
            </p>

            {/* Links */}
            <div className="flex flex-wrap items-center justify-center gap-4 sm:gap-6 text-sm">
              <a
                href="mailto:woulder.pnw@gmail.com"
                className="flex items-center gap-1.5 text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
              >
                <Mail className="w-4 h-4" />
                <span>Contact</span>
              </a>
              <a
                href="https://github.com/alexscott64/woulder"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1.5 text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
              >
                <Github className="w-4 h-4" />
                <span>GitHub</span>
              </a>
            </div>

            {/* Open Source Badge */}
            <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-100 dark:bg-gray-700 rounded-full">
              <Heart className="w-4 h-4 text-red-500" />
              <span className="text-xs font-medium text-gray-700 dark:text-gray-300">
                Always free, always open source
              </span>
            </div>
          </div>
        </div>
      </footer>

      {/* Settings Modal */}
      {showSettings && <SettingsModal onClose={() => setShowSettings(false)} />}
    </div>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <SettingsProvider>
        <Dashboard />
      </SettingsProvider>
    </QueryClientProvider>
  );
}

export default App;
