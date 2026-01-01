import axios from 'axios';
import { Location, WeatherForecast, AllWeatherResponse } from '../types/weather';
import { Area, AreaWithLocations } from '../types/area';

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

export default api;
