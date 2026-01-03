import { describe, it, expect } from 'vitest';
import { ConditionCalculator } from '../ConditionCalculator';
import { WeatherData } from '../../../../types/weather';

// Helper to create weather data
function createWeather(overrides: Partial<WeatherData> = {}): WeatherData {
  return {
    timestamp: '2026-01-07T12:00:00Z',
    temperature: 50,
    feels_like: 50,
    precipitation: 0,
    humidity: 60,
    wind_speed: 5,
    wind_direction: 180,
    cloud_cover: 50,
    pressure: 1013,
    description: 'Clear',
    icon: '01d',
    ...overrides,
  };
}

describe('RealWorldConditions', () => {
  describe('Skykomish Paradise Tuesday scenario', () => {
    it('should rate persistent 0.05+ in/hr rain with freezing temps as bad', () => {
      const weather = createWeather({
        temperature: 32, // Freezing
        precipitation: 0.05, // Every hour
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
      expect(result.reasons.some(r => r.includes('rain') || r.includes('Rain'))).toBe(true);
    });

    it('should rate 0.05 in/hr as moderate rain (bad)', () => {
      const weather = createWeather({
        precipitation: 0.05,
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
      expect(result.reasons.some(r => r.toLowerCase().includes('rain'))).toBe(true);
    });

    it('should rate freezing temps (32°F) as marginal at minimum', () => {
      const weather = createWeather({
        temperature: 32,
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).not.toBe('good');
      expect(result.reasons.some(r => r.toLowerCase().includes('cold') || r.toLowerCase().includes('freez'))).toBe(true);
    });

    it('should rate freezing + moderate rain as bad (multiple factors)', () => {
      const weather = createWeather({
        temperature: 32,
        precipitation: 0.06,
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
      expect(result.reasons.length).toBeGreaterThan(1);
    });
  });

  describe('Precipitation thresholds', () => {
    it('should rate 0.02 in/hr as marginal (light drizzle)', () => {
      const weather = createWeather({ precipitation: 0.02 });
      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('marginal');
    });

    it('should rate 0.05 in/hr as bad (moderate rain)', () => {
      const weather = createWeather({ precipitation: 0.05 });
      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
    });

    it('should rate 0.10 in/hr as bad (heavy rain)', () => {
      const weather = createWeather({ precipitation: 0.10 });
      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
    });

    it('should rate 0.30+ in/hr as bad (very heavy rain)', () => {
      const weather = createWeather({ precipitation: 0.35 });
      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
    });
  });

  describe('Temperature extremes', () => {
    it('should rate 28°F (too cold) as bad', () => {
      const weather = createWeather({ temperature: 28 });
      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
    });

    it('should rate 35°F (cold) as marginal', () => {
      const weather = createWeather({ temperature: 35 });
      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('marginal');
    });

    it('should rate 80°F (hot) as marginal', () => {
      const weather = createWeather({ temperature: 80 });
      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('marginal');
    });
  });

  describe('Multiple marginal factors should downgrade', () => {
    it('should rate cold + windy + humid as bad (3 marginal factors)', () => {
      const weather = createWeather({
        temperature: 38, // Cold (marginal)
        wind_speed: 18,  // Windy (marginal)
        humidity: 90,    // High humidity (marginal)
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
      expect(result.reasons.length).toBeGreaterThanOrEqual(3);
    });

    it('should rate cold + light rain as bad (2 marginal factors)', () => {
      const weather = createWeather({
        temperature: 35,      // Cold (marginal)
        precipitation: 0.02,  // Light rain (marginal)
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
    });
  });

  describe('Good conditions', () => {
    it('should rate perfect conditions as good', () => {
      const weather = createWeather({
        temperature: 55,
        precipitation: 0,
        wind_speed: 5,
        humidity: 50,
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('good');
    });

    it('should rate slightly warm with low wind as good', () => {
      const weather = createWeather({
        temperature: 68,
        wind_speed: 8,
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('good');
    });
  });
});
