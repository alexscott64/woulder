import { describe, it, expect } from 'vitest';
import { ConditionCalculator } from '../ConditionCalculator';
import { WeatherData } from '../../../../types/weather';

// Helper function to create test weather data
function createWeatherData(overrides: Partial<WeatherData> = {}): WeatherData {
  return {
    timestamp: new Date().toISOString(),
    temperature: 60,
    feels_like: 60,
    precipitation: 0,
    humidity: 50,
    wind_speed: 5,
    wind_direction: 180,
    cloud_cover: 20,
    pressure: 1013,
    description: 'clear sky',
    icon: '01d',
    ...overrides
  };
}

describe('ConditionCalculator', () => {
  describe('calculateCondition', () => {
    describe('perfect conditions', () => {
      it('should return good for ideal climbing weather', () => {
        const perfect = createWeatherData({
          temperature: 55,
          precipitation: 0,
          wind_speed: 5,
          humidity: 50
        });

        const result = ConditionCalculator.calculateCondition(perfect);

        expect(result.level).toBe('good');
        expect(result.reasons).toEqual(['Good climbing conditions']);
      });
    });

    describe('precipitation-driven conditions', () => {
      it('should return bad for heavy rain', () => {
        const heavyRain = createWeatherData({ precipitation: 0.35 }); // >0.3 in/hr
        const result = ConditionCalculator.calculateCondition(heavyRain);

        expect(result.level).toBe('bad');
        expect(result.reasons).toContain('Heavy rain (0.35in/hr)');
      });

      it('should return bad for moderate rain', () => {
        const moderateRain = createWeatherData({ precipitation: 0.15 }); // >= 0.05 in/hr
        const result = ConditionCalculator.calculateCondition(moderateRain);

        expect(result.level).toBe('bad');
        expect(result.reasons).toContain('Moderate rain (0.15in/hr)');
      });

      it('should return marginal for persistent drizzle', () => {
        const current = createWeatherData({
          precipitation: 0.02,
          temperature: 48,
          cloud_cover: 95,
          wind_speed: 3
        });
        const recent = [
          createWeatherData({ precipitation: 0.02 }),
          createWeatherData({ precipitation: 0.02 })
        ];

        const result = ConditionCalculator.calculateCondition(current, recent);

        expect(result.level).toBe('marginal');
        expect(result.reasons.some(r => r.includes('Persistent drizzle'))).toBe(true);
      });

      it('should return marginal for light rain even with good drying', () => {
        const current = createWeatherData({
          precipitation: 0.02, // Light rain is always marginal
          temperature: 60, // Ideal temp (not warm)
          cloud_cover: 20,
          wind_speed: 10
        });

        const result = ConditionCalculator.calculateCondition(current);

        expect(result.level).toBe('marginal');
        expect(result.reasons.some(r => r.includes('Light rain'))).toBe(true);
      });
    });

    describe('temperature-driven conditions', () => {
      it('should return bad for too cold', () => {
        const tooCold = createWeatherData({ temperature: 25 });
        const result = ConditionCalculator.calculateCondition(tooCold);

        expect(result.level).toBe('bad');
        expect(result.reasons).toContain('Too cold (25°F)');
      });

      it('should return marginal for cold', () => {
        const cold = createWeatherData({ temperature: 38 });
        const result = ConditionCalculator.calculateCondition(cold);

        expect(result.level).toBe('marginal');
        expect(result.reasons).toContain('Cold (38°F)');
      });

      it('should return marginal for warm', () => {
        const warm = createWeatherData({ temperature: 72 });
        const result = ConditionCalculator.calculateCondition(warm);

        expect(result.level).toBe('marginal');
        expect(result.reasons).toContain('Warm (72°F)');
      });

      it('should return marginal for warm weather', () => {
        const tooHot = createWeatherData({ temperature: 85 }); // Warm (71-85°F)
        const result = ConditionCalculator.calculateCondition(tooHot);

        expect(result.level).toBe('marginal');
        expect(result.reasons).toContain('Warm (85°F)');
      });
    });

    describe('wind-driven conditions', () => {
      it('should return marginal for moderate winds', () => {
        const moderateWind = createWeatherData({ wind_speed: 15 });
        const result = ConditionCalculator.calculateCondition(moderateWind);

        expect(result.level).toBe('marginal');
        expect(result.reasons).toContain('Moderate winds (15mph)');
      });

      it('should return bad for high winds', () => {
        const highWind = createWeatherData({ wind_speed: 25 });
        const result = ConditionCalculator.calculateCondition(highWind);

        expect(result.level).toBe('bad');
        expect(result.reasons).toContain('High winds (25mph)');
      });

      it('should return bad for dangerous winds', () => {
        const dangerousWind = createWeatherData({ wind_speed: 35 });
        const result = ConditionCalculator.calculateCondition(dangerousWind);

        expect(result.level).toBe('bad');
        expect(result.reasons).toContain('Dangerous winds (35mph)');
      });
    });

    describe('humidity-driven conditions', () => {
      it('should return marginal for high humidity', () => {
        const humid = createWeatherData({ humidity: 90 });
        const result = ConditionCalculator.calculateCondition(humid);

        expect(result.level).toBe('marginal');
        expect(result.reasons).toContain('High humidity (90%)');
      });

      it('should not downgrade from good for moderate humidity', () => {
        const normal = createWeatherData({ humidity: 70 });
        const result = ConditionCalculator.calculateCondition(normal);

        expect(result.level).toBe('good');
      });
    });

    describe('multi-factor conditions', () => {
      it('should combine multiple marginal factors into bad', () => {
        const marginal = createWeatherData({
          temperature: 38, // Cold (marginal)
          wind_speed: 15,  // Moderate wind (marginal)
          humidity: 90     // High humidity (marginal)
        });

        const result = ConditionCalculator.calculateCondition(marginal);

        // 2+ marginal factors = bad
        expect(result.level).toBe('bad');
        expect(result.reasons).toHaveLength(3);
        expect(result.reasons).toContain('Cold (38°F)');
        expect(result.reasons).toContain('Moderate winds (15mph)');
        expect(result.reasons).toContain('High humidity (90%)');
      });

      it('should prioritize bad over marginal', () => {
        const bad = createWeatherData({
          temperature: 25, // Too cold (bad)
          wind_speed: 15,  // Moderate wind (marginal)
          humidity: 90     // High humidity (marginal)
        });

        const result = ConditionCalculator.calculateCondition(bad);

        expect(result.level).toBe('bad');
        expect(result.reasons).toContain('Too cold (25°F)');
      });

      it('should stay bad even with one bad factor', () => {
        const bad = createWeatherData({
          precipitation: 0.35, // Heavy rain (bad) - >0.3 in/hr
          temperature: 55,     // Ideal (good)
          wind_speed: 5        // Calm (good)
        });

        const result = ConditionCalculator.calculateCondition(bad);

        expect(result.level).toBe('bad');
        expect(result.reasons).toContain('Heavy rain (0.35in/hr)');
      });
    });
  });

  describe('calculateSnowProbability', () => {
    it('should return high probability for multiple freezing periods with precipitation', () => {
      const snowyData = Array(8).fill(null).map(() =>
        createWeatherData({
          temperature: 28,
          precipitation: 0.05
        })
      );

      const result = ConditionCalculator.calculateSnowProbability(snowyData);

      expect(result.hasSnow).toBe(true);
      expect(result.probability).toBe('High');
      expect(result.accumulationInches).toBeGreaterThan(0);
    });

    it('should calculate snow accumulation correctly (10:1 ratio)', () => {
      const snowyData = [
        createWeatherData({ temperature: 30, precipitation: 0.1 }), // 1 inch snow
        createWeatherData({ temperature: 28, precipitation: 0.05 }), // 0.5 inch snow
        createWeatherData({ temperature: 25, precipitation: 0.1 })   // 1 inch snow
      ];

      const result = ConditionCalculator.calculateSnowProbability(snowyData);

      expect(result.accumulationInches).toBeCloseTo(2.5, 1); // 0.25 * 10 = 2.5 inches
    });

    it('should return moderate probability for some freezing with precipitation', () => {
      const data = [
        createWeatherData({ temperature: 30, precipitation: 0.05 }),
        createWeatherData({ temperature: 35, precipitation: 0 }),
        createWeatherData({ temperature: 40, precipitation: 0 })
      ];

      const result = ConditionCalculator.calculateSnowProbability(data);

      expect(result.hasSnow).toBe(true);
      expect(result.probability).toBe('Moderate');
    });

    it('should return low probability for freezing without precipitation', () => {
      const data = [
        createWeatherData({ temperature: 28, precipitation: 0 }),
        createWeatherData({ temperature: 30, precipitation: 0 }),
        createWeatherData({ temperature: 35, precipitation: 0 })
      ];

      const result = ConditionCalculator.calculateSnowProbability(data);

      expect(result.hasSnow).toBe(false);
      expect(result.probability).toBe('Low');
    });

    it('should return none probability for warm weather', () => {
      const data = Array(8).fill(null).map(() =>
        createWeatherData({
          temperature: 55,
          precipitation: 0.05
        })
      );

      const result = ConditionCalculator.calculateSnowProbability(data);

      expect(result.hasSnow).toBe(false);
      expect(result.probability).toBe('None');
      expect(result.accumulationInches).toBe(0);
    });

    it('should not count precipitation above freezing as snow', () => {
      const data = [
        createWeatherData({ temperature: 35, precipitation: 0.2 }), // Rain, not snow
        createWeatherData({ temperature: 28, precipitation: 0.05 })  // Snow
      ];

      const result = ConditionCalculator.calculateSnowProbability(data);

      expect(result.accumulationInches).toBeCloseTo(0.5, 1); // Only 0.05 * 10
    });
  });

  describe('calculate48HourRain', () => {
    it('should calculate total precipitation over 48 hours', () => {
      const now = new Date();
      const data: WeatherData[] = [
        createWeatherData({ precipitation: 0.1, timestamp: new Date(now.getTime() - 1 * 3600000).toISOString() }),
        createWeatherData({ precipitation: 0.2, timestamp: new Date(now.getTime() - 6 * 3600000).toISOString() }),
        createWeatherData({ precipitation: 0.15, timestamp: new Date(now.getTime() - 12 * 3600000).toISOString() }),
        createWeatherData({ precipitation: 0.05, timestamp: new Date(now.getTime() - 24 * 3600000).toISOString() })
      ];

      const result = ConditionCalculator.calculate48HourRain(data);

      expect(result).toBeCloseTo(0.5, 2);
    });
  });

  describe('real-world scenarios', () => {
    it('should handle typical morning after rain', () => {
      const current = createWeatherData({
        precipitation: 0,
        temperature: 65,
        cloud_cover: 30,
        wind_speed: 8,
        humidity: 75
      });

      const recent = [
        createWeatherData({ precipitation: 0.05 }),
        createWeatherData({ precipitation: 0.03 })
      ];

      const result = ConditionCalculator.calculateCondition(current, recent);

      // Should be good or marginal depending on drying conditions
      expect(['good', 'marginal']).toContain(result.level);
    });

    it('should handle cold windy day with no precipitation', () => {
      const weather = createWeatherData({
        temperature: 35, // Cold (marginal)
        wind_speed: 18,  // Moderate wind (marginal)
        precipitation: 0
      });

      const result = ConditionCalculator.calculateCondition(weather);

      // 2 marginal factors = bad
      expect(result.level).toBe('bad');
      expect(result.reasons).toContain('Cold (35°F)');
      expect(result.reasons).toContain('Moderate winds (18mph)');
    });

    it('should handle hot humid afternoon', () => {
      const weather = createWeatherData({
        temperature: 82, // Warm (71-85°F, marginal)
        humidity: 88,    // High humidity (marginal)
        wind_speed: 4
      });

      const result = ConditionCalculator.calculateCondition(weather);

      // 2 marginal factors = bad
      expect(result.level).toBe('bad');
      expect(result.reasons).toContain('Warm (82°F)');
      expect(result.reasons).toContain('High humidity (88%)');
    });

    it('should handle stormy conditions', () => {
      const weather = createWeatherData({
        temperature: 55,
        precipitation: 0.2,
        wind_speed: 28,
        humidity: 95
      });

      const result = ConditionCalculator.calculateCondition(weather);

      expect(result.level).toBe('bad');
      expect(result.reasons.length).toBeGreaterThan(1);
    });
  });
});
