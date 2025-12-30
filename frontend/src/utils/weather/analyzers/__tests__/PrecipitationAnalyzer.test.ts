import { describe, it, expect } from 'vitest';
import { PrecipitationAnalyzer } from '../PrecipitationAnalyzer';
import { WeatherData } from '../../../../types/weather';

describe('PrecipitationAnalyzer', () => {
  // Helper to create mock weather data
  const createWeatherData = (overrides: Partial<WeatherData> = {}): WeatherData => ({
    timestamp: new Date().toISOString(),
    temperature: 60,
    feels_like: 60,
    precipitation: 0,
    humidity: 50,
    wind_speed: 10,
    wind_direction: 180,
    cloud_cover: 30,
    pressure: 1013,
    description: 'Clear',
    icon: '01d',
    ...overrides
  });

  describe('getHourlyRate', () => {
    it('should return hourly precipitation (already hourly, no conversion)', () => {
      expect(PrecipitationAnalyzer.getHourlyRate(0.01)).toBe(0.01);
      expect(PrecipitationAnalyzer.getHourlyRate(0.05)).toBe(0.05);
      expect(PrecipitationAnalyzer.getHourlyRate(0.1)).toBe(0.1);
    });

    it('should handle zero precipitation', () => {
      expect(PrecipitationAnalyzer.getHourlyRate(0)).toBe(0);
    });
  });

  describe('getTotalPrecipitation', () => {
    it('should sum precipitation from multiple periods', () => {
      const data = [
        createWeatherData({ precipitation: 0.05 }),
        createWeatherData({ precipitation: 0.10 }),
        createWeatherData({ precipitation: 0.03 })
      ];

      expect(PrecipitationAnalyzer.getTotalPrecipitation(data)).toBeCloseTo(0.18, 2);
    });

    it('should return 0 for empty array', () => {
      expect(PrecipitationAnalyzer.getTotalPrecipitation([])).toBe(0);
    });
  });

  describe('hasPersistentPrecipitation', () => {
    it('should detect persistent drizzle (2+ periods)', () => {
      const persistent = [
        createWeatherData({ precipitation: 0.02 }),
        createWeatherData({ precipitation: 0.02 })
      ];

      expect(PrecipitationAnalyzer.hasPersistentPrecipitation(persistent)).toBe(true);
    });

    it('should not flag single period of rain as persistent', () => {
      const single = [
        createWeatherData({ precipitation: 0.05 }),
        createWeatherData({ precipitation: 0 })
      ];

      expect(PrecipitationAnalyzer.hasPersistentPrecipitation(single)).toBe(false);
    });

    it('should filter out trace amounts below threshold', () => {
      const trace = [
        createWeatherData({ precipitation: 0.005 }),
        createWeatherData({ precipitation: 0.005 })
      ];

      expect(PrecipitationAnalyzer.hasPersistentPrecipitation(trace)).toBe(false);
    });
  });

  describe('getIntensity', () => {
    it('should categorize none', () => {
      expect(PrecipitationAnalyzer.getIntensity(0)).toBe('none');
    });

    it('should categorize light rain (0-0.1 in/hr)', () => {
      expect(PrecipitationAnalyzer.getIntensity(0.01)).toBe('light');
      expect(PrecipitationAnalyzer.getIntensity(0.05)).toBe('light');
      expect(PrecipitationAnalyzer.getIntensity(0.1)).toBe('light');
    });

    it('should categorize moderate rain (0.1-0.3 in/hr)', () => {
      expect(PrecipitationAnalyzer.getIntensity(0.15)).toBe('moderate');
      expect(PrecipitationAnalyzer.getIntensity(0.25)).toBe('moderate');
      expect(PrecipitationAnalyzer.getIntensity(0.3)).toBe('moderate');
    });

    it('should categorize heavy rain (>0.3 in/hr)', () => {
      expect(PrecipitationAnalyzer.getIntensity(0.31)).toBe('heavy');
      expect(PrecipitationAnalyzer.getIntensity(0.5)).toBe('heavy');
      expect(PrecipitationAnalyzer.getIntensity(1.0)).toBe('heavy');
    });
  });

  describe('assessDryingConditions', () => {
    it('should recognize good drying conditions (warm, sunny, breezy)', () => {
      const good = createWeatherData({
        temperature: 65,
        cloud_cover: 20,
        wind_speed: 10
      });

      const result = PrecipitationAnalyzer.assessDryingConditions(good);
      expect(result.canDryQuickly).toBe(true);
      expect(result.factors).toContain('warm temperature');
      expect(result.factors).toContain('sunny (UV helps dry)');
      expect(result.factors).toContain('moderate wind');
    });

    it('should recognize poor drying conditions (cool, overcast, calm)', () => {
      const poor = createWeatherData({
        temperature: 45,
        cloud_cover: 90,
        wind_speed: 3
      });

      const result = PrecipitationAnalyzer.assessDryingConditions(poor);
      expect(result.canDryQuickly).toBe(false);
    });

    it('should require at least 2 of 3 factors for quick drying', () => {
      // Only 1 factor: warm
      const oneFactorGood = createWeatherData({
        temperature: 65,
        cloud_cover: 90,
        wind_speed: 3
      });
      expect(PrecipitationAnalyzer.assessDryingConditions(oneFactorGood).canDryQuickly).toBe(false);

      // 2 factors: warm + sunny
      const twoFactorsGood = createWeatherData({
        temperature: 65,
        cloud_cover: 20,
        wind_speed: 3
      });
      expect(PrecipitationAnalyzer.assessDryingConditions(twoFactorsGood).canDryQuickly).toBe(true);
    });
  });

  describe('assessCondition', () => {
    it('should rate heavy rain as bad (>0.3 in/hr)', () => {
      const heavyRain = createWeatherData({ precipitation: 0.35 });
      const result = PrecipitationAnalyzer.assessCondition(heavyRain);

      expect(result.level).toBe('bad');
      expect(result.reason).toContain('Heavy rain');
      expect(result.reason).toContain('in/hr');
    });

    it('should rate moderate rain as marginal (0.1-0.3 in/hr)', () => {
      const moderateRain = createWeatherData({ precipitation: 0.15 });
      const result = PrecipitationAnalyzer.assessCondition(moderateRain);

      expect(result.level).toBe('marginal');
      expect(result.reason).toContain('Moderate rain');
      expect(result.reason).toContain('in/hr');
    });

    it('should rate persistent drizzle as marginal', () => {
      const current = createWeatherData({ precipitation: 0.02 });
      const recent = [
        createWeatherData({ precipitation: 0.02 }),
        createWeatherData({ precipitation: 0.02 })
      ];

      const result = PrecipitationAnalyzer.assessCondition(current, recent);

      expect(result.level).toBe('marginal');
      expect(result.reason).toContain('Persistent drizzle');
    });

    it('should rate brief light rain with good drying as good', () => {
      const current = createWeatherData({
        precipitation: 0.02,
        temperature: 70,
        cloud_cover: 20,
        wind_speed: 10
      });

      const result = PrecipitationAnalyzer.assessCondition(current);

      expect(result.level).toBe('good');
      expect(result.reason).toContain('drying fast');
    });

    it('should rate recent rain with poor drying as marginal', () => {
      const now = new Date();
      const current = createWeatherData({
        precipitation: 0,
        temperature: 45,
        cloud_cover: 90,
        wind_speed: 2
      });
      // Need enough recent rain in last 24h to trigger warning (>0.05 in total)
      const recent = [
        createWeatherData({
          precipitation: 0.03,
          timestamp: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString() // 2h ago
        }),
        createWeatherData({
          precipitation: 0.03,
          timestamp: new Date(now.getTime() - 4 * 60 * 60 * 1000).toISOString() // 4h ago
        })
      ];

      const result = PrecipitationAnalyzer.assessCondition(current, recent);

      expect(result.level).toBe('marginal');
      expect(result.reason).toContain('Drying slowly');
      expect(result.reason).toContain('last 24h');
    });

    it('should rate no precipitation as good', () => {
      const noPrecip = createWeatherData({ precipitation: 0 });
      const result = PrecipitationAnalyzer.assessCondition(noPrecip);

      expect(result.level).toBe('good');
      expect(result.reason).toBe(null);
    });

    it('should only consider last 24h, not all historical data (regression test)', () => {
      const now = new Date();
      const current = createWeatherData({
        precipitation: 0,
        temperature: 45,
        cloud_cover: 90,
        wind_speed: 2
      });

      // Simulate 14 days of historical data with 0.5in/day average
      // This totals 7 inches over 14 days, but should NOT trigger "drying slowly"
      // because it's old rain (>24h ago)
      const oldRain = Array.from({ length: 336 }, (_, i) => // 14 days * 24 hours
        createWeatherData({
          precipitation: 0.02, // Small amount per hour
          timestamp: new Date(now.getTime() - (i + 25) * 60 * 60 * 1000).toISOString() // 25+ hours ago
        })
      );

      const result = PrecipitationAnalyzer.assessCondition(current, oldRain);

      // Should be "good" because no rain in last 24 hours
      expect(result.level).toBe('good');
      expect(result.reason).toBe(null);
    });
  });

  describe('getPrecipitationInWindow', () => {
    it('should calculate total precipitation in time window', () => {
      const now = new Date();
      const data = [
        createWeatherData({
          precipitation: 0.05,
          timestamp: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString() // 2h ago
        }),
        createWeatherData({
          precipitation: 0.10,
          timestamp: new Date(now.getTime() - 10 * 60 * 60 * 1000).toISOString() // 10h ago
        }),
        createWeatherData({
          precipitation: 0.03,
          timestamp: new Date(now.getTime() - 50 * 60 * 60 * 1000).toISOString() // 50h ago
        })
      ];

      // Last 48 hours should include first two
      const total = PrecipitationAnalyzer.getPrecipitationInWindow(data, 48);
      expect(total).toBeCloseTo(0.15, 2);
    });

    it('should not overcount precipitation when no rain fell (regression test)', () => {
      const now = new Date();
      // Simulate 48 hours of no rain (hourly data points with 0 precipitation)
      const noRainData = Array.from({ length: 48 }, (_, i) =>
        createWeatherData({
          precipitation: 0,
          timestamp: new Date(now.getTime() - i * 60 * 60 * 1000).toISOString()
        })
      );

      const total = PrecipitationAnalyzer.getPrecipitationInWindow(noRainData, 48);
      expect(total).toBe(0);
    });

    it('should correctly sum hourly precipitation values (regression test)', () => {
      const now = new Date();
      // 7 hours of 0.1 in/hr rain = 0.7 inches total
      const sevenHoursRain = Array.from({ length: 7 }, (_, i) =>
        createWeatherData({
          precipitation: 0.1,
          timestamp: new Date(now.getTime() - i * 60 * 60 * 1000).toISOString()
        })
      );

      const total = PrecipitationAnalyzer.getTotalPrecipitation(sevenHoursRain);
      expect(total).toBeCloseTo(0.7, 2);

      // Verify this doesn't get incorrectly multiplied by 3
      expect(total).not.toBeCloseTo(2.1, 2);
    });
  });
});
