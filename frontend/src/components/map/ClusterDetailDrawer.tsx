import React, { useState, useMemo, useEffect } from 'react';
import { X, TrendingUp, Users, MapPin, Search, ChevronDown, ChevronRight } from 'lucide-react';
import { HeatMapPoint, RouteActivity } from '../../types/heatmap';
import { heatMapApi } from '../../services/api';
import { formatDistanceToNow } from 'date-fns';

interface ClusterDetailDrawerProps {
  areas: HeatMapPoint[];
  isOpen: boolean;
  onClose: () => void;
  onAreaClick: (areaId: number) => void;
  dateRange: { start: Date; end: Date };
}

export function ClusterDetailDrawer({ areas, isOpen, onClose, onAreaClick, dateRange }: ClusterDetailDrawerProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<RouteActivity[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [expandedAreaIds, setExpandedAreaIds] = useState<Set<number>>(new Set());
  
  // Reset search when drawer closes
  useEffect(() => {
    if (!isOpen) {
      setSearchQuery('');
      setSearchResults([]);
      setExpandedAreaIds(new Set());
    }
  }, [isOpen]);

  // Perform route search when query changes
  useEffect(() => {
    const searchRoutes = async () => {
      if (!searchQuery.trim() || areas.length === 0) {
        setSearchResults([]);
        return;
      }

      setIsSearching(true);
      try {
        const areaIds = areas.map(a => a.mp_area_id);
        const response = await heatMapApi.searchRoutesInCluster({
          areaIds,
          query: searchQuery,
          startDate: dateRange.start,
          endDate: dateRange.end,
          limit: 200,
        });
        setSearchResults(response.routes);
      } catch (error) {
        console.error('Failed to search routes:', error);
        setSearchResults([]);
      } finally {
        setIsSearching(false);
      }
    };

    // Debounce search
    const timeoutId = setTimeout(searchRoutes, 300);
    return () => clearTimeout(timeoutId);
  }, [searchQuery, areas, dateRange]);

  // Filter areas by name
  const filteredAreas = useMemo(() => {
    if (!isOpen || areas.length === 0) return [];
    
    if (!searchQuery.trim()) {
      // No search - return all areas sorted by activity
      return [...areas].sort((a, b) => b.activity_score - a.activity_score);
    }

    // When searching, only show areas that have matching routes
    const areasWithMatches = new Set(searchResults.map(r => r.mp_area_id));
    
    // Also include areas whose names match the search
    const query = searchQuery.toLowerCase();
    const matchingAreas = areas.filter(area => 
      area.name.toLowerCase().includes(query) || areasWithMatches.has(area.mp_area_id)
    );

    return matchingAreas.sort((a, b) => b.activity_score - a.activity_score);
  }, [areas, searchQuery, searchResults, isOpen]);

  // Group routes by area
  const routesByArea = useMemo(() => {
    const map = new Map<number, RouteActivity[]>();
    searchResults.forEach(route => {
      const routes = map.get(route.mp_area_id) || [];
      routes.push(route);
      map.set(route.mp_area_id, routes);
    });
    return map;
  }, [searchResults]);

  // Calculate aggregate stats
  const totalTicks = useMemo(() => areas.reduce((sum, a) => sum + a.total_ticks, 0), [areas]);
  const totalRoutes = useMemo(() => areas.reduce((sum, a) => sum + a.active_routes, 0), [areas]);
  const totalClimbers = useMemo(() => areas.reduce((sum, a) => sum + a.unique_climbers, 0), [areas]);
  const totalActivityScore = useMemo(() => areas.reduce((sum, a) => sum + a.activity_score, 0), [areas]);
  
  const toggleAreaExpansion = (areaId: number) => {
    setExpandedAreaIds(prev => {
      const next = new Set(prev);
      if (next.has(areaId)) {
        next.delete(areaId);
      } else {
        next.add(areaId);
      }
      return next;
    });
  };

  if (!isOpen || areas.length === 0) return null;

  return (
    <>
      {/* Overlay */}
      <div
        className="fixed inset-0 bg-black bg-opacity-30 z-40 transition-opacity md:hidden"
        onClick={onClose}
      />

      {/* Drawer */}
      <div className="fixed right-0 top-0 bottom-0 w-full sm:w-96 lg:w-[480px] bg-white dark:bg-gray-800 z-50 shadow-2xl overflow-y-auto transform transition-transform">
        {/* Header */}
        <div className="sticky top-0 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-3 sm:p-4 z-10">
          <div className="flex items-center justify-between gap-3">
            <h2 className="text-lg sm:text-xl font-bold text-gray-900 dark:text-white">
              {areas.length} Climbing {areas.length === 1 ? 'Area' : 'Areas'} in this Region
            </h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors shrink-0"
              title="Close"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="p-3 sm:p-4 space-y-4 sm:space-y-6">
          {/* Aggregate Stats Cards */}
          <div className="grid grid-cols-2 gap-2 sm:gap-3">
            <StatCard
              icon={<TrendingUp className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Total Activity"
              value={totalActivityScore}
            />
            <StatCard
              icon={<TrendingUp className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Total Ticks"
              value={totalTicks}
            />
            <StatCard
              icon={<MapPin className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Active Routes"
              value={totalRoutes}
            />
            <StatCard
              icon={<Users className="w-4 h-4 sm:w-5 sm:h-5" />}
              label="Unique Climbers"
              value={totalClimbers}
            />
          </div>

          {/* Search Bar with improved styling */}
          <div className="mb-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search areas and routes..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-3 border-2 border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all"
              />
            </div>
            <div className="flex items-center justify-between mt-2">
              {searchQuery ? (
                <>
                  {isSearching ? (
                    <p className="text-xs text-gray-600 dark:text-gray-400">
                      Searching...
                    </p>
                  ) : (
                    <p className="text-xs text-gray-600 dark:text-gray-400">
                      <span className="font-semibold">{filteredAreas.length}</span> areas, <span className="font-semibold">{searchResults.length}</span> routes match
                    </p>
                  )}
                  <button
                    onClick={() => setSearchQuery('')}
                    className="text-xs text-blue-600 dark:text-blue-400 hover:underline"
                  >
                    Clear
                  </button>
                </>
              ) : (
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  💡 Tip: Search by area name or route name
                </p>
              )}
            </div>
          </div>

          {/* Areas List with Routes */}
          <div>
            <h3 className="text-base font-semibold text-gray-900 dark:text-white mb-3">
              {searchQuery ? 'Search Results' : 'All Areas in this Cluster'}
            </h3>
            <div className="space-y-2">
              {filteredAreas.length > 0 ? (
                filteredAreas.map((area, index) => {
                  const daysSince = (Date.now() - new Date(area.last_activity).getTime()) / (1000 * 60 * 60 * 24);
                  const colorClass = daysSince <= 7 ? 'bg-red-500'
                    : daysSince <= 30 ? 'bg-orange-500'
                    : daysSince <= 90 ? 'bg-yellow-500'
                    : 'bg-blue-500';
                  
                  const areaRoutes = routesByArea.get(area.mp_area_id) || [];
                  const hasMatchingRoutes = areaRoutes.length > 0;
                  const isExpanded = expandedAreaIds.has(area.mp_area_id);
                  const query = searchQuery.toLowerCase();
                  const areaNameMatches = area.name.toLowerCase().includes(query);

                  return (
                    <div key={area.mp_area_id} className="bg-gray-50 dark:bg-gray-900 rounded-lg overflow-hidden">
                      {/* Area Header */}
                      <div className="flex items-start gap-3 p-3">
                        {/* Expand/Collapse button - only show if there are matching routes */}
                        {hasMatchingRoutes && searchQuery && (
                          <button
                            onClick={() => toggleAreaExpansion(area.mp_area_id)}
                            className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors shrink-0 mt-0.5"
                            title={isExpanded ? "Collapse routes" : "Expand routes"}
                          >
                            {isExpanded ? (
                              <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                            ) : (
                              <ChevronRight className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                            )}
                          </button>
                        )}

                        {/* Area Content */}
                        <button
                          onClick={() => onAreaClick(area.mp_area_id)}
                          className="flex-1 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors text-left p-2 -m-2"
                        >
                          <div className="flex justify-between items-start gap-3">
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="text-xs font-medium text-gray-500 dark:text-gray-400">
                                  #{index + 1}
                                </span>
                                <div className={`w-2 h-2 rounded-full ${colorClass}`} />
                                <p className="font-semibold text-sm text-gray-900 dark:text-white truncate">
                                  {areaNameMatches && searchQuery ? (
                                    <HighlightMatch text={area.name} query={searchQuery} />
                                  ) : (
                                    area.name
                                  )}
                                </p>
                                {hasMatchingRoutes && searchQuery && (
                                  <span className="text-xs px-1.5 py-0.5 bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 rounded">
                                    {areaRoutes.length} route{areaRoutes.length !== 1 ? 's' : ''}
                                  </span>
                                )}
                              </div>
                              <div className="grid grid-cols-2 gap-x-3 gap-y-1 text-xs text-gray-600 dark:text-gray-400 mt-2">
                                <div className="flex justify-between">
                                  <span>Score:</span>
                                  <span className="font-semibold text-gray-900 dark:text-white">{area.activity_score}</span>
                                </div>
                                <div className="flex justify-between">
                                  <span>Ticks:</span>
                                  <span className="font-semibold text-gray-900 dark:text-white">{area.total_ticks}</span>
                                </div>
                                <div className="flex justify-between">
                                  <span>Routes:</span>
                                  <span className="font-semibold text-gray-900 dark:text-white">{area.active_routes}</span>
                                </div>
                                <div className="flex justify-between">
                                  <span>Climbers:</span>
                                  <span className="font-semibold text-gray-900 dark:text-white">{area.unique_climbers}</span>
                                </div>
                              </div>
                              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                Last activity: {formatDistanceToNow(new Date(area.last_activity), { addSuffix: true })}
                              </p>
                            </div>
                          </div>
                        </button>
                      </div>

                      {/* Routes List - show when expanded and has matching routes */}
                      {hasMatchingRoutes && isExpanded && searchQuery && (
                        <div className="border-t border-gray-200 dark:border-gray-700 px-3 pb-2">
                          <div className="space-y-1 mt-2">
                            {areaRoutes.map((route) => (
                              <div
                                key={route.mp_route_id}
                                className="pl-8 pr-2 py-2 bg-white dark:bg-gray-800 rounded text-xs hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
                              >
                                <div className="flex items-start justify-between gap-2">
                                  <div className="flex-1 min-w-0">
                                    <p className="font-medium text-gray-900 dark:text-white truncate">
                                      <HighlightMatch text={route.name} query={searchQuery} />
                                    </p>
                                    <div className="flex items-center gap-2 mt-1 text-gray-600 dark:text-gray-400">
                                      <span className="font-mono">{route.rating}</span>
                                      <span>•</span>
                                      <span>{route.tick_count} tick{route.tick_count !== 1 ? 's' : ''}</span>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  );
                })
              ) : (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <p>No areas or routes found matching "{searchQuery}"</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number }) {
  return (
    <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-2 sm:p-3">
      <div className="flex items-center justify-center text-gray-500 dark:text-gray-400 mb-1">
        {icon}
      </div>
      <div className="text-xs text-gray-600 dark:text-gray-400 text-center mb-1">{label}</div>
      <div className="text-lg sm:text-xl font-bold text-gray-900 dark:text-white text-center">
        {value.toLocaleString()}
      </div>
    </div>
  );
}

function HighlightMatch({ text, query }: { text: string; query: string }) {
  if (!query) return <>{text}</>;
  
  const parts = text.split(new RegExp(`(${query})`, 'gi'));
  return (
    <>
      {parts.map((part, i) => 
        part.toLowerCase() === query.toLowerCase() ? (
          <mark key={i} className="bg-yellow-200 dark:bg-yellow-700 text-gray-900 dark:text-white">
            {part}
          </mark>
        ) : (
          <span key={i}>{part}</span>
        )
      )}
    </>
  );
}
