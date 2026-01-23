import { ChevronRight, MapPin, ChevronLeft, Loader2, AlertCircle, Filter, Sun, Droplets, Clock } from 'lucide-react';
import { useState, useMemo } from 'react';
import { useAreasOrderedByActivity, useSubareasOrderedByActivity, useRoutesOrderedByActivity, useSearchInLocation, useBoulderDryingStatuses, useAreaDryingStats, useMultipleAreaDryingStats } from '../hooks/useClimbActivity';
import { AreaActivitySummary, SearchResult, RouteActivitySummary } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';
import { RouteListItem } from './RouteListItem';
import { AreaConditionCard } from './AreaConditionCard';
import { AdvancedSearchBar } from './AdvancedSearchBar';

interface AreaDrillDownViewProps {
  locationId: number;
  locationName: string;
  searchQuery?: string;
}

interface BreadcrumbItem {
  name: string;
  areaId: string | null; // null for root
}

type FilterStatus = 'all' | 'dry' | 'drying' | 'wet';
type SortOption = 'activity' | 'dry-time';

export function AreaDrillDownView({ locationId, locationName, searchQuery = '' }: AreaDrillDownViewProps) {
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([
    { name: locationName, areaId: null }
  ]);
  const [expandedRoutes, setExpandedRoutes] = useState<Set<string>>(new Set());
  const [filterStatus, setFilterStatus] = useState<FilterStatus>('all');
  const [sortBy, setSortBy] = useState<SortOption>('activity');

  // Current area ID is the last item in breadcrumbs
  const currentAreaId = breadcrumbs[breadcrumbs.length - 1].areaId;

  // Fetch root areas or subareas based on current position
  const { data: rootAreas, isLoading: isLoadingRootAreas, error: rootAreasError } = useAreasOrderedByActivity(locationId);
  const { data: subareas, isLoading: isLoadingSubareas, error: subareasError } = useSubareasOrderedByActivity(
    locationId,
    currentAreaId
  );
  const { data: routes, isLoading: isLoadingRoutes, error: routesError } = useRoutesOrderedByActivity(
    locationId,
    currentAreaId,
    200 // Fetch all routes (max 200)
  );

  // Fetch area-level drying stats when viewing routes in an area
  const { data: areaDryingStats, isLoading: isLoadingDryingStats } = useAreaDryingStats(
    locationId,
    currentAreaId
  );

  // Global search for all areas and routes in location (when search is active)
  const { data: searchResults, isLoading: isSearching, error: searchError } = useSearchInLocation(
    locationId,
    searchQuery,
    200
  );

  // Determine which data to show based on search state
  const isSearchActive = searchQuery.length >= 2;

  // Separate search results into areas and routes with explicit typing
  const searchAreaResults: SearchResult[] = isSearchActive ? (searchResults?.filter(r => r.result_type === 'area') || []) : [];
  const searchRouteResults: SearchResult[] = isSearchActive ? (searchResults?.filter(r => r.result_type === 'route') || []) : [];

  const areas = currentAreaId === null ? rootAreas : subareas;
  const isLoadingAreas = currentAreaId === null ? isLoadingRootAreas : isLoadingSubareas;
  const areasError = currentAreaId === null ? rootAreasError : subareasError;

  // When searching globally, show search results; otherwise show current view
  const filteredAreas: (AreaActivitySummary | SearchResult)[] = isSearchActive ? searchAreaResults : (areas || []);
  const filteredRoutes: (RouteActivitySummary | SearchResult)[] = isSearchActive ? searchRouteResults : (routes || []);
  const isLoadingData = isSearchActive ? isSearching : (isLoadingAreas || isLoadingRoutes);
  const dataError = isSearchActive ? searchError : (areasError || routesError);

  // Show routes if we're in an area that has no subareas OR if search is active
  const showRoutes = isSearchActive || (currentAreaId !== null && (!areas || areas.length === 0));

  // Get area IDs for fetching drying stats
  const areaIdsForDryingStats = useMemo(() => {
    if (!filteredAreas || showRoutes) return [];
    return filteredAreas
      .filter(area => !('result_type' in area)) // Only fetch for non-search results
      .map(area => ('result_type' in area ? area.id : area.mp_area_id));
  }, [filteredAreas, showRoutes]);

  // Fetch drying stats for all visible areas
  const areaDryingStatsQueries = useMultipleAreaDryingStats(locationId, areaIdsForDryingStats);

  // Create a map of area ID to drying stats for easy lookup
  const areaDryingStatsMap = useMemo(() => {
    const map = new Map();
    areaIdsForDryingStats.forEach((id, index) => {
      const query = areaDryingStatsQueries[index];
      if (query?.data) {
        map.set(id, query.data);
      }
    });
    return map;
  }, [areaIdsForDryingStats, areaDryingStatsQueries]);

  // Get route IDs for fetching drying statuses
  const routeIds = useMemo(() => {
    if (!showRoutes || !filteredRoutes) return [];
    return filteredRoutes.map(r => ('result_type' in r ? r.id : r.mp_route_id));
  }, [showRoutes, filteredRoutes]);

  // Fetch drying statuses for all visible routes
  const dryingStatusQueries = useBoulderDryingStatuses(routeIds);

  // Create a map of route ID to drying status for easy lookup
  const dryingStatusMap = useMemo(() => {
    const map = new Map();
    routeIds.forEach((id, index) => {
      const query = dryingStatusQueries[index];
      if (query?.data) {
        map.set(id, query.data);
      }
    });
    return map;
  }, [routeIds, dryingStatusQueries]);

  // Apply filtering and sorting to routes
  const processedRoutes = useMemo(() => {
    if (!filteredRoutes) return [];

    let processed = filteredRoutes.map(routeData => {
      const routeId = 'result_type' in routeData ? routeData.id : routeData.mp_route_id;
      const dryingStatus = dryingStatusMap.get(routeId);
      return { routeData, dryingStatus };
    });

    // Apply filter
    if (filterStatus !== 'all') {
      processed = processed.filter(({ dryingStatus }) => {
        if (!dryingStatus) return false; // Hide routes without drying data when filtering

        if (filterStatus === 'dry') {
          return !dryingStatus.is_wet && dryingStatus.hours_until_dry === 0;
        } else if (filterStatus === 'drying') {
          return dryingStatus.is_wet || (dryingStatus.hours_until_dry > 0 && dryingStatus.hours_until_dry < 48);
        } else if (filterStatus === 'wet') {
          return dryingStatus.is_wet && dryingStatus.hours_until_dry >= 48;
        }
        return true;
      });
    }

    // Apply sorting
    if (sortBy === 'dry-time') {
      processed.sort((a, b) => {
        const hoursA = a.dryingStatus?.hours_until_dry ?? 999;
        const hoursB = b.dryingStatus?.hours_until_dry ?? 999;
        return hoursA - hoursB;
      });
    }
    // 'activity' sort is already applied from API, no need to re-sort

    return processed.map(({ routeData }) => routeData);
  }, [filteredRoutes, filterStatus, sortBy, dryingStatusMap]);

  const handleAreaClick = (area: AreaActivitySummary | SearchResult) => {
    // Handle both AreaActivitySummary and SearchResult types
    let areaId: string;
    if ('result_type' in area) {
      // SearchResult type - use id for areas, mp_area_id for routes
      areaId = area.result_type === 'area' ? area.id : area.mp_area_id;
    } else {
      // AreaActivitySummary type
      areaId = area.mp_area_id;
    }
    setBreadcrumbs([...breadcrumbs, { name: area.name, areaId }]);
    setExpandedRoutes(new Set()); // Reset expanded routes when navigating
  };

  const handleBreadcrumbClick = (index: number) => {
    setBreadcrumbs(breadcrumbs.slice(0, index + 1));
    setExpandedRoutes(new Set()); // Reset expanded routes when navigating
  };

  const toggleRouteExpansion = (routeId: string) => {
    const newExpanded = new Set(expandedRoutes);
    if (newExpanded.has(routeId)) {
      newExpanded.delete(routeId);
    } else {
      newExpanded.add(routeId);
    }
    setExpandedRoutes(newExpanded);
  };

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Breadcrumbs */}
      <nav className="sticky top-0 bg-gray-50 dark:bg-gray-900 px-2.5 sm:px-3 py-2 border-b border-gray-200 dark:border-gray-700 z-10">
        <ol className="flex items-center gap-1.5 text-xs overflow-x-auto">
          {breadcrumbs.map((crumb, index) => (
            <li key={index} className="flex items-center gap-1.5 shrink-0">
              {index > 0 && (
                <ChevronRight className="w-3 h-3 text-gray-400 dark:text-gray-600" />
              )}
              <button
                onClick={() => handleBreadcrumbClick(index)}
                className={`hover:underline transition-colors ${
                  index === breadcrumbs.length - 1
                    ? 'text-gray-900 dark:text-gray-100 font-medium'
                    : 'text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300'
                }`}
                disabled={index === breadcrumbs.length - 1}
              >
                {crumb.name}
              </button>
            </li>
          ))}
        </ol>
      </nav>

      {/* Area Condition Card - Show when viewing routes in an area */}
      {showRoutes && currentAreaId && areaDryingStats && !isLoadingData && !dataError && (
        <div className="bg-gray-50 dark:bg-gray-900 p-3 border-b border-gray-200 dark:border-gray-700">
          <AreaConditionCard
            areaName={breadcrumbs[breadcrumbs.length - 1].name}
            stats={areaDryingStats}
          />
        </div>
      )}

      {/* Advanced Search & Filter Bar - Only show for routes view */}
      {showRoutes && !isLoadingData && !dataError && filteredRoutes && filteredRoutes.length > 0 && (
        <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-3">
          <div className="space-y-3">
            {/* Filter buttons and sort */}
            <div className="flex flex-wrap items-center gap-2">
              <div className="flex items-center gap-1.5">
                <Filter className="w-3.5 h-3.5 text-gray-500 dark:text-gray-400" />
                <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Filter:</span>
              </div>
              <div className="flex gap-1.5">
                {(['all', 'dry', 'drying', 'wet'] as FilterStatus[]).map((status) => (
                  <button
                    key={status}
                    onClick={() => setFilterStatus(status)}
                    className={`px-2 py-1 text-xs rounded transition-colors ${
                      filterStatus === status
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                    }`}
                  >
                    {status.charAt(0).toUpperCase() + status.slice(1)}
                  </button>
                ))}
              </div>
              <div className="ml-auto flex items-center gap-1.5">
                <span className="text-xs font-medium text-gray-700 dark:text-gray-300">Sort:</span>
                <select
                  value={sortBy}
                  onChange={(e) => setSortBy(e.target.value as SortOption)}
                  className="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600"
                >
                  <option value="activity">Recent Activity</option>
                  <option value="dry-time">Hours Until Dry</option>
                </select>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Content Area */}
      <div className="flex-1 overflow-y-scroll custom-scrollbar p-2.5 sm:p-3">
        {/* Loading State */}
        {isLoadingData && (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 text-blue-500 animate-spin" />
          </div>
        )}

        {/* Error State */}
        {dataError && (
          <div className="flex flex-col items-center justify-center py-12 gap-3">
            <AlertCircle className="w-12 h-12 text-red-500" />
            <p className="text-red-600 dark:text-red-400 text-center">
              Failed to load activity data
            </p>
            <button
              onClick={() => window.location.reload()}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm"
            >
              Retry
            </button>
          </div>
        )}

        {/* Areas List */}
        {!isLoadingData && !dataError && filteredAreas && filteredAreas.length > 0 && !showRoutes && (
          <div className="space-y-3">
            {filteredAreas.map((area) => {
              // Handle both AreaActivitySummary and SearchResult types
              const isSearchResult = 'result_type' in area;
              const areaId = isSearchResult ? area.id : area.mp_area_id;
              const totalTicks = isSearchResult ? (area.total_ticks || 0) : area.total_ticks;
              const uniqueRoutes = isSearchResult ? (area.unique_routes || 0) : area.unique_routes;
              const hasSubareas = isSearchResult ? false : area.has_subareas;
              const subareaCount = isSearchResult ? 0 : area.subarea_count;
              const dryingStats = areaDryingStatsMap.get(areaId);

              // Get status color based on drying stats
              const getStatusColor = () => {
                if (!dryingStats) return 'border-gray-200 dark:border-gray-700';
                if (dryingStats.percent_dry >= 80) return 'border-l-4 border-l-green-500';
                if (dryingStats.percent_dry >= 50) return 'border-l-4 border-l-yellow-500';
                return 'border-l-4 border-l-red-500';
              };

              return (
                <button
                  key={areaId}
                  onClick={() => handleAreaClick(area)}
                  className={`w-full p-3 bg-white dark:bg-gray-800 rounded-lg border ${getStatusColor()} hover:shadow-md dark:hover:bg-gray-750 active:bg-gray-50 dark:active:bg-gray-700 transition-all text-left`}
                >
                  <div className="space-y-2.5">
                    {/* Header */}
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex items-center gap-2 min-w-0 flex-1">
                        <MapPin className="w-5 h-5 text-blue-500 flex-shrink-0" />
                        <div className="min-w-0 flex-1">
                          <h3 className="text-base font-bold text-gray-900 dark:text-gray-100 truncate">
                            {area.name}
                          </h3>
                          <div className="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-400 mt-0.5">
                            {hasSubareas ? (
                              <span>{subareaCount} sub-area{subareaCount !== 1 ? 's' : ''}</span>
                            ) : (
                              <span>{uniqueRoutes} boulder{uniqueRoutes !== 1 ? 's' : ''}</span>
                            )}
                            <span>•</span>
                            <span>{totalTicks} tick{totalTicks !== 1 ? 's' : ''}</span>
                            <span>•</span>
                            <span>{formatDaysAgo(area.days_since_climb)}</span>
                          </div>
                        </div>
                      </div>
                      <ChevronRight className="w-5 h-5 text-gray-400 dark:text-gray-600 flex-shrink-0" />
                    </div>

                    {/* Drying Stats */}
                    {dryingStats && (
                      <div className="flex items-center gap-4 pt-2 border-t border-gray-100 dark:border-gray-700">
                        {/* Percent Dry */}
                        <div className="flex items-center gap-1.5">
                          <div className={`w-2 h-2 rounded-full ${
                            dryingStats.percent_dry >= 80 ? 'bg-green-500' :
                            dryingStats.percent_dry >= 50 ? 'bg-yellow-500' :
                            'bg-red-500'
                          }`} />
                          <span className="text-sm font-bold text-gray-900 dark:text-gray-100">
                            {Math.round(dryingStats.percent_dry)}% Dry
                          </span>
                        </div>

                        {/* Dry/Drying/Wet Counts */}
                        <div className="flex items-center gap-3 text-xs text-gray-600 dark:text-gray-400">
                          <div className="flex items-center gap-1">
                            <Sun className="w-3.5 h-3.5 text-green-600 dark:text-green-400" />
                            <span>{dryingStats.dry_count}</span>
                          </div>
                          {dryingStats.drying_count > 0 && (
                            <div className="flex items-center gap-1">
                              <Clock className="w-3.5 h-3.5 text-yellow-600 dark:text-yellow-400" />
                              <span>{dryingStats.drying_count}</span>
                            </div>
                          )}
                          {dryingStats.wet_count > 0 && (
                            <div className="flex items-center gap-1">
                              <Droplets className="w-3.5 h-3.5 text-red-600 dark:text-red-400" />
                              <span>{dryingStats.wet_count}</span>
                            </div>
                          )}
                        </div>

                        {/* Confidence */}
                        {dryingStats.confidence_score > 0 && (
                          <div className="ml-auto text-xs text-gray-500 dark:text-gray-500">
                            {dryingStats.confidence_score}% confident
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        )}

        {/* Routes List */}
        {showRoutes && !isLoadingData && !dataError && processedRoutes && processedRoutes.length > 0 && (
          <div className="space-y-2.5">
            {processedRoutes.map((routeData) => {
              // Convert SearchResult to RouteActivitySummary if needed
              const route: any = 'result_type' in routeData
                ? {
                    mp_route_id: routeData.id,
                    name: routeData.name,
                    rating: routeData.rating || '',
                    mp_area_id: routeData.mp_area_id,
                    last_climb_at: routeData.last_climb_at,
                    most_recent_tick: routeData.most_recent_tick,
                    days_since_climb: routeData.days_since_climb,
                  }
                : routeData;

              return (
                <RouteListItem
                  key={route.mp_route_id}
                  route={route}
                  isExpanded={expandedRoutes.has(route.mp_route_id)}
                  onToggleExpand={() => toggleRouteExpansion(route.mp_route_id)}
                />
              );
            })}
          </div>
        )}

        {/* Empty State - No Search Results */}
        {!isLoadingData && !dataError && isSearchActive && filteredRoutes && filteredRoutes.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 gap-2 text-center">
            <MapPin className="w-12 h-12 text-gray-400 dark:text-gray-600" />
            <p className="text-gray-600 dark:text-gray-400">
              No routes match "{searchQuery}"
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-500">
              Try a different search term
            </p>
          </div>
        )}

        {/* Empty State - No Areas */}
        {!isLoadingData && !dataError && !isSearchActive && areas && areas.length === 0 && !showRoutes && (
          <div className="flex flex-col items-center justify-center py-12 gap-2 text-center">
            <MapPin className="w-12 h-12 text-gray-400 dark:text-gray-600" />
            <p className="text-gray-600 dark:text-gray-400">
              No areas with recent activity
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-500">
              Check back after some climbs are logged
            </p>
          </div>
        )}

        {/* Empty State - No Routes */}
        {showRoutes && !isLoadingData && !dataError && !isSearchActive && (!processedRoutes || processedRoutes.length === 0) && (
          <div className="flex flex-col items-center justify-center py-12 gap-2 text-center">
            <MapPin className="w-12 h-12 text-gray-400 dark:text-gray-600" />
            <p className="text-gray-600 dark:text-gray-400">
              No routes with recent activity in this area
            </p>
            <button
              onClick={() => handleBreadcrumbClick(breadcrumbs.length - 2)}
              className="mt-2 flex items-center gap-1.5 text-sm text-blue-600 dark:text-blue-400 hover:underline"
            >
              <ChevronLeft className="w-4 h-4" />
              Go back
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
