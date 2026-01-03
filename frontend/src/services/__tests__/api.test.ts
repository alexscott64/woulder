import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';
import type { AllWeatherResponse, WeatherForecast } from '../../types/weather';

// Mock axios module
vi.mock('axios', () => {
  const mockAxiosInstance = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  };

  return {
    default: {
      create: vi.fn(() => mockAxiosInstance),
    },
  };
});

describe('weatherApi', () => {
  let mockGet: any;

  beforeEach(() => {
    // Clear all mocks before each test
    vi.clearAllMocks();

    // Get the mock axios instance
    const axiosInstance = axios.create();
    mockGet = axiosInstance.get as any;
  });

  describe('getAllWeather', () => {
    it('should return response with forecasts array (not weather)', async () => {
      // Import weatherApi after mocks are set up
      const { weatherApi } = await import('../api');

      // Mock API response matching the new backend format
      const mockResponse: AllWeatherResponse = {
        forecasts: [
          {
            location_id: 1,
            location: {
              id: 1,
              name: 'Test Location',
              latitude: 47.5,
              longitude: -121.5,
              elevation_ft: 1000,
              area_id: 1,
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            current: {
              timestamp: '2024-01-01T12:00:00Z',
              temperature: 50,
              feels_like: 48,
              precipitation: 0,
              humidity: 60,
              wind_speed: 5,
              wind_direction: 180,
              cloud_cover: 20,
              pressure: 1013,
              description: 'Clear',
              icon: '01d',
            },
            hourly: [],
            historical: [],
          },
        ],
        count: 1,
        updated_at: '2024-01-01T12:00:00Z',
      };

      mockGet.mockResolvedValue({ data: mockResponse });

      const result = await weatherApi.getAllWeather();

      expect(result).toHaveProperty('forecasts');
      expect(result).toHaveProperty('count');
      expect(result).toHaveProperty('updated_at');
      expect(result.forecasts).toBeInstanceOf(Array);
      expect(result.forecasts).toHaveLength(1);
    });

    it('should NOT have a weather property (old format)', async () => {
      const { weatherApi } = await import('../api');

      const mockResponse: AllWeatherResponse = {
        forecasts: [],
        count: 0,
        updated_at: '2024-01-01T12:00:00Z',
      };

      mockGet.mockResolvedValue({ data: mockResponse });

      const result = await weatherApi.getAllWeather();

      // Ensure the old 'weather' property doesn't exist
      expect(result).not.toHaveProperty('weather');
      expect(result).toHaveProperty('forecasts');
    });

    it('should pass area_id as query parameter when provided', async () => {
      const { weatherApi } = await import('../api');

      mockGet.mockResolvedValue({
        data: { forecasts: [], count: 0, updated_at: '2024-01-01T12:00:00Z' },
      });

      await weatherApi.getAllWeather(5);

      expect(mockGet).toHaveBeenCalledWith('/weather/all', {
        params: { area_id: 5 },
        timeout: 30000,
      });
    });

    it('should not pass area_id when null', async () => {
      const { weatherApi } = await import('../api');

      mockGet.mockResolvedValue({
        data: { forecasts: [], count: 0, updated_at: '2024-01-01T12:00:00Z' },
      });

      await weatherApi.getAllWeather(null);

      expect(mockGet).toHaveBeenCalledWith('/weather/all', {
        params: {},
        timeout: 30000,
      });
    });
  });

  describe('getLocations', () => {
    it('should return response with locations array', async () => {
      const { weatherApi } = await import('../api');

      const mockResponse = {
        locations: [
          {
            id: 1,
            name: 'Test Location',
            latitude: 47.5,
            longitude: -121.5,
            elevation_ft: 1000,
            area_id: 1,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        ],
        count: 1,
      };

      mockGet.mockResolvedValue({ data: mockResponse });

      const result = await weatherApi.getLocations();

      expect(result).toHaveProperty('locations');
      expect(result.locations).toBeInstanceOf(Array);
    });
  });

  describe('getAreas', () => {
    it('should return response with areas array', async () => {
      const { weatherApi } = await import('../api');

      const mockResponse = {
        areas: [
          {
            id: 1,
            name: 'Test Area',
            location_count: 5,
          },
        ],
        count: 1,
      };

      mockGet.mockResolvedValue({ data: mockResponse });

      const result = await weatherApi.getAreas();

      expect(result).toHaveProperty('areas');
      expect(result.areas).toBeInstanceOf(Array);
    });
  });

  describe('getWeatherForLocation', () => {
    it('should return a single WeatherForecast object', async () => {
      const { weatherApi } = await import('../api');

      const mockForecast: WeatherForecast = {
        location_id: 1,
        location: {
          id: 1,
          name: 'Test Location',
          latitude: 47.5,
          longitude: -121.5,
          elevation_ft: 1000,
          area_id: 1,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        current: {
          timestamp: '2024-01-01T12:00:00Z',
          temperature: 50,
          feels_like: 48,
          precipitation: 0,
          humidity: 60,
          wind_speed: 5,
          wind_direction: 180,
          cloud_cover: 20,
          pressure: 1013,
          description: 'Clear',
          icon: '01d',
        },
        hourly: [],
        historical: [],
      };

      mockGet.mockResolvedValue({ data: mockForecast });

      const result = await weatherApi.getWeatherForLocation(1);

      expect(result).toHaveProperty('location_id');
      expect(result).toHaveProperty('location');
      expect(result).toHaveProperty('current');
      expect(result).toHaveProperty('hourly');
      expect(result).toHaveProperty('historical');
    });
  });
});
