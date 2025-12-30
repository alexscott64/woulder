import { describe, it, expect } from 'vitest';
import { PestAnalyzer } from '../PestAnalyzer';
import { WeatherData } from '../../../../types/weather';

// Helper to create weather data
function createWeatherData(overrides: Partial<WeatherData> = {}): WeatherData {
  return {
    timestamp: new Date().toISOString(),
    temperature: 75,
    feels_like: 75,
    precipitation: 0,
    humidity: 50,
    wind_speed: 5,
    wind_direction: 180,
    cloud_cover: 0,
    pressure: 1013,
    description: 'Clear',
    icon: '01d',
    ...overrides
  };
}

describe('PestAnalyzer', () => {
  describe('calculateMosquitoScore', () => {
    it('should return low score for cold temperatures', () => {
      const result = PestAnalyzer.calculateMosquitoScore(45, 70, 5, 1, 7);
      expect(result.score).toBeLessThan(10);
      expect(result.factors.some(f => f.includes('Too cold'))).toBe(true);
    });

    it('should return high score for optimal conditions', () => {
      const result = PestAnalyzer.calculateMosquitoScore(75, 70, 5, 2, 7);
      expect(result.score).toBeGreaterThan(70);
      expect(result.factors.some(f => f.includes('Optimal mosquito temperature'))).toBe(true);
    });

    it('should penalize high wind', () => {
      const calmWind = PestAnalyzer.calculateMosquitoScore(75, 70, 3, 1, 7);
      const strongWind = PestAnalyzer.calculateMosquitoScore(75, 70, 20, 1, 7);
      expect(calmWind.score).toBeGreaterThan(strongWind.score);
    });

    it('should increase score with recent rainfall', () => {
      const noRain = PestAnalyzer.calculateMosquitoScore(75, 70, 5, 0, 7);
      const withRain = PestAnalyzer.calculateMosquitoScore(75, 70, 5, 2, 7);
      expect(withRain.score).toBeGreaterThan(noRain.score);
    });

    it('should consider season', () => {
      const winter = PestAnalyzer.calculateMosquitoScore(75, 70, 5, 1, 1);
      const summer = PestAnalyzer.calculateMosquitoScore(75, 70, 5, 1, 7);
      expect(summer.score).toBeGreaterThan(winter.score);
    });
  });

  describe('calculateOutdoorPestScore', () => {
    it('should return low score for cold temperatures', () => {
      const result = PestAnalyzer.calculateOutdoorPestScore(45, 50, 0.5, 7);
      expect(result.score).toBeLessThan(15);
    });

    it('should return high score for warm temperatures', () => {
      const result = PestAnalyzer.calculateOutdoorPestScore(80, 70, 1, 7);
      expect(result.score).toBeGreaterThan(70);
    });

    it('should consider humidity', () => {
      const lowHumidity = PestAnalyzer.calculateOutdoorPestScore(75, 30, 1, 7);
      const highHumidity = PestAnalyzer.calculateOutdoorPestScore(75, 70, 1, 7);
      expect(highHumidity.score).toBeGreaterThan(lowHumidity.score);
    });

    it('should increase score with recent rainfall', () => {
      const noRain = PestAnalyzer.calculateOutdoorPestScore(75, 70, 0, 7);
      const withRain = PestAnalyzer.calculateOutdoorPestScore(75, 70, 1, 7);
      expect(withRain.score).toBeGreaterThan(noRain.score);
    });
  });

  describe('assessConditions', () => {
    it('should return low levels for winter cold weather', () => {
      const current = createWeatherData({ temperature: 35, humidity: 40 });
      const historical: WeatherData[] = [];

      const result = PestAnalyzer.assessConditions(current, historical);

      expect(result.mosquitoLevel).toBe('low');
      expect(result.outdoorPestLevel).toBe('low');
    });

    it('should return high levels for summer warm weather', () => {
      const current = createWeatherData({ temperature: 80, humidity: 75 });
      const historical = Array(20).fill(null).map(() =>
        createWeatherData({
          temperature: 80,
          precipitation: 0.1,
          timestamp: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString()
        })
      );

      const result = PestAnalyzer.assessConditions(current, historical);

      expect(['high', 'very_high', 'extreme']).toContain(result.mosquitoLevel);
      expect(['high', 'very_high', 'extreme']).toContain(result.outdoorPestLevel);
    });

    it('should return factors array', () => {
      const current = createWeatherData();
      const historical: WeatherData[] = [];

      const result = PestAnalyzer.assessConditions(current, historical);

      expect(result.factors).toBeInstanceOf(Array);
      expect(result.factors.length).toBeGreaterThan(0);
      expect(result.factors.length).toBeLessThanOrEqual(4);
    });

    it('should return scores between 0 and 100', () => {
      const current = createWeatherData();
      const historical: WeatherData[] = [];

      const result = PestAnalyzer.assessConditions(current, historical);

      expect(result.mosquitoScore).toBeGreaterThanOrEqual(0);
      expect(result.mosquitoScore).toBeLessThanOrEqual(100);
      expect(result.outdoorPestScore).toBeGreaterThanOrEqual(0);
      expect(result.outdoorPestScore).toBeLessThanOrEqual(100);
    });
  });

  describe('assessDayConditions', () => {
    it('should return low for empty day hours', () => {
      const result = PestAnalyzer.assessDayConditions([], new Date(), 0);

      expect(result.mosquitoLevel).toBe('low');
      expect(result.outdoorPestLevel).toBe('low');
      expect(result.worstLevel).toBe('low');
    });

    it('should use peak temperatures', () => {
      const dayHours = [
        createWeatherData({ temperature: 60 }),
        createWeatherData({ temperature: 80 }), // Peak
        createWeatherData({ temperature: 70 }),
      ];

      const result = PestAnalyzer.assessDayConditions(dayHours, new Date(2024, 6, 15), 1);

      // With 80°F peak temp, should have higher pest activity
      expect(['moderate', 'high', 'very_high', 'extreme']).toContain(result.outdoorPestLevel);
    });

    it('should consider season in assessment', () => {
      const summerHours = [createWeatherData({ temperature: 75 })];
      const winterHours = [createWeatherData({ temperature: 75 })];

      const summerResult = PestAnalyzer.assessDayConditions(summerHours, new Date(2024, 6, 15), 0);
      const winterResult = PestAnalyzer.assessDayConditions(winterHours, new Date(2024, 0, 15), 0);

      // Summer should have higher activity (compare levels not scores)
      const levels = ['low', 'moderate', 'high', 'very_high', 'extreme'];
      const summerIndex = levels.indexOf(summerResult.mosquitoLevel);
      const winterIndex = levels.indexOf(winterResult.mosquitoLevel);
      expect(summerIndex).toBeGreaterThanOrEqual(winterIndex);
    });

    it('should return worst level between mosquito and pest levels', () => {
      const dayHours = [createWeatherData({ temperature: 85, humidity: 80 })];

      const result = PestAnalyzer.assessDayConditions(dayHours, new Date(2024, 6, 15), 2);

      // worstLevel should be at least as high as both individual levels
      const levels = ['low', 'moderate', 'high', 'very_high', 'extreme'];
      const mosquitoIndex = levels.indexOf(result.mosquitoLevel);
      const pestIndex = levels.indexOf(result.outdoorPestLevel);
      const worstIndex = levels.indexOf(result.worstLevel);

      expect(worstIndex).toBeGreaterThanOrEqual(mosquitoIndex);
      expect(worstIndex).toBeGreaterThanOrEqual(pestIndex);
    });
  });
});
