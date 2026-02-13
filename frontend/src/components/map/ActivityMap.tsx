import React, { useEffect, useRef } from 'react';
import { MapContainer, TileLayer, CircleMarker, Popup, useMap } from 'react-leaflet';
import MarkerClusterGroup from 'react-leaflet-markercluster';
import L from 'leaflet';
import { HeatMapPoint } from '../../types/heatmap';
import 'leaflet/dist/leaflet.css';
import 'leaflet.markercluster/dist/MarkerCluster.css';
import 'leaflet.markercluster/dist/MarkerCluster.Default.css';

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

// Size marker by activity score using logarithmic scaling
function getMarkerRadius(activityScore: number, allScores: number[]): number {
  const maxScore = Math.max(...allScores);
  const minScore = Math.min(...allScores);
  
  // Use logarithmic scaling to emphasize high-activity areas while still showing low-activity ones
  const logScore = Math.log10(activityScore + 1);
  const logMax = Math.log10(maxScore + 1);
  const logMin = Math.log10(minScore + 1);
  const logRange = logMax - logMin;
  
  // Scale between 3 and 20 pixels (smaller minimum for low-activity areas)
  const normalized = logRange > 0 ? (logScore - logMin) / logRange : 0.5;
  return 3 + normalized * 17;
}

// Get opacity based on activity score - lower activity = much lower opacity
function getMarkerOpacity(activityScore: number, allScores: number[]): number {
  const maxScore = Math.max(...allScores);
  
  // Use logarithmic scaling for opacity
  const logScore = Math.log10(activityScore + 1);
  const logMax = Math.log10(maxScore + 1);
  
  // Scale opacity between 0.15 and 0.8
  const normalized = logMax > 0 ? logScore / logMax : 0.5;
  return 0.15 + normalized * 0.65;
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
        
        <MarkerClusterGroup
          chunkedLoading
          maxClusterRadius={60}
          spiderfyOnMaxZoom={true}
          showCoverageOnHover={false}
          zoomToBoundsOnClick={true}
          removeOutsideVisibleBounds={true}
          animate={true}
          animateAddingMarkers={false}
          iconCreateFunction={(cluster: any) => {
            const count = cluster.getChildCount();
            let size = 'small';
            let sizeClass = 'w-10 h-10 text-xs';
            
            if (count > 100) {
              size = 'large';
              sizeClass = 'w-16 h-16 text-lg';
            } else if (count > 20) {
              size = 'medium';
              sizeClass = 'w-12 h-12 text-sm';
            }
            
            return L.divIcon({
              html: `<div class="flex items-center justify-center ${sizeClass} bg-blue-500 bg-opacity-80 rounded-full text-white font-bold border-2 border-white shadow-lg">${count}</div>`,
              className: 'marker-cluster',
              iconSize: L.point(40, 40, true),
            });
          }}
        >
          {points.map((point) => {
            const color = getMarkerColor(point.last_activity);
            const radius = getMarkerRadius(point.activity_score, allScores);
            const opacity = getMarkerOpacity(point.activity_score, allScores);
            const isSelected = selectedAreaId === point.mp_area_id;
            
            return (
              <CircleMarker
                key={point.mp_area_id}
                center={[point.latitude, point.longitude]}
                radius={radius}
                pathOptions={{
                  fillColor: color,
                  fillOpacity: isSelected ? 0.9 : opacity,
                  color: isSelected ? '#1e40af' : '#fff',
                  weight: isSelected ? 3 : 1,
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
        </MarkerClusterGroup>
      </MapContainer>
      
      {/* Legend - Position adjusted to avoid overlap with drawer */}
      <div className={`absolute bottom-6 left-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-4 border border-gray-200 dark:border-gray-700 transition-all duration-300 ${selectedAreaId ? 'z-[30]' : 'z-[1000]'}`}>
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
            <span className="font-medium">Size & Opacity</span> = Activity Level
          </p>
          <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
            Lower activity areas are less visible
          </p>
        </div>
      </div>
    </div>
  );
}
