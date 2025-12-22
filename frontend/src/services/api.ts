import axios from 'axios';
import { Location, WeatherForecast, AllWeatherResponse } from '../types/weather';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

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

  // Get weather for all locations
  getAllWeather: async (): Promise<AllWeatherResponse> => {
    const response = await api.get('/weather/all');
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
};

export default api;
