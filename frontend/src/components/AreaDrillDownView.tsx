import { ChevronRight, MapPin, ChevronLeft, Loader2, AlertCircle } from 'lucide-react';
import { useState } from 'react';
import { useAreasOrderedByActivity, useSubareasOrderedByActivity, useRoutesOrderedByActivity, useSearchInLocation } from '../hooks/useClimbActivity';
import { AreaActivitySummary, SearchResult } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';
import { RouteListItem } from './RouteListItem';

interface AreaDrillDownViewProps {
  locationId: number;
  locationName: string;
  searchQuery?: string;
}

interface BreadcrumbItem {
  name: string;
  areaId: string | null; // null for root
}

export function AreaDrillDownView({ locationId, locationName, searchQuery = '' }: AreaDrillDownViewProps) {
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([
    { name: locationName, areaId: null }
  ]);
  const [expandedRoutes, setExpandedRoutes] = useState<Set<string>>(new Set());

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

  // Global search for all areas and routes in location (when search is active)
  const { data: searchResults, isLoading: isSearching, error: searchError } = useSearchInLocation(
    locationId,
    searchQuery,
    200
  );

  // Determine which data to show based on search state
  const isSearchActive = searchQuery.length >= 2;

  // Separate search results into areas and routes
  const searchAreaResults = isSearchActive ? searchResults?.filter(r => r.result_type === 'area') || [] : [];
  const searchRouteResults = isSearchActive ? searchResults?.filter(r => r.result_type === 'route') || [] : [];

  const areas = currentAreaId === null ? rootAreas : subareas;
  const isLoadingAreas = currentAreaId === null ? isLoadingRootAreas : isLoadingSubareas;
  const areasError = currentAreaId === null ? rootAreasError : subareasError;

  // When searching globally, show search results; otherwise show current view
  const filteredAreas = isSearchActive ? searchAreaResults : areas;
  const filteredRoutes = isSearchActive ? searchRouteResults : routes;
  const isLoadingData = isSearchActive ? isSearching : (isLoadingAreas || isLoadingRoutes);
  const dataError = isSearchActive ? searchError : (areasError || routesError);

  // Show routes if we're in an area that has no subareas OR if search is active
  const showRoutes = isSearchActive || (currentAreaId !== null && (!areas || areas.length === 0));

  // Show areas in search results if any match
  const showSearchAreas = isSearchActive && searchAreaResults.length > 0;

  const handleAreaClick = (area: AreaActivitySummary | SearchResult) => {
    // Handle both AreaActivitySummary and SearchResult types
    const areaId = 'mp_area_id' in area ? area.mp_area_id : area.result_type === 'area' ? area.id : area.mp_area_id;
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
          <div className="space-y-1.5">
            {filteredAreas.map((area) => {
              // Handle both AreaActivitySummary and SearchResult types
              const isSearchResult = 'result_type' in area;
              const areaId = isSearchResult ? area.id : area.mp_area_id;
              const totalTicks = isSearchResult ? (area.total_ticks || 0) : area.total_ticks;
              const uniqueRoutes = isSearchResult ? (area.unique_routes || 0) : area.unique_routes;
              const hasSubareas = isSearchResult ? false : area.has_subareas;
              const subareaCount = isSearchResult ? 0 : area.subarea_count;

              return (
                <button
                  key={areaId}
                  onClick={() => handleAreaClick(area)}
                  className="w-full p-2.5 bg-white dark:bg-gray-800 rounded-md border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-750 active:bg-gray-100 dark:active:bg-gray-700 transition-colors text-left"
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0 flex-1">
                      <MapPin className="w-4 h-4 text-blue-500 flex-shrink-0" />
                      <div className="min-w-0 flex-1">
                        <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">
                          {area.name}
                        </h3>
                        <div className="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-400">
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
                    <ChevronRight className="w-4 h-4 text-gray-400 dark:text-gray-600 flex-shrink-0" />
                  </div>
                </button>
              );
            })}
          </div>
        )}

        {/* Routes List */}
        {showRoutes && !isLoadingData && !dataError && filteredRoutes && filteredRoutes.length > 0 && (
          <div className="space-y-2.5">
            {filteredRoutes.map((routeData) => {
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
        {showRoutes && !isLoadingData && !dataError && !isSearchActive && (!filteredRoutes || filteredRoutes.length === 0) && (
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
