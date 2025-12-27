import React, { useState, useEffect } from 'react';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { weatherApi } from './services/api';
import { WeatherCard } from './components/WeatherCard';
import { ForecastView } from './components/ForecastView';
import { RefreshCw, WifiOff, Wifi, ChevronUp } from 'lucide-react';
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
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [expandedLocationId, setExpandedLocationId] = useState<number | null>(null);

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

  // Fetch all weather data
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ['allWeather'],
    queryFn: async () => {
      const response = await weatherApi.getAllWeather();
      setLastUpdated(new Date());
      return response;
    },
    refetchInterval: 10 * 60 * 1000, // Refetch every 10 minutes
  });

  const handleRefresh = () => {
    refetch();
  };

  // Sort weather data: Skykomish locations first, then Index, then alphabetically
  const sortedWeather = data?.weather.sort((a, b) => {
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
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between flex-wrap gap-4">
            <div className="flex items-center gap-3">
              <img src="/woulder-logo.svg" alt="woulder logo" className="w-12 h-12" />
              <div>
                <h1 className="text-4xl text-gray-900" style={{ fontFamily: "'Righteous', cursive" }}>
                  woulder
                </h1>
              </div>
            </div>

            <div className="flex items-center gap-4">
              {/* Online Status */}
              <div className="flex items-center gap-2">
                {isOnline ? (
                  <>
                    <Wifi className="w-5 h-5 text-green-600" />
                    <span className="text-sm text-gray-700 font-medium">Online</span>
                  </>
                ) : (
                  <>
                    <WifiOff className="w-5 h-5 text-red-600" />
                    <span className="text-sm text-gray-700 font-medium">Offline</span>
                  </>
                )}
              </div>

              {/* Refresh Button */}
              <button
                onClick={handleRefresh}
                disabled={isFetching || !isOnline}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium shadow-sm"
              >
                <RefreshCw className={`w-4 h-4 ${isFetching ? 'animate-spin' : ''}`} />
                <span>{isFetching ? 'Refreshing...' : 'Refresh'}</span>
              </button>
            </div>
          </div>

          {/* Last Updated */}
          {lastUpdated && (
            <div className="mt-3 text-xs text-gray-500">
              Last updated: {format(lastUpdated, 'MMM d, yyyy h:mm:ss a')}
            </div>
          )}
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {isLoading && (
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <RefreshCw className="w-12 h-12 animate-spin text-blue-600 mx-auto mb-4" />
              <p className="text-gray-700 font-medium">Loading weather data...</p>
            </div>
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <p className="text-red-900 font-medium">
              Failed to load weather data. {!isOnline && 'You are currently offline.'}
            </p>
          </div>
        )}

        {sortedWeather && sortedWeather.length > 0 && (
          <>
            {/* Mobile Layout - Single column with inline forecast */}
            <div className="md:hidden space-y-6">
              {sortedWeather.map((forecast) => (
                <div key={forecast.location_id}>
                  <WeatherCard
                    forecast={forecast}
                    isExpanded={expandedLocationId === forecast.location_id}
                    onToggleExpand={(expanded) => setExpandedLocationId(expanded ? forecast.location_id : null)}
                  />
                  {/* Expanded forecast directly after this card on mobile */}
                  {expandedLocationId === forecast.location_id && (
                    <div className="mt-4 bg-white rounded-xl shadow-md border border-gray-200">
                      <div className="p-4 border-b border-gray-200">
                        <h3 className="text-lg font-bold text-gray-900">
                          {forecast.location.name} - 6-Day Forecast
                        </h3>
                      </div>
                      <div className="bg-gray-50 p-4">
                        <ForecastView
                          hourlyData={forecast.hourly || []}
                          currentWeather={forecast.current}
                          historicalData={forecast.historical || []}
                          elevationFt={forecast.location.elevation_ft || 0}
                        />
                      </div>
                      <button
                        onClick={() => setExpandedLocationId(null)}
                        className="w-full px-6 py-3 border-t border-gray-200 flex items-center justify-center gap-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
                      >
                        <ChevronUp className="w-4 h-4" />
                        Hide Forecast
                      </button>
                    </div>
                  )}
                </div>
              ))}
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
                      {expandedInThisRow && expandedLocationId && (
                        <div className="mt-6 bg-white rounded-xl shadow-md border border-gray-200">
                          <div className="p-6 border-b border-gray-200">
                            <h3 className="text-xl font-bold text-gray-900">
                              {sortedWeather.find(f => f.location_id === expandedLocationId)?.location.name} - 6-Day Forecast
                            </h3>
                          </div>
                          <div className="bg-gray-50 p-6">
                            <ForecastView
                              hourlyData={sortedWeather.find(f => f.location_id === expandedLocationId)?.hourly || []}
                              currentWeather={sortedWeather.find(f => f.location_id === expandedLocationId)?.current}
                              historicalData={sortedWeather.find(f => f.location_id === expandedLocationId)?.historical || []}
                              elevationFt={sortedWeather.find(f => f.location_id === expandedLocationId)?.location.elevation_ft || 0}
                            />
                          </div>
                          <button
                            onClick={() => setExpandedLocationId(null)}
                            className="w-full px-6 py-3 border-t border-gray-200 flex items-center justify-center gap-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
                          >
                            <ChevronUp className="w-4 h-4" />
                            Hide Forecast
                          </button>
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

        {!isLoading && !error && (!data || data.weather.length === 0) && (
          <div className="text-center py-12">
            <p className="text-gray-700 font-medium">No weather data available</p>
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <p className="text-center text-sm text-gray-600">
            woulder - Weather dashboard for climbers
          </p>
        </div>
      </footer>
    </div>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Dashboard />
    </QueryClientProvider>
  );
}

export default App;
