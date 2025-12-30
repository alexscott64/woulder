import { describe, it, expect } from 'vitest';
import { TemperatureAnalyzer } from '../TemperatureAnalyzer';

describe('TemperatureAnalyzer', () => {
  describe('getCategory', () => {
    it('should categorize temperature as too-cold', () => {
      expect(TemperatureAnalyzer.getCategory(29)).toBe('too-cold');
      expect(TemperatureAnalyzer.getCategory(0)).toBe('too-cold');
      expect(TemperatureAnalyzer.getCategory(-10)).toBe('too-cold');
    });

    it('should categorize temperature as cold', () => {
      expect(TemperatureAnalyzer.getCategory(30)).toBe('cold');
      expect(TemperatureAnalyzer.getCategory(35)).toBe('cold');
      expect(TemperatureAnalyzer.getCategory(40)).toBe('cold');
    });

    it('should categorize temperature as ideal', () => {
      expect(TemperatureAnalyzer.getCategory(41)).toBe('ideal');
      expect(TemperatureAnalyzer.getCategory(50)).toBe('ideal');
      expect(TemperatureAnalyzer.getCategory(65)).toBe('ideal');
    });

    it('should categorize temperature as warm', () => {
      expect(TemperatureAnalyzer.getCategory(66)).toBe('warm');
      expect(TemperatureAnalyzer.getCategory(70)).toBe('warm');
      expect(TemperatureAnalyzer.getCategory(79)).toBe('warm');
    });

    it('should categorize temperature as too-hot', () => {
      expect(TemperatureAnalyzer.getCategory(80)).toBe('too-hot');
      expect(TemperatureAnalyzer.getCategory(90)).toBe('too-hot');
      expect(TemperatureAnalyzer.getCategory(100)).toBe('too-hot');
    });
  });

  describe('getColor', () => {
    it('should return green for ideal temperatures', () => {
      expect(TemperatureAnalyzer.getColor(50)).toBe('text-green-600 dark:text-green-400');
      expect(TemperatureAnalyzer.getColor(60)).toBe('text-green-600 dark:text-green-400');
    });

    it('should return yellow for cold/warm temperatures', () => {
      expect(TemperatureAnalyzer.getColor(35)).toBe('text-yellow-600 dark:text-yellow-400');
      expect(TemperatureAnalyzer.getColor(70)).toBe('text-yellow-600 dark:text-yellow-400');
    });

    it('should return red for too-cold/too-hot temperatures', () => {
      expect(TemperatureAnalyzer.getColor(25)).toBe('text-red-600 dark:text-red-400');
      expect(TemperatureAnalyzer.getColor(85)).toBe('text-red-600 dark:text-red-400');
    });
  });

  describe('assessCondition', () => {
    it('should assess too-cold temperature as bad', () => {
      const result = TemperatureAnalyzer.assessCondition(28);
      expect(result.level).toBe('bad');
      expect(result.reason).toBe('Too cold (28°F)');
    });

    it('should assess cold temperature as marginal', () => {
      const result = TemperatureAnalyzer.assessCondition(38);
      expect(result.level).toBe('marginal');
      expect(result.reason).toBe('Cold (38°F)');
    });

    it('should assess ideal temperature as good', () => {
      const result = TemperatureAnalyzer.assessCondition(55);
      expect(result.level).toBe('good');
      expect(result.reason).toBeNull();
    });

    it('should assess warm temperature as marginal', () => {
      const result = TemperatureAnalyzer.assessCondition(72);
      expect(result.level).toBe('marginal');
      expect(result.reason).toBe('Warm (72°F)');
    });

    it('should assess too-hot temperature as bad', () => {
      const result = TemperatureAnalyzer.assessCondition(85);
      expect(result.level).toBe('bad');
      expect(result.reason).toBe('Too hot (85°F)');
    });
  });

  describe('fahrenheitToCelsius', () => {
    it('should convert Fahrenheit to Celsius correctly', () => {
      expect(TemperatureAnalyzer.fahrenheitToCelsius(32)).toBeCloseTo(0, 1);
      expect(TemperatureAnalyzer.fahrenheitToCelsius(212)).toBeCloseTo(100, 1);
      expect(TemperatureAnalyzer.fahrenheitToCelsius(68)).toBeCloseTo(20, 1);
      expect(TemperatureAnalyzer.fahrenheitToCelsius(-40)).toBeCloseTo(-40, 1);
    });
  });

  describe('celsiusToFahrenheit', () => {
    it('should convert Celsius to Fahrenheit correctly', () => {
      expect(TemperatureAnalyzer.celsiusToFahrenheit(0)).toBeCloseTo(32, 1);
      expect(TemperatureAnalyzer.celsiusToFahrenheit(100)).toBeCloseTo(212, 1);
      expect(TemperatureAnalyzer.celsiusToFahrenheit(20)).toBeCloseTo(68, 1);
      expect(TemperatureAnalyzer.celsiusToFahrenheit(-40)).toBeCloseTo(-40, 1);
    });
  });

  describe('isFreezing', () => {
    it('should return true for temperatures at or below freezing', () => {
      expect(TemperatureAnalyzer.isFreezing(32)).toBe(true);
      expect(TemperatureAnalyzer.isFreezing(30)).toBe(true);
      expect(TemperatureAnalyzer.isFreezing(0)).toBe(true);
    });

    it('should return false for temperatures above freezing', () => {
      expect(TemperatureAnalyzer.isFreezing(33)).toBe(false);
      expect(TemperatureAnalyzer.isFreezing(40)).toBe(false);
      expect(TemperatureAnalyzer.isFreezing(70)).toBe(false);
    });
  });

  describe('supportsDrying', () => {
    it('should return true for temperatures above 55°F', () => {
      expect(TemperatureAnalyzer.supportsDrying(56)).toBe(true);
      expect(TemperatureAnalyzer.supportsDrying(70)).toBe(true);
      expect(TemperatureAnalyzer.supportsDrying(90)).toBe(true);
    });

    it('should return false for temperatures at or below 55°F', () => {
      expect(TemperatureAnalyzer.supportsDrying(55)).toBe(false);
      expect(TemperatureAnalyzer.supportsDrying(50)).toBe(false);
      expect(TemperatureAnalyzer.supportsDrying(30)).toBe(false);
    });
  });

  describe('adjustForElevation', () => {
    it('should decrease temperature with elevation gain', () => {
      // At 1000ft elevation, expect ~3.5°F drop
      expect(TemperatureAnalyzer.adjustForElevation(70, 1000)).toBeCloseTo(66.5, 1);

      // At 5000ft elevation, expect ~17.5°F drop
      expect(TemperatureAnalyzer.adjustForElevation(70, 5000)).toBeCloseTo(52.5, 1);

      // At 10000ft elevation, expect ~35°F drop
      expect(TemperatureAnalyzer.adjustForElevation(70, 10000)).toBeCloseTo(35, 1);
    });

    it('should handle sea level correctly', () => {
      expect(TemperatureAnalyzer.adjustForElevation(70, 0)).toBe(70);
    });

    it('should handle negative elevations (below sea level)', () => {
      // Death Valley (-280ft) should be slightly warmer
      expect(TemperatureAnalyzer.adjustForElevation(100, -280)).toBeCloseTo(100.98, 1);
    });
  });

  describe('calculateFeelsLike', () => {
    describe('wind chill (cold + wind)', () => {
      it('should calculate wind chill for cold windy conditions', () => {
        const feelsLike = TemperatureAnalyzer.calculateFeelsLike(40, 15, 50);
        expect(feelsLike).toBeLessThan(40); // Should feel colder
        expect(feelsLike).toBeCloseTo(32, 0); // Roughly 32°F
      });

      it('should calculate stronger wind chill with higher wind speeds', () => {
        const feelsLike10mph = TemperatureAnalyzer.calculateFeelsLike(30, 10, 50);
        const feelsLike20mph = TemperatureAnalyzer.calculateFeelsLike(30, 20, 50);
        expect(feelsLike20mph).toBeLessThan(feelsLike10mph); // More wind = colder
      });

      it('should not apply wind chill when temp is above 50°F', () => {
        const feelsLike = TemperatureAnalyzer.calculateFeelsLike(60, 15, 50);
        expect(feelsLike).toBe(60); // No adjustment
      });

      it('should not apply wind chill when wind is very light', () => {
        const feelsLike = TemperatureAnalyzer.calculateFeelsLike(40, 2, 50);
        expect(feelsLike).toBe(40); // No adjustment
      });
    });

    describe('heat index (hot + humidity)', () => {
      it('should calculate heat index for hot humid conditions', () => {
        const feelsLike = TemperatureAnalyzer.calculateFeelsLike(90, 5, 70);
        expect(feelsLike).toBeGreaterThan(90); // Should feel hotter
        expect(feelsLike).toBeCloseTo(106, 0); // Roughly 106°F
      });

      it('should calculate stronger heat index with higher humidity', () => {
        const feelsLike50 = TemperatureAnalyzer.calculateFeelsLike(90, 5, 50);
        const feelsLike80 = TemperatureAnalyzer.calculateFeelsLike(90, 5, 80);
        expect(feelsLike80).toBeGreaterThan(feelsLike50); // More humidity = hotter
      });

      it('should not apply heat index when temp is below 80°F', () => {
        const feelsLike = TemperatureAnalyzer.calculateFeelsLike(75, 5, 70);
        expect(feelsLike).toBe(75); // No adjustment
      });

      it('should not apply heat index when humidity is low', () => {
        const feelsLike = TemperatureAnalyzer.calculateFeelsLike(90, 5, 30);
        expect(feelsLike).toBe(90); // No adjustment
      });
    });

    describe('comfortable range (no adjustment)', () => {
      it('should return actual temperature for comfortable conditions', () => {
        expect(TemperatureAnalyzer.calculateFeelsLike(65, 5, 50)).toBe(65);
        expect(TemperatureAnalyzer.calculateFeelsLike(70, 8, 45)).toBe(70);
        expect(TemperatureAnalyzer.calculateFeelsLike(55, 10, 60)).toBe(55);
      });
    });
  });

  describe('edge cases', () => {
    it('should handle boundary temperatures correctly', () => {
      // Test exact boundary values
      expect(TemperatureAnalyzer.getCategory(30)).toBe('cold');
      expect(TemperatureAnalyzer.getCategory(41)).toBe('ideal');
      expect(TemperatureAnalyzer.getCategory(65)).toBe('ideal');
      expect(TemperatureAnalyzer.getCategory(66)).toBe('warm');
      expect(TemperatureAnalyzer.getCategory(79)).toBe('warm');
      expect(TemperatureAnalyzer.getCategory(80)).toBe('too-hot');
    });

    it('should handle extreme temperatures', () => {
      expect(TemperatureAnalyzer.getCategory(-50)).toBe('too-cold');
      expect(TemperatureAnalyzer.getCategory(150)).toBe('too-hot');
    });

    it('should round temperatures consistently in condition reasons', () => {
      const result1 = TemperatureAnalyzer.assessCondition(38.4);
      expect(result1.reason).toBe('Cold (38°F)');

      const result2 = TemperatureAnalyzer.assessCondition(38.6);
      expect(result2.reason).toBe('Cold (39°F)');
    });
  });
});
