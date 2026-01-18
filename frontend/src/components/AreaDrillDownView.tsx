import { ChevronRight, MapPin, ChevronLeft, Loader2, AlertCircle } from 'lucide-react';
import { useState } from 'react';
import { useAreasOrderedByActivity, useSubareasOrderedByActivity, useRoutesOrderedByActivity } from '../hooks/useClimbActivity';
import { AreaActivitySummary } from '../types/weather';
import { formatDaysAgo } from '../utils/weather/formatters';
import { RouteListItem } from './RouteListItem';

interface AreaDrillDownViewProps {
  locationId: number;
  locationName: string;
}

interface BreadcrumbItem {
  name: string;
  areaId: string | null; // null for root
}

export function AreaDrillDownView({ locationId, locationName }: AreaDrillDownViewProps) {
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
    50
  );

  // Determine which data to show
  const areas = currentAreaId === null ? rootAreas : subareas;
  const isLoadingAreas = currentAreaId === null ? isLoadingRootAreas : isLoadingSubareas;
  const areasError = currentAreaId === null ? rootAreasError : subareasError;

  // Show routes if we're in an area that has no subareas
  const showRoutes = currentAreaId !== null && (!areas || areas.length === 0);

  const handleAreaClick = (area: AreaActivitySummary) => {
    setBreadcrumbs([...breadcrumbs, { name: area.name, areaId: area.mp_area_id }]);
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
    <div className="flex flex-col h-full">
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
      <div className="flex-1 overflow-y-auto p-2.5 sm:p-3">
        {/* Loading State */}
        {(isLoadingAreas || isLoadingRoutes) && (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 text-blue-500 animate-spin" />
          </div>
        )}

        {/* Error State */}
        {(areasError || routesError) && (
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
        {!isLoadingAreas && !areasError && areas && areas.length > 0 && !showRoutes && (
          <div className="space-y-1.5">
            {areas.map((area) => (
              <button
                key={area.mp_area_id}
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
                        {area.has_subareas ? (
                          <span>{area.subarea_count} sub-area{area.subarea_count !== 1 ? 's' : ''}</span>
                        ) : (
                          <span>{area.unique_routes} boulder{area.unique_routes !== 1 ? 's' : ''}</span>
                        )}
                        <span>•</span>
                        <span>{area.total_ticks} tick{area.total_ticks !== 1 ? 's' : ''}</span>
                        <span>•</span>
                        <span>{formatDaysAgo(area.days_since_climb)}</span>
                      </div>
                    </div>
                  </div>
                  <ChevronRight className="w-4 h-4 text-gray-400 dark:text-gray-600 flex-shrink-0" />
                </div>
              </button>
            ))}
          </div>
        )}

        {/* Routes List */}
        {showRoutes && !isLoadingRoutes && !routesError && routes && routes.length > 0 && (
          <div className="space-y-2.5">
            {routes.map((route) => (
              <RouteListItem
                key={route.mp_route_id}
                route={route}
                isExpanded={expandedRoutes.has(route.mp_route_id)}
                onToggleExpand={() => toggleRouteExpansion(route.mp_route_id)}
              />
            ))}
          </div>
        )}

        {/* Empty State - No Areas */}
        {!isLoadingAreas && !areasError && areas && areas.length === 0 && !showRoutes && (
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
        {showRoutes && !isLoadingRoutes && !routesError && (!routes || routes.length === 0) && (
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
