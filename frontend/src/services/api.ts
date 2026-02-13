import axios from 'axios';
import { Location, WeatherForecast, AllWeatherResponse, AreaActivitySummary, RouteActivitySummary, ClimbHistoryEntry, SearchResult, BoulderDryingStatus, AreaDryingStats } from '../types/weather';
import { Area, AreaWithLocations } from '../types/area';
import { HeatMapActivityResponse, AreaActivityDetail, RoutesResponse, GeoBounds } from '../types/heatmap';

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
});

export const weatherApi = {
  // Get all locations
  getLocations: async (): Promise<{ locations: Location[] }> => {
    const response = await api.get('/locations');
    return response.data;
  },

  // Get weather for specific location
  getWeatherForLocation: async (locationId: number): Promise<WeatherForecast> => {
    const response = await api.get(`/weather/${locationId}`);
    return response.data;
  },

  // Get weather for all locations (optionally filtered by area)
  getAllWeather: async (areaId?: number | null): Promise<AllWeatherResponse> => {
    const params = areaId ? { area_id: areaId } : {};
    // Use longer timeout for getAllWeather since it fetches multiple locations
    const response = await api.get('/weather/all', { params, timeout: 30000 });
    return response.data;
  },

  // Get weather by coordinates (for custom locations)
  getWeatherByCoordinates: async (lat: number, lon: number): Promise<{ current: any; forecast: any[] }> => {
    const response = await api.get('/weather/coordinates', {
      params: { lat, lon },
    });
    return response.data;
  },

  // Health check
  healthCheck: async (): Promise<{ status: string; message: string; time: string }> => {
    const response = await api.get('/health');
    return response.data;
  },

  // Get all areas with location counts
  getAreas: async (): Promise<{ areas: Area[] }> => {
    const response = await api.get('/areas');
    return response.data;
  },

  // Get locations for a specific area
  getLocationsByArea: async (areaId: number): Promise<AreaWithLocations> => {
    const response = await api.get(`/areas/${areaId}/locations`);
    return response.data;
  },
};

export const climbActivityApi = {
  // Get areas ordered by recent activity for a location
  getAreasOrderedByActivity: async (locationId: number): Promise<AreaActivitySummary[]> => {
    const response = await api.get(`/climbs/location/${locationId}/areas`);
    return response.data;
  },

  // Get subareas of a parent area ordered by recent activity
  getSubareasOrderedByActivity: async (locationId: number, areaId: number): Promise<AreaActivitySummary[]> => {
    const response = await api.get(`/climbs/location/${locationId}/areas/${areaId}/subareas`);
    return response.data;
  },

  // Get routes in an area ordered by recent activity
  getRoutesOrderedByActivity: async (locationId: number, areaId: number, limit = 200): Promise<RouteActivitySummary[]> => {
    const response = await api.get(`/climbs/location/${locationId}/areas/${areaId}/routes`, {
      params: { limit }
    });
    return response.data;
  },

  // Get recent ticks for a specific route
  getRecentTicksForRoute: async (routeId: number, limit = 5): Promise<ClimbHistoryEntry[]> => {
    const response = await api.get(`/climbs/routes/${routeId}/ticks`, {
      params: { limit }
    });
    return response.data;
  },

  // Search all areas and routes in a location by name
  searchInLocation: async (locationId: number, searchQuery: string, limit = 50): Promise<SearchResult[]> => {
    const response = await api.get(`/climbs/location/${locationId}/search-all`, {
      params: { q: searchQuery, limit }
    });
    return response.data;
  },

  // Search all routes in a location by name, grade, or area
  searchRoutesInLocation: async (locationId: number, searchQuery: string, limit = 50): Promise<RouteActivitySummary[]> => {
    const response = await api.get(`/climbs/location/${locationId}/search`, {
      params: { q: searchQuery, limit }
    });
    return response.data;
  },

  // Get boulder-specific drying status for a route
  getBoulderDryingStatus: async (routeId: number): Promise<BoulderDryingStatus> => {
    const response = await api.get(`/climbs/routes/${routeId}/drying-status`);
    return response.data;
  },

  // Get boulder-specific drying status for multiple routes in batch
  getBatchBoulderDryingStatus: async (routeIds: number[]): Promise<Record<number, BoulderDryingStatus>> => {
    if (routeIds.length === 0) {
      return {};
    }
    const response = await api.get('/climbs/routes/batch-drying-status', {
      params: { route_ids: routeIds.join(',') },
      timeout: 30000, // Longer timeout for batch request
    });
    return response.data;
  },

  // Get area-level drying statistics
  getAreaDryingStats: async (locationId: number, areaId: number): Promise<AreaDryingStats> => {
    const response = await api.get(`/climbs/location/${locationId}/areas/${areaId}/drying-stats`);
    return response.data;
  },

  // Get area-level drying statistics for multiple areas in batch
  getBatchAreaDryingStats: async (locationId: number, areaIds: number[]): Promise<Record<number, AreaDryingStats>> => {
    if (areaIds.length === 0) {
      return {};
    }
    const response = await api.get(`/climbs/location/${locationId}/batch-area-drying-stats`, {
      params: { area_ids: areaIds.join(',') },
      timeout: 30000, // Longer timeout for batch request
    });
    return response.data;
  },
};

export const heatMapApi = {
  // Get heat map activity data for visualization
  getHeatMapActivity: async (params: {
    startDate: Date;
    endDate: Date;
    bounds?: GeoBounds;
    minActivity?: number;
    limit?: number;
  }): Promise<HeatMapActivityResponse> => {
    const queryParams: Record<string, string | number> = {
      start_date: params.startDate.toISOString().split('T')[0],
      end_date: params.endDate.toISOString().split('T')[0],
    };

    if (params.bounds) {
      queryParams.min_lat = params.bounds.minLat;
      queryParams.max_lat = params.bounds.maxLat;
      queryParams.min_lon = params.bounds.minLon;
      queryParams.max_lon = params.bounds.maxLon;
    }

    if (params.minActivity) queryParams.min_activity = params.minActivity;
    if (params.limit) queryParams.limit = params.limit;

    const response = await api.get('/heat-map/activity', {
      params: queryParams,
      timeout: 30000, // 30s timeout for potentially large datasets
    });
    return response.data;
  },

  // Get detailed activity information for a specific area
  getAreaDetail: async (
    areaId: number,
    dateRange: { start: Date; end: Date }
  ): Promise<AreaActivityDetail> => {
    const response = await api.get(`/heat-map/area/${areaId}/detail`, {
      params: {
        start_date: dateRange.start.toISOString().split('T')[0],
        end_date: dateRange.end.toISOString().split('T')[0],
      },
      timeout: 20000, // 20s timeout
    });
    return response.data;
  },

  // Get routes within geographic bounds with activity
  getRoutesByBounds: async (params: {
    bounds: GeoBounds;
    startDate: Date;
    endDate: Date;
    limit?: number;
  }): Promise<RoutesResponse> => {
    const response = await api.get('/heat-map/routes', {
      params: {
        min_lat: params.bounds.minLat,
        max_lat: params.bounds.maxLat,
        min_lon: params.bounds.minLon,
        max_lon: params.bounds.maxLon,
        start_date: params.startDate.toISOString().split('T')[0],
        end_date: params.endDate.toISOString().split('T')[0],
        limit: params.limit || 100,
      },
      timeout: 20000, // 20s timeout
    });
    return response.data;
  },
};

export default api;
