import React, { useEffect, useRef } from 'react';
import { MapContainer, TileLayer, CircleMarker, Popup, useMap } from 'react-leaflet';
import { HeatMapPoint } from '../../types/heatmap';
import 'leaflet/dist/leaflet.css';

interface ActivityMapProps {
  points: HeatMapPoint[];
  onAreaClick: (areaId: number) => void;
  selectedAreaId?: number | null;
}

// Color-code markers by recency
function getMarkerColor(lastActivity: string): string {
  const daysSince = (Date.now() - new Date(lastActivity).getTime()) / (1000 * 60 * 60 * 24);
  
  if (daysSince <= 7) return '#ef4444';      // Red - Hot (last week)
  if (daysSince <= 30) return '#f97316';     // Orange - Warm (last month)
  if (daysSince <= 90) return '#eab308';     // Yellow - Recent (last 3 months)
  return '#3b82f6';                          // Blue - Older
}

// Size marker by activity score
function getMarkerRadius(activityScore: number, allScores: number[]): number {
  const maxScore = Math.max(...allScores);
  const minScore = Math.min(...allScores);
  const range = maxScore - minScore;
  
  // Scale between 8 and 24 pixels
  const normalized = range > 0 ? (activityScore - minScore) / range : 0.5;
  return 8 + normalized * 16;
}

// Component to fit map bounds to points
function FitBounds({ points }: { points: HeatMapPoint[] }) {
  const map = useMap();
  
  useEffect(() => {
    if (points.length > 0) {
      const bounds = points.map(p => [p.latitude, p.longitude] as [number, number]);
      map.fitBounds(bounds, { padding: [50, 50] });
    }
  }, [points, map]);
  
  return null;
}

export function ActivityMap({ points, onAreaClick, selectedAreaId }: ActivityMapProps) {
  const mapRef = useRef(null);
  
  // North America center as default
  const center: [number, number] = [45.0, -100.0];
  const zoom = 4;
  
  const allScores = points.map(p => p.activity_score);

  return (
    <div className="h-full w-full relative">
      <MapContainer
        center={center}
        zoom={zoom}
        style={{ height: '100%', width: '100%' }}
        ref={mapRef}
        className="z-0"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        
        {points.length > 0 && <FitBounds points={points} />}
        
        {points.map((point) => {
          const color = getMarkerColor(point.last_activity);
          const radius = getMarkerRadius(point.activity_score, allScores);
          const isSelected = selectedAreaId === point.mp_area_id;
          
          return (
            <CircleMarker
              key={point.mp_area_id}
              center={[point.latitude, point.longitude]}
              radius={radius}
              pathOptions={{
                fillColor: color,
                fillOpacity: isSelected ? 0.9 : 0.7,
                color: isSelected ? '#1e40af' : '#fff',
                weight: isSelected ? 3 : 2,
              }}
              eventHandlers={{
                click: () => onAreaClick(point.mp_area_id),
              }}
            >
              <Popup>
                <div className="text-sm min-w-[200px]">
                  <h3 className="font-bold text-base mb-2">{point.name}</h3>
                  <div className="space-y-1 text-gray-700">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Activity Score:</span>
                      <span className="font-semibold">{point.activity_score}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Total Ticks:</span>
                      <span className="font-semibold">{point.total_ticks}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Active Routes:</span>
                      <span className="font-semibold">{point.active_routes}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Climbers:</span>
                      <span className="font-semibold">{point.unique_climbers}</span>
                    </div>
                  </div>
                  <button
                    onClick={() => onAreaClick(point.mp_area_id)}
                    className="mt-3 w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded transition-colors"
                  >
                    View Details
                  </button>
                </div>
              </Popup>
            </CircleMarker>
          );
        })}
      </MapContainer>
      
      {/* Legend */}
      <div className="absolute bottom-6 left-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 z-[1000] border border-gray-200 dark:border-gray-700">
        <h4 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">Activity Recency</h4>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded-full bg-red-500" />
            <span className="text-xs text-gray-700 dark:text-gray-300">Last Week</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded-full bg-orange-500" />
            <span className="text-xs text-gray-700 dark:text-gray-300">Last Month</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded-full bg-yellow-500" />
            <span className="text-xs text-gray-700 dark:text-gray-300">Last 3 Months</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded-full bg-blue-500" />
            <span className="text-xs text-gray-700 dark:text-gray-300">Older</span>
          </div>
        </div>
        <div className="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600">
          <p className="text-xs text-gray-600 dark:text-gray-400">
            <span className="font-medium">Size</span> = Activity Score
          </p>
        </div>
      </div>
    </div>
  );
}
