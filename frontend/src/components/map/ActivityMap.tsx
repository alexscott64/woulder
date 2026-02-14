import React, { useEffect, useRef, useState } from 'react';
import { MapContainer, TileLayer, CircleMarker, Popup, useMap } from 'react-leaflet';
import MarkerClusterGroup from 'react-leaflet-markercluster';
import L from 'leaflet';
import { HeatMapPoint } from '../../types/heatmap';
import { ChevronDown, ChevronUp } from 'lucide-react';
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
  console.log('===== ActivityMap component rendering with', points.length, 'points =====');
  
  const mapRef = useRef(null);
  const clusterRef = useRef<any>(null);
  const [legendExpanded, setLegendExpanded] = useState(false);
  
  // North America center as default
  const center: [number, number] = [45.0, -100.0];
  const zoom = 4;
  
  const allScores = points.map(p => p.activity_score);
  
  // Create a lookup map for faster point lookups in iconCreateFunction
  const pointsByName = useRef<Map<string, HeatMapPoint>>(new Map());
  useEffect(() => {
    pointsByName.current = new Map(points.map(p => [p.name, p]));
  }, [points]);

  // Callback when cluster group is created
  const handleClusterGroupCreated = (clusterGroup: any) => {
    console.log('Cluster group created!', clusterGroup);
    clusterRef.current = clusterGroup;
    
    const handleClusterClick = (e: any) => {
      console.log('Cluster click event fired');
      // Prevent default zoom behavior
      if (e.originalEvent) {
        e.originalEvent.stopPropagation();
      }
      L.DomEvent.stopPropagation(e);
      
      const cluster = e.layer;
      const childCount = cluster.getChildCount();
      
      console.log(`Cluster has ${childCount} items`);
      
      // Get all child markers and extract their data
      const childMarkers = cluster.getAllChildMarkers();
      
      // Calculate aggregate statistics
      let totalTicks = 0;
      let totalActiveRoutes = 0;
      let totalUniqueClimbers = 0;
      let totalActivityScore = 0;
      const areasList: string[] = [];
      
      childMarkers.forEach((marker: any) => {
        const pathOptions: any = marker.options?.pathOptions || {};
        const point = pointsByName.current.get(pathOptions.areaName);
        if (point) {
          totalTicks += point.total_ticks;
          totalActiveRoutes += point.active_routes;
          totalUniqueClimbers += point.unique_climbers;
          totalActivityScore += point.activity_score;
          areasList.push(point.name);
        }
      });
      
      console.log(`Calculated stats: ${totalTicks} ticks, ${totalActiveRoutes} routes`);
      
      // Show aggregated data popup
      let popupContent = `<div class="p-3 min-w-[280px]" style="font-family: system-ui, -apple-system, sans-serif;">
        <h3 style="font-weight: bold; font-size: 1.125rem; margin-bottom: 0.75rem; color: #111827; border-bottom: 1px solid #d1d5db; padding-bottom: 0.5rem;">${childCount} Climbing ${childCount === 1 ? 'Area' : 'Areas'}</h3>
        
        <div style="display: flex; flex-direction: column; gap: 0.5rem; margin-bottom: 0.75rem;">
          <div style="display: flex; justify-content: space-between; align-items: center; background-color: #eff6ff; padding: 0.5rem; border-radius: 0.375rem;">
            <span style="font-size: 0.875rem; font-weight: 500; color: #374151;">Total Activity Score:</span>
            <span style="font-size: 1rem; font-weight: bold; color: #2563eb;">${totalActivityScore.toLocaleString()}</span>
          </div>
          <div style="display: flex; justify-between; align-items: center; background-color: #f9fafb; padding: 0.5rem; border-radius: 0.375rem;">
            <span style="font-size: 0.875rem; font-weight: 500; color: #374151;">Total Ticks:</span>
            <span style="font-size: 1rem; font-weight: bold; color: #111827;">${totalTicks.toLocaleString()}</span>
          </div>
          <div style="display: flex; justify-between; align-items: center; background-color: #f9fafb; padding: 0.5rem; border-radius: 0.375rem;">
            <span style="font-size: 0.875rem; font-weight: 500; color: #374151;">Active Routes:</span>
            <span style="font-size: 1rem; font-weight: bold; color: #111827;">${totalActiveRoutes.toLocaleString()}</span>
          </div>
          <div style="display: flex; justify-between; align-items: center; background-color: #f9fafb; padding: 0.5rem; border-radius: 0.375rem;">
            <span style="font-size: 0.875rem; font-weight: 500; color: #374151;">Unique Climbers:</span>
            <span style="font-size: 1rem; font-weight: bold; color: #111827;">${totalUniqueClimbers.toLocaleString()}</span>
          </div>
        </div>`;
      
      // Show list of areas (up to 30)
      if (childCount <= 30) {
        popupContent += `<div style="border-top: 1px solid #d1d5db; padding-top: 0.5rem; margin-top: 0.5rem;">
          <div style="font-size: 0.75rem; font-weight: 600; color: #4b5563; margin-bottom: 0.5rem;">AREAS IN THIS CLUSTER:</div>
          <div style="display: flex; flex-direction: column; gap: 0.25rem; max-height: 12rem; overflow-y: auto;">`;
        
        areasList.slice(0, 30).forEach((areaName) => {
          popupContent += `<div style="font-size: 0.75rem; color: #374151; padding: 0.25rem 0.5rem; background-color: #f9fafb; border-radius: 0.25rem;">• ${areaName}</div>`;
        });
        
        popupContent += `</div></div>`;
      }
      
      popupContent += `<div style="font-size: 0.75rem; color: #6b7280; margin-top: 0.75rem; padding-top: 0.5rem; border-top: 1px solid #e5e7eb;">
        💡 Tip: Zoom in further to see individual climbing areas
      </div></div>`;
      
      cluster.bindPopup(popupContent, {
        maxWidth: 350,
        className: 'cluster-aggregate-popup'
      }).openPopup();
      
      console.log('Popup should now be visible');
    };
    
    console.log('Attaching clusterclick handler');
    clusterGroup.on('clusterclick', handleClusterClick);
  };

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
          ref={clusterRef}
          whenCreated={handleClusterGroupCreated}
          chunkedLoading
          maxClusterRadius={80}
          spiderfyOnMaxZoom={true}
          showCoverageOnHover={false}
          zoomToBoundsOnClick={false}
          disableClusteringAtZoom={15}
          removeOutsideVisibleBounds={true}
          animate={true}
          animateAddingMarkers={false}
          spiderfyDistanceMultiplier={2}
          iconCreateFunction={(cluster: any) => {
            const count = cluster.getChildCount();
            
            // Simplified approach: determine color based on count tiers
            // Higher count clusters likely have more recent activity
            let bgColor = 'bg-blue-500';
            
            // Only get child markers for smaller clusters to avoid performance issues
            if (count <= 50) {
              const childMarkers = cluster.getAllChildMarkers();
              let mostRecentTimestamp = 0;
              
              // Use cached lookup instead of array.find()
              for (let i = 0; i < childMarkers.length && i < 20; i++) {
                const marker = childMarkers[i];
                const pathOptions: any = marker.options?.pathOptions || {};
                const point = pointsByName.current.get(pathOptions.areaName);
                if (point) {
                  const timestamp = new Date(point.last_activity).getTime();
                  if (timestamp > mostRecentTimestamp) {
                    mostRecentTimestamp = timestamp;
                  }
                }
              }
              
              if (mostRecentTimestamp > 0) {
                const daysSince = (Date.now() - mostRecentTimestamp) / (1000 * 60 * 60 * 24);
                if (daysSince <= 7) bgColor = 'bg-red-500';
                else if (daysSince <= 30) bgColor = 'bg-orange-500';
                else if (daysSince <= 90) bgColor = 'bg-yellow-500';
              }
            } else {
              // For large clusters, use a heuristic: very large clusters in popular areas
              // are more likely to have recent activity
              if (count > 200) bgColor = 'bg-orange-500';
              else if (count > 100) bgColor = 'bg-yellow-500';
            }
            
            // Size based on count
            let sizeClass = 'w-10 h-10 text-xs';
            if (count > 100) {
              sizeClass = 'w-16 h-16 text-lg';
            } else if (count > 20) {
              sizeClass = 'w-12 h-12 text-sm';
            }
            
            return L.divIcon({
              html: `<div class="flex items-center justify-center ${sizeClass} ${bgColor} bg-opacity-80 rounded-full text-white font-bold border-2 border-white shadow-lg">${count}</div>`,
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
                  // @ts-ignore - Store area name in options for cluster popup
                  areaName: point.name,
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
      
      {/* Legend - Collapsible and Compact */}
      <div className={`absolute bottom-6 left-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 transition-all duration-300 ${selectedAreaId ? 'z-[30]' : 'z-[1000]'}`}>
        {/* Legend Header - Always Visible */}
        <button
          onClick={() => setLegendExpanded(!legendExpanded)}
          className="flex items-center justify-between w-full px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors"
        >
          <div className="flex items-center gap-2">
            <div className="flex gap-1">
              <div className="w-2 h-2 rounded-full bg-red-500" />
              <div className="w-2 h-2 rounded-full bg-orange-500" />
              <div className="w-2 h-2 rounded-full bg-yellow-500" />
              <div className="w-2 h-2 rounded-full bg-blue-500" />
            </div>
            <span className="text-xs font-medium text-gray-900 dark:text-white">Legend</span>
          </div>
          {legendExpanded ? (
            <ChevronDown className="w-4 h-4 text-gray-500" />
          ) : (
            <ChevronUp className="w-4 h-4 text-gray-500" />
          )}
        </button>
        
        {/* Expanded Legend Details */}
        {legendExpanded && (
          <div className="px-3 pb-3 pt-1 border-t border-gray-200 dark:border-gray-700">
            <div className="space-y-1.5 mt-2">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500" />
                <span className="text-xs text-gray-700 dark:text-gray-300">Last Week</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-orange-500" />
                <span className="text-xs text-gray-700 dark:text-gray-300">Last Month</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-yellow-500" />
                <span className="text-xs text-gray-700 dark:text-gray-300">Last 3 Months</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-blue-500" />
                <span className="text-xs text-gray-700 dark:text-gray-300">Older</span>
              </div>
            </div>
            <div className="mt-2 pt-2 border-t border-gray-200 dark:border-gray-600">
              <p className="text-xs text-gray-600 dark:text-gray-400">
                <span className="font-medium">Size & Opacity:</span> Activity Level
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
