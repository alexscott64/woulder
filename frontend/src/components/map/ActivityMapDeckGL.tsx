import { useState, useMemo, useEffect } from 'react';
import DeckGL from '@deck.gl/react';
import { Map } from 'react-map-gl/maplibre';
import { ScatterplotLayer } from '@deck.gl/layers';
import { HexagonLayer } from '@deck.gl/aggregation-layers';
import { HeatMapPoint } from '../../types/heatmap';
import { ChevronDown, ChevronUp } from 'lucide-react';
import 'maplibre-gl/dist/maplibre-gl.css';

interface ActivityMapDeckGLProps {
  points: HeatMapPoint[];
  onAreaClick: (areaId: number) => void;
  selectedAreaId?: number | null;
  onShowCluster?: (areas: HeatMapPoint[]) => void;
}

// Color-code by recency
function getColorByRecency(lastActivity: string): [number, number, number, number] {
  const daysSince = (Date.now() - new Date(lastActivity).getTime()) / (1000 * 60 * 60 * 24);
  
  if (daysSince <= 7) return [239, 68, 68, 200];      // Red - Hot
  if (daysSince <= 30) return [249, 115, 22, 200];    // Orange - Warm
  if (daysSince <= 90) return [234, 179, 8, 200];     // Yellow - Recent
  return [59, 130, 246, 200];                         // Blue - Older
}

// Get size based on activity score
function getRadiusByActivity(activityScore: number, maxScore: number): number {
  const normalized = Math.log10(activityScore + 1) / Math.log10(maxScore + 1);
  return 50 + normalized * 400; // 50-450 meters
}

type ViewMode = 'scatter' | 'hexagon';

const MAP_VIEW_STORAGE_KEY = 'activityMapViewMode';

// Load map view mode from localStorage
function loadMapViewMode(): ViewMode {
  try {
    const stored = localStorage.getItem(MAP_VIEW_STORAGE_KEY);
    if (stored === 'scatter' || stored === 'hexagon') {
      return stored;
    }
  } catch (error) {
    console.error('Failed to load map view mode from localStorage:', error);
  }
  return 'hexagon'; // default
}

// Save map view mode to localStorage
function saveMapViewMode(mode: ViewMode) {
  try {
    localStorage.setItem(MAP_VIEW_STORAGE_KEY, mode);
  } catch (error) {
    console.error('Failed to save map view mode to localStorage:', error);
  }
}

export function ActivityMapDeckGL({ points, onAreaClick, selectedAreaId, onShowCluster }: ActivityMapDeckGLProps) {
  const [viewState, setViewState] = useState({
    longitude: -100.0,
    latitude: 45.0,
    zoom: 3.5,
    pitch: 0,
    bearing: 0,
  });
  
  const [viewMode, setViewMode] = useState<ViewMode>(loadMapViewMode());
  
  // Legend collapsed by default on mobile
  const [legendExpanded, setLegendExpanded] = useState(false);
  
  useEffect(() => {
    // Set initial legend state based on screen size
    const isMobile = window.innerWidth < 768;
    setLegendExpanded(!isMobile);
  }, []);

  // Save view mode to localStorage when it changes
  useEffect(() => {
    saveMapViewMode(viewMode);
  }, [viewMode]);
  const [hoveredObject, setHoveredObject] = useState<any>(null);
  const [hoverInfo, setHoverInfo] = useState<{ x: number; y: number } | null>(null);
  const [isReady, setIsReady] = useState(false);

  // Suppress known Firefox WebGL errors (cosmetic only, doesn't affect functionality)
  useEffect(() => {
    const handleError = (event: ErrorEvent) => {
      const errorMsg = event.message || '';
      if (errorMsg.includes('device.limits') || errorMsg.includes('maxTextureDimension2D')) {
        event.preventDefault();
        event.stopPropagation();
        return false;
      }
    };

    window.addEventListener('error', handleError);
    return () => window.removeEventListener('error', handleError);
  }, []);

  // Initialize map when document is visible
  useEffect(() => {
    if (document.visibilityState === 'visible' && points.length > 0) {
      setIsReady(true);
    }

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && points.length > 0) {
        setIsReady(true);
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, [points.length]);

  const maxScore = useMemo(() => Math.max(...points.map(p => p.activity_score)), [points]);

  // Calculate the actual date range from the data
  const dateRangeInfo = useMemo(() => {
    if (points.length === 0) return { oldestDays: 0, newestDays: 0, totalDays: 0 };
    
    const now = Date.now();
    const dates = points.map(p => new Date(p.last_activity).getTime());
    const oldest = Math.min(...dates);
    const newest = Math.max(...dates);
    
    return {
      oldestDays: Math.floor((now - oldest) / (1000 * 60 * 60 * 24)),
      newestDays: Math.floor((now - newest) / (1000 * 60 * 60 * 24)),
      totalDays: Math.floor((newest - oldest) / (1000 * 60 * 60 * 24)),
    };
  }, [points]);

  // Prepare data with colors and sizes
  const preparedData = useMemo(() => {
    return points.map(point => ({
      ...point,
      position: [point.longitude, point.latitude, 0],
      color: getColorByRecency(point.last_activity),
      radius: getRadiusByActivity(point.activity_score, maxScore),
    }));
  }, [points, maxScore]);

  // Individual scatter plot layer
  const scatterLayer = new ScatterplotLayer({
    id: 'scatter-layer',
    data: preparedData,
    pickable: true,
    opacity: 0.8,
    stroked: true,
    filled: true,
    radiusScale: 1,
    radiusMinPixels: 3,
    radiusMaxPixels: 20,
    lineWidthMinPixels: 1,
    getPosition: (d: any) => d.position,
    getRadius: (d: any) => d.radius,
    getFillColor: (d: any) => d.color,
    getLineColor: (d: any) => {
      // Highlight selected
      if (selectedAreaId === d.mp_area_id) {
        return [30, 64, 175, 255];
      }
      return [255, 255, 255, 100];
    },
    getLineWidth: (d: any) => selectedAreaId === d.mp_area_id ? 3 : 1,
    onClick: (info: any) => {
      if (info.object) {
        onAreaClick(info.object.mp_area_id);
      }
      return true;
    },
    onHover: (info: any) => {
      if (info.object) {
        setHoveredObject(info.object);
        setHoverInfo({ x: info.x, y: info.y });
      } else {
        setHoveredObject(null);
        setHoverInfo(null);
      }
      return true;
    },
  });

  // Hexagon layer for clustering view
  // We manually find nearby points on click since the layer doesn't expose them
  const cellSizeM = 50000; // 50km cells
  const hexagonLayer = new HexagonLayer({
    id: 'hexagon-layer',
    data: preparedData,
    pickable: true,
    extruded: false,
    radius: cellSizeM,
    elevationScale: 0,
    getPosition: (d: any) => d.position,
    getColorWeight: (d: any) => d.activity_score,
    colorAggregation: 'SUM',
    colorRange: [
      [59, 130, 246, 180],      // Blue - Low
      [234, 179, 8, 180],       // Yellow - Medium
      [249, 115, 22, 180],      // Orange - High
      [239, 68, 68, 180],       // Red - Very High
    ],
    onClick: (info: any) => {
      if (info.coordinate) {
        const [clickLon, clickLat] = info.coordinate;
        
        // Find all points within ~70km of click (larger than cell size to catch all in hexagon)
        const searchRadiusKm = 70;
        const nearbyPoints = preparedData.filter(point => {
          const [lon, lat] = point.position;
          // Rough distance calculation (good enough for filtering)
          const dLat = Math.abs(lat - clickLat);
          const dLon = Math.abs(lon - clickLon);
          const distanceKm = Math.sqrt((dLat * 111) ** 2 + (dLon * 111 * Math.cos(clickLat * Math.PI / 180)) ** 2);
          return distanceKm <= searchRadiusKm;
        });
        
        console.log('Found', nearbyPoints.length, 'nearby points');
        
        if (nearbyPoints.length >= 1 && onShowCluster) {
          // Always show cluster drawer (even for single areas)
          onShowCluster(nearbyPoints);
        }
      }
      return true;
    },
    onHover: (info: any) => {
      if (info.coordinate) {
        const [clickLon, clickLat] = info.coordinate;
        
        // Find nearby points for hover preview
        const searchRadiusKm = 70;
        const nearbyPoints = preparedData.filter(point => {
          const [lon, lat] = point.position;
          const dLat = Math.abs(lat - clickLat);
          const dLon = Math.abs(lon - clickLon);
          const distanceKm = Math.sqrt((dLat * 111) ** 2 + (dLon * 111 * Math.cos(clickLat * Math.PI / 180)) ** 2);
          return distanceKm <= searchRadiusKm;
        });
        
        if (nearbyPoints.length > 0) {
          const totalActivity = nearbyPoints.reduce((sum, p) => sum + p.activity_score, 0);
          setHoveredObject({
            isCluster: true,
            count: nearbyPoints.length,
            points: nearbyPoints.slice(0, 10),
            colorValue: totalActivity,
          });
          setHoverInfo({ x: info.x, y: info.y });
        } else {
          setHoveredObject(null);
          setHoverInfo(null);
        }
      } else {
        setHoveredObject(null);
        setHoverInfo(null);
      }
      return true;
    },
  });

  const layers = viewMode === 'scatter' ? [scatterLayer] : [hexagonLayer];

  // Only render deck.gl when component is ready and has data to avoid WebGL errors
  if (!isReady || points.length === 0) {
    return <div className="h-full w-full relative flex items-center justify-center">
      <p className="text-gray-500">{points.length === 0 ? 'No activity data to display' : 'Loading map...'}</p>
    </div>;
  }

  return (
    <div className="h-full w-full relative" style={{ minHeight: '400px' }}>
      <DeckGL
        initialViewState={viewState}
        controller={true}
        layers={layers}
        onViewStateChange={({ viewState }: any) => setViewState(viewState)}
        getTooltip={() => null}
        style={{ position: 'absolute', top: '0', left: '0', right: '0', bottom: '0' }}
      >
        <Map
          mapStyle="https://basemaps.cartocdn.com/gl/positron-gl-style/style.json"
          style={{ width: '100%', height: '100%' }}
        />
      </DeckGL>

      {/* Hover Tooltip */}
      {hoveredObject && hoverInfo && (
        <div
          className="absolute pointer-events-none z-50 bg-white dark:bg-gray-800 rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 p-3 max-w-sm"
          style={{
            left: hoverInfo.x + 10,
            top: hoverInfo.y + 10,
          }}
        >
          {hoveredObject.isCluster ? (
            <div>
              <h3 className="font-bold text-base mb-2">
                {hoveredObject.count} Climbing {hoveredObject.count === 1 ? 'Area' : 'Areas'}
              </h3>
              <div className="space-y-1 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Total Activity:</span>
                  <span className="font-semibold">{Math.round(hoveredObject.colorValue || 0).toLocaleString()}</span>
                </div>
                {hoveredObject.points.length > 0 && (
                  <div className="mt-2 pt-2 border-t border-gray-200 dark:border-gray-700">
                    <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">Sample areas:</div>
                    {hoveredObject.points.slice(0, 5).map((p: HeatMapPoint) => (
                      <div key={p.mp_area_id} className="text-xs text-gray-700 dark:text-gray-300">
                        • {p.name}
                      </div>
                    ))}
                    {hoveredObject.count > 5 && (
                      <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                        ...and {hoveredObject.count - 5} more
                      </div>
                    )}
                  </div>
                )}
                <div className="text-xs text-gray-500 dark:text-gray-400 mt-2 pt-2 border-t border-gray-200 dark:border-gray-700">
                  💡 Click to {hoveredObject.count === 1 ? 'view details' : 'zoom in'}
                </div>
              </div>
            </div>
          ) : (
            <div>
              <h3 className="font-bold text-base mb-2">{hoveredObject.name}</h3>
              <div className="space-y-1 text-sm text-gray-700 dark:text-gray-300">
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Activity Score:</span>
                  <span className="font-semibold">{hoveredObject.activity_score}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Total Ticks:</span>
                  <span className="font-semibold">{hoveredObject.total_ticks}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Active Routes:</span>
                  <span className="font-semibold">{hoveredObject.active_routes}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Climbers:</span>
                  <span className="font-semibold">{hoveredObject.unique_climbers}</span>
                </div>
              </div>
              <div className="text-xs text-gray-500 dark:text-gray-400 mt-2 pt-2 border-t border-gray-200 dark:border-gray-700">
                💡 Click to view details
              </div>
            </div>
          )}
        </div>
      )}

      {/* View Mode Toggle */}
      <div className="absolute top-2 sm:top-4 right-2 sm:right-4 z-10 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
        <div className="flex items-center gap-1 p-1">
          <button
            onClick={() => setViewMode('hexagon')}
            className={`px-2 sm:px-3 py-1.5 sm:py-2 rounded text-xs sm:text-sm font-medium transition-colors ${
              viewMode === 'hexagon'
                ? 'bg-blue-600 text-white'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
            }`}
            title="Hexagon clusters show aggregate activity"
          >
            Clusters
          </button>
          <button
            onClick={() => setViewMode('scatter')}
            className={`px-2 sm:px-3 py-1.5 sm:py-2 rounded text-xs sm:text-sm font-medium transition-colors ${
              viewMode === 'scatter'
                ? 'bg-blue-600 text-white'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
            }`}
            title="Individual points"
          >
            Points
          </button>
        </div>
      </div>

      {/* Legend */}
      <div className={`absolute bottom-2 sm:bottom-6 left-2 sm:left-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 transition-all duration-300 z-10 max-w-[calc(100vw-1rem)] sm:max-w-sm`}>
        <button
          onClick={() => setLegendExpanded(!legendExpanded)}
          className="flex items-center justify-between w-full px-2 sm:px-3 py-1.5 sm:py-2 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors"
        >
          <div className="flex items-center gap-1.5 sm:gap-2">
            <div className="flex gap-0.5 sm:gap-1">
              <div className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-red-500" />
              <div className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-orange-500" />
              <div className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-yellow-500" />
              <div className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-blue-500" />
            </div>
            <span className="text-xs font-medium text-gray-900 dark:text-white">Legend</span>
          </div>
          {legendExpanded ? (
            <ChevronDown className="w-3 h-3 sm:w-4 sm:h-4 text-gray-500" />
          ) : (
            <ChevronUp className="w-3 h-3 sm:w-4 sm:h-4 text-gray-500" />
          )}
        </button>
        
        {legendExpanded && (
          <div className="px-2 sm:px-3 pb-2 sm:pb-3 pt-1 border-t border-gray-200 dark:border-gray-700">
            {viewMode === 'scatter' ? (
              // Individual points mode - show recency colors
              <div className="space-y-1 sm:space-y-1.5 mt-1.5 sm:mt-2">
                <div className="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1 sm:mb-2">Point Color = Recency</div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-red-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">
                    {dateRangeInfo.oldestDays <= 7 ? 'Most Recent' : 'Last Week'}
                  </span>
                </div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-orange-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">
                    {dateRangeInfo.oldestDays <= 30 ? '7+ Days Ago' : 'Last Month'}
                  </span>
                </div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-yellow-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">
                    {dateRangeInfo.oldestDays <= 90 ? '30+ Days Ago' : 'Last 3 Months'}
                  </span>
                </div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-blue-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">
                    {dateRangeInfo.oldestDays <= 90 ? '90+ Days Ago' : 'Older'}
                  </span>
                </div>
              </div>
            ) : (
              // Cluster mode - show activity level gradient
              <div className="space-y-1 sm:space-y-1.5 mt-1.5 sm:mt-2">
                <div className="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1 sm:mb-2">Cluster Color = Activity Level</div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-red-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">Very High Activity</span>
                </div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-orange-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">High Activity</span>
                </div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-yellow-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">Medium Activity</span>
                </div>
                <div className="flex items-center gap-1.5 sm:gap-2">
                  <div className="w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-blue-500 flex-shrink-0" />
                  <span className="text-xs text-gray-700 dark:text-gray-300">Low Activity</span>
                </div>
                <div className="mt-2 sm:mt-3 p-1.5 sm:p-2 bg-blue-50 dark:bg-blue-900/20 rounded text-xs text-gray-600 dark:text-gray-400">
                  💡 <span className="font-medium">Click cluster</span> to see all areas
                </div>
              </div>
            )}
            <div className="mt-1.5 sm:mt-2 pt-1.5 sm:pt-2 border-t border-gray-200 dark:border-gray-600">
              <p className="text-xs text-gray-600 dark:text-gray-400 leading-relaxed">
                <span className="font-medium">Size:</span> {viewMode === 'scatter' ? 'Activity Score' : 'Area Count'}
              </p>
              <p className="text-xs text-gray-600 dark:text-gray-400 mt-0.5 sm:mt-1 leading-relaxed">
                <span className="font-medium">Mode:</span> {viewMode === 'hexagon' ? 'Clustered' : 'Individual'}
              </p>
              {dateRangeInfo.oldestDays > 0 && (
                <p className="text-xs text-gray-500 dark:text-gray-500 mt-0.5 sm:mt-1 italic leading-relaxed">
                  Data: {dateRangeInfo.newestDays === 0 ? 'today' : `${dateRangeInfo.newestDays}d ago`} to {dateRangeInfo.oldestDays}d ago
                </p>
              )}
            </div>
          </div>
        )}
      </div>

    </div>
  );
}
