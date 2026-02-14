import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { X, TrendingUp, Users, MapPin, Loader2, ExternalLink, Calendar, ChevronRight, ChevronLeft, MessageSquare, Activity, Info } from 'lucide-react';
import { heatMapApi } from '../../services/api';
import { formatDistanceToNow, format } from 'date-fns';

interface AreaDetailDrawerProps {
  areaId: number | null;
  dateRange: { start: Date; end: Date };
  isOpen: boolean;
  onClose: () => void;
  onBack?: () => void; // Optional back callback for cluster navigation
}

type View = 'area' | 'route';
type Tab = 'routes' | 'ticks' | 'comments';

export function AreaDetailDrawer({ areaId, dateRange, isOpen, onClose, onBack }: AreaDetailDrawerProps) {
  const [view, setView] = useState<View>('area');
  const [selectedRouteId, setSelectedRouteId] = useState<number | null>(null);
  const [selectedTab, setSelectedTab] = useState<Tab>('routes');

  const { data: detail, isLoading } = useQuery({
    queryKey: ['areaDetail', areaId, dateRange],
    queryFn: () => heatMapApi.getAreaDetail(areaId!, dateRange),
    enabled: !!areaId && isOpen,
    staleTime: 10 * 60 * 1000,
  });

  // Fetch route-specific ticks when a route is selected
  const { data: routeTicksData, isLoading: isLoadingRouteTicks } = useQuery({
    queryKey: ['routeTicks', selectedRouteId, dateRange],
    queryFn: () => heatMapApi.getRouteTicksInDateRange({
      routeId: selectedRouteId!,
      startDate: dateRange.start,
      endDate: dateRange.end,
      limit: 500, // Fetch up to 500 ticks for the route
    }),
    enabled: !!selectedRouteId && view === 'route',
    staleTime: 10 * 60 * 1000,
  });

  const handleRouteClick = (routeId: number) => {
    setSelectedRouteId(routeId);
    setView('route');
  };

  const handleBackToArea = () => {
    setView('area');
    setSelectedRouteId(null);
    setSelectedTab('routes');
  };

  if (!isOpen) return null;

  const selectedRoute = detail?.top_routes?.find(r => r.mp_route_id === selectedRouteId);
  
  // Use route-specific ticks when available, otherwise fall back to filtered recent ticks
  const routeTicks = routeTicksData?.ticks || detail?.recent_ticks?.filter(t => t.mp_route_id === selectedRouteId) || [];
  const routeComments = detail?.recent_comments?.filter(c => c.mp_route_id === selectedRouteId) || [];

  // Format date range for display
  const dateRangeText = `${format(dateRange.start, 'MMM d, yyyy')} - ${format(dateRange.end, 'MMM d, yyyy')}`;
  const dayCount = Math.ceil((dateRange.end.getTime() - dateRange.start.getTime()) / (1000 * 60 * 60 * 24));

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
            {(view === 'route' || (view === 'area' && onBack)) && (
              <button
                onClick={view === 'route' ? handleBackToArea : onBack}
                className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors shrink-0"
                title={view === 'route' ? 'Back to area' : 'Back to cluster'}
              >
                <ChevronLeft className="w-5 h-5" />
              </button>
            )}
            <h2 className="text-lg sm:text-xl font-bold text-gray-900 dark:text-white truncate flex-1">
              {view === 'area'
                ? (detail?.name || 'Loading...')
                : (selectedRoute?.name || 'Route Details')}
            </h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors shrink-0"
              title="Close"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
          
          {/* Date Range Indicator */}
          <div className="mt-2 flex items-center gap-2 px-2 py-1.5 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded text-xs">
            <Calendar className="w-3.5 h-3.5 text-blue-600 dark:text-blue-400 shrink-0" />
            <span className="text-blue-700 dark:text-blue-300 font-medium">
              {dateRangeText}
            </span>
            <span className="text-blue-600 dark:text-blue-400">
              ({dayCount} days)
            </span>
          </div>

          {view === 'route' && selectedRoute && (
            <div className="mt-2 flex items-center gap-3 text-sm text-gray-600 dark:text-gray-400">
              <span className="font-semibold text-blue-600 dark:text-blue-400">{selectedRoute.rating}</span>
              <span>·</span>
              <span>{selectedRoute.tick_count} {selectedRoute.tick_count === 1 ? 'tick' : 'ticks'} in period</span>
            </div>
          )}
        </div>

        {/* Content */}
        {isLoading ? (
          <div className="p-8 text-center">
            <Loader2 className="w-12 h-12 animate-spin text-blue-600 dark:text-blue-400 mx-auto mb-4" />
            <p className="text-gray-600 dark:text-gray-400">Loading details...</p>
          </div>
        ) : detail ? (
          <>
            {view === 'area' ? (
              <div className="p-3 sm:p-4 space-y-4 sm:space-y-6">
                {/* Stats Cards */}
                <div className="grid grid-cols-3 gap-2 sm:gap-3">
                  <StatCard
                    icon={<TrendingUp className="w-4 h-4 sm:w-5 sm:h-5" />}
                    label="Total Ticks"
                    value={detail.total_ticks}
                    subtitle="in period"
                  />
                  <StatCard
                    icon={<MapPin className="w-4 h-4 sm:w-5 sm:h-5" />}
                    label="Routes"
                    value={detail.active_routes}
                    subtitle="with ticks"
                  />
                  <StatCard
                    icon={<Users className="w-4 h-4 sm:w-5 sm:h-5" />}
                    label="Climbers"
                    value={detail.unique_climbers}
                    subtitle="unique"
                  />
                </div>

                {/* Last Activity */}
                <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-3">
                  <div className="flex items-center gap-2 text-blue-700 dark:text-blue-400 text-sm mb-1">
                    <Calendar className="w-4 h-4" />
                    <span className="font-medium">Most Recent Activity</span>
                  </div>
                  <p className="text-gray-900 dark:text-white font-semibold">
                    {formatDistanceToNow(new Date(detail.last_activity), { addSuffix: true })}
                  </p>
                </div>

                {/* Tabs */}
                <div className="border-b border-gray-200 dark:border-gray-700">
                  <div className="flex -mb-px">
                    <TabButton
                      active={selectedTab === 'routes'}
                      onClick={() => setSelectedTab('routes')}
                      icon={<MapPin className="w-4 h-4" />}
                      label="Routes"
                      count={detail.top_routes?.length || 0}
                    />
                    <TabButton
                      active={selectedTab === 'ticks'}
                      onClick={() => setSelectedTab('ticks')}
                      icon={<Activity className="w-4 h-4" />}
                      label="Ticks"
                      count={detail.recent_ticks?.length || 0}
                    />
                    <TabButton
                      active={selectedTab === 'comments'}
                      onClick={() => setSelectedTab('comments')}
                      icon={<MessageSquare className="w-4 h-4" />}
                      label="Comments"
                      count={detail.recent_comments?.length || 0}
                    />
                  </div>
                </div>

                {/* Tab Content */}
                <div className="space-y-2">
                  {selectedTab === 'routes' && (
                    <>
                      <div className="flex items-start gap-2 p-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded text-xs">
                        <Info className="w-4 h-4 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" />
                        <div className="text-blue-700 dark:text-blue-300">
                          <p className="font-medium mb-0.5">Top {detail.top_routes?.length || 0} most active routes</p>
                          <p>Sorted by number of ticks in the selected time period. Click to see tick details.</p>
                        </div>
                      </div>
                      {detail.top_routes && detail.top_routes.length > 0 && (
                        <>
                          {detail.top_routes.slice(0, 20).map((route) => (
                            <button
                              key={route.mp_route_id}
                              onClick={() => handleRouteClick(route.mp_route_id)}
                              className="w-full bg-gray-50 dark:bg-gray-900 rounded-lg p-3 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-left group"
                            >
                              <div className="flex justify-between items-start">
                                <div className="flex-1 min-w-0 pr-3">
                                  <p className="font-semibold text-sm text-gray-900 dark:text-white truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
                                    {route.name}
                                  </p>
                                  <p className="text-xs text-gray-600 dark:text-gray-400 mt-0.5">
                                    {route.rating}
                                  </p>
                                </div>
                                <div className="flex items-center gap-2 shrink-0">
                                  <div className="text-right">
                                    <p className="text-sm font-bold text-blue-600 dark:text-blue-400">
                                      {route.tick_count}
                                    </p>
                                    <p className="text-xs text-gray-500 dark:text-gray-400">
                                      {route.tick_count === 1 ? 'tick' : 'ticks'}
                                    </p>
                                  </div>
                                  <ChevronRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors" />
                                </div>
                              </div>
                            </button>
                          ))}
                        </>
                      )}
                    </>
                  )}

                  {selectedTab === 'ticks' && (
                    <>
                      <div className="flex items-start gap-2 p-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded text-xs">
                        <Info className="w-4 h-4 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" />
                        <div className="text-blue-700 dark:text-blue-300">
                          <p className="font-medium mb-0.5">Showing {detail.recent_ticks?.length || 0} most recent ticks</p>
                          <p>Total: {detail.total_ticks} ticks in the selected time period</p>
                        </div>
                      </div>
                      {detail.recent_ticks && detail.recent_ticks.length > 0 && (
                        <>
                          {detail.recent_ticks.slice(0, 20).map((tick, index) => (
                            <div
                              key={`${tick.mp_route_id}-${index}`}
                              className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3"
                            >
                              <div className="flex justify-between items-start mb-1">
                                <div className="flex-1 min-w-0">
                                  <p className="font-semibold text-sm text-gray-900 dark:text-white truncate">
                                    {tick.route_name}
                                  </p>
                                  <p className="text-xs text-gray-600 dark:text-gray-400 mt-0.5">
                                    {tick.rating} · <span className="font-medium">{tick.user_name || 'Anonymous'}</span>
                                    {tick.style && <span> · {tick.style}</span>}
                                  </p>
                                </div>
                                <span className="text-xs text-gray-500 dark:text-gray-400 ml-2 shrink-0">
                                  {formatDistanceToNow(new Date(tick.climbed_at), { addSuffix: true })}
                                </span>
                              </div>
                              {tick.comment && (
                                <p className="text-xs text-gray-600 dark:text-gray-400 mt-2 italic line-clamp-3">
                                  "{tick.comment}"
                                </p>
                              )}
                            </div>
                          ))}
                        </>
                      )}
                    </>
                  )}

                  {selectedTab === 'comments' && (
                    <>
                      <div className="flex items-start gap-2 p-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded text-xs">
                        <Info className="w-4 h-4 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" />
                        <div className="text-blue-700 dark:text-blue-300">
                          <p className="font-medium mb-0.5">Showing {detail.recent_comments?.length || 0} most recent comments</p>
                          <p>Comments on routes in this area during the selected time period</p>
                        </div>
                      </div>
                      {detail.recent_comments && detail.recent_comments.length > 0 && (
                        <>
                          {detail.recent_comments.slice(0, 20).map((comment) => (
                            <div
                              key={comment.id}
                              className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3"
                            >
                              <div className="flex justify-between items-start mb-2">
                                <span className="text-sm font-semibold text-gray-900 dark:text-white">
                                  {comment.user_name}
                                </span>
                                <span className="text-xs text-gray-500 dark:text-gray-400 ml-2 shrink-0">
                                  {formatDistanceToNow(new Date(comment.commented_at), { addSuffix: true })}
                                </span>
                              </div>
                              <p className="text-sm text-gray-700 dark:text-gray-300 mb-1">
                                {comment.comment_text}
                              </p>
                              {comment.route_name && (
                                <p className="text-xs text-blue-600 dark:text-blue-400 font-medium">
                                  on {comment.route_name}
                                </p>
                              )}
                            </div>
                          ))}
                        </>
                      )}
                    </>
                  )}
                </div>

                {/* External Link */}
                <a
                  href={`https://www.mountainproject.com/area/${areaId}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center gap-2 w-full py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition-colors"
                >
                  <span>View on Mountain Project</span>
                  <ExternalLink className="w-4 h-4" />
                </a>
              </div>
            ) : (
              <div className="p-3 sm:p-4 space-y-4 sm:space-y-6">
                {/* Route Stats */}
                {selectedRoute && (
                  <div className="bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">Ticks in Period</div>
                        <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{selectedRoute.tick_count}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">({dateRangeText})</div>
                      </div>
                      <div>
                        <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">Last Climbed</div>
                        <div className="text-sm font-semibold text-gray-900 dark:text-white">
                          {formatDistanceToNow(new Date(selectedRoute.last_activity), { addSuffix: true })}
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {/* Route Ticks */}
                <div>
                  <h3 className="text-base font-semibold text-gray-900 dark:text-white mb-2 flex items-center gap-2">
                    <Activity className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                    Ticks {isLoadingRouteTicks && <Loader2 className="w-4 h-4 animate-spin" />}
                  </h3>
                  {routeTicksData ? (
                    <div className="flex items-start gap-2 p-2 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded text-xs mb-3">
                      <Info className="w-4 h-4 text-green-600 dark:text-green-400 shrink-0 mt-0.5" />
                      <div className="text-green-700 dark:text-green-300">
                        <p className="font-medium mb-0.5">
                          Showing all {routeTicks.length} ticks in the selected time period
                        </p>
                        {routeTicks.length >= 500 && <p className="text-xs mt-0.5">(Limited to 500 most recent. View on MP for complete history.)</p>}
                      </div>
                    </div>
                  ) : (
                    <div className="flex items-start gap-2 p-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded text-xs mb-3">
                      <Info className="w-4 h-4 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" />
                      <div className="text-blue-700 dark:text-blue-300">
                        <p className="font-medium">Loading all ticks for this route...</p>
                      </div>
                    </div>
                  )}
                  <div className="space-y-2">
                    {isLoadingRouteTicks ? (
                      <div className="text-center py-8">
                        <Loader2 className="w-8 h-8 animate-spin text-blue-600 dark:text-blue-400 mx-auto mb-2" />
                        <p className="text-sm text-gray-500 dark:text-gray-400">Loading ticks...</p>
                      </div>
                    ) : routeTicks.length > 0 ? (
                      routeTicks.map((tick, index) => (
                        <div
                          key={`tick-${index}`}
                          className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3 border border-gray-200 dark:border-gray-700"
                        >
                          <div className="flex justify-between items-start mb-2">
                            <div className="flex-1">
                              <span className="font-semibold text-sm text-gray-900 dark:text-white">
                                {tick.user_name || 'Anonymous'}
                              </span>
                              {tick.style && (
                                <span className="text-xs text-gray-600 dark:text-gray-400 ml-2">
                                  · {tick.style}
                                </span>
                              )}
                            </div>
                            <span className="text-xs text-gray-500 dark:text-gray-400 ml-2 shrink-0">
                              {formatDistanceToNow(new Date(tick.climbed_at), { addSuffix: true })}
                            </span>
                          </div>
                          {tick.comment && (
                            <p className="text-sm text-gray-700 dark:text-gray-300 italic">
                              "{tick.comment}"
                            </p>
                          )}
                        </div>
                      ))
                    ) : (
                      <p className="text-sm text-gray-500 dark:text-gray-400 text-center py-4 bg-gray-50 dark:bg-gray-900 rounded-lg">
                        No tick details available in the loaded data.
                        <br />
                        <span className="text-xs">View on Mountain Project for complete tick history.</span>
                      </p>
                    )}
                  </div>
                </div>

                {/* Route Comments */}
                {routeComments.length > 0 && (
                  <div>
                    <h3 className="text-base font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                      <MessageSquare className="w-5 h-5 text-purple-600 dark:text-purple-400" />
                      Comments ({routeComments.length})
                    </h3>
                    <div className="space-y-2">
                      {routeComments.map((comment) => (
                        <div
                          key={comment.id}
                          className="bg-gray-50 dark:bg-gray-900 rounded-lg p-3 border border-gray-200 dark:border-gray-700"
                        >
                          <div className="flex justify-between items-start mb-2">
                            <span className="font-semibold text-sm text-gray-900 dark:text-white">
                              {comment.user_name}
                            </span>
                            <span className="text-xs text-gray-500 dark:text-gray-400 ml-2 shrink-0">
                              {formatDistanceToNow(new Date(comment.commented_at), { addSuffix: true })}
                            </span>
                          </div>
                          <p className="text-sm text-gray-700 dark:text-gray-300">
                            {comment.comment_text}
                          </p>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* External Link for Route */}
                {selectedRouteId && (
                  <a
                    href={`https://www.mountainproject.com/route/${selectedRouteId}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center justify-center gap-2 w-full py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition-colors"
                  >
                    <span>View Complete History on Mountain Project</span>
                    <ExternalLink className="w-4 h-4" />
                  </a>
                )}
              </div>
            )}
          </>
        ) : null}
      </div>
    </>
  );
}

function StatCard({ icon, label, value, subtitle }: { icon: React.ReactNode; label: string; value: number; subtitle?: string }) {
  return (
    <div className="bg-gray-50 dark:bg-gray-900 rounded-lg p-2 sm:p-3">
      <div className="flex items-center justify-center text-gray-500 dark:text-gray-400 mb-1">
        {icon}
      </div>
      <div className="text-xs text-gray-600 dark:text-gray-400 text-center mb-1">{label}</div>
      <div className="text-lg sm:text-xl font-bold text-gray-900 dark:text-white text-center">
        {value.toLocaleString()}
      </div>
      {subtitle && (
        <div className="text-xs text-gray-500 dark:text-gray-400 text-center mt-0.5">
          {subtitle}
        </div>
      )}
    </div>
  );
}

function TabButton({ 
  active, 
  onClick, 
  icon, 
  label, 
  count 
}: { 
  active: boolean; 
  onClick: () => void; 
  icon: React.ReactNode; 
  label: string; 
  count: number;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-1.5 px-2 sm:px-4 py-2 border-b-2 text-xs sm:text-sm font-medium transition-colors ${
        active
          ? 'border-blue-600 text-blue-600 dark:text-blue-400'
          : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
      }`}
    >
      {icon}
      <span>{label}</span>
      <span className={`text-xs px-1.5 py-0.5 rounded-full ${
        active
          ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300'
          : 'bg-gray-100 dark:bg-gray-700'
      }`}>
        {count}
      </span>
    </button>
  );
}
