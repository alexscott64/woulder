import { describe, it, expect } from 'vitest';
import { WindAnalyzer } from '../WindAnalyzer';

describe('WindAnalyzer', () => {
  describe('getCategory', () => {
    it('should categorize wind as calm', () => {
      expect(WindAnalyzer.getCategory(0)).toBe('calm');
      expect(WindAnalyzer.getCategory(3)).toBe('calm');
      expect(WindAnalyzer.getCategory(5)).toBe('calm');
    });

    it('should categorize wind as light', () => {
      expect(WindAnalyzer.getCategory(6)).toBe('light');
      expect(WindAnalyzer.getCategory(10)).toBe('light');
      expect(WindAnalyzer.getCategory(12)).toBe('light');
    });

    it('should categorize wind as moderate', () => {
      expect(WindAnalyzer.getCategory(13)).toBe('moderate');
      expect(WindAnalyzer.getCategory(15)).toBe('moderate');
      expect(WindAnalyzer.getCategory(20)).toBe('moderate');
    });

    it('should categorize wind as strong', () => {
      expect(WindAnalyzer.getCategory(21)).toBe('strong');
      expect(WindAnalyzer.getCategory(25)).toBe('strong');
      expect(WindAnalyzer.getCategory(30)).toBe('strong');
    });

    it('should categorize wind as dangerous', () => {
      expect(WindAnalyzer.getCategory(31)).toBe('dangerous');
      expect(WindAnalyzer.getCategory(40)).toBe('dangerous');
      expect(WindAnalyzer.getCategory(50)).toBe('dangerous');
    });
  });

  describe('getColor', () => {
    it('should return green for calm/light winds', () => {
      expect(WindAnalyzer.getColor(3)).toBe('text-green-600 dark:text-green-400');
      expect(WindAnalyzer.getColor(10)).toBe('text-green-600 dark:text-green-400');
    });

    it('should return yellow for moderate winds', () => {
      expect(WindAnalyzer.getColor(15)).toBe('text-yellow-600 dark:text-yellow-400');
      expect(WindAnalyzer.getColor(18)).toBe('text-yellow-600 dark:text-yellow-400');
    });

    it('should return red for strong/dangerous winds', () => {
      expect(WindAnalyzer.getColor(25)).toBe('text-red-600 dark:text-red-400');
      expect(WindAnalyzer.getColor(40)).toBe('text-red-600 dark:text-red-400');
    });
  });

  describe('assessCondition', () => {
    it('should assess calm winds as good', () => {
      const result = WindAnalyzer.assessCondition(4);
      expect(result.level).toBe('good');
      expect(result.reason).toBeNull();
    });

    it('should assess light winds as good', () => {
      const result = WindAnalyzer.assessCondition(10);
      expect(result.level).toBe('good');
      expect(result.reason).toBeNull();
    });

    it('should assess moderate winds as marginal', () => {
      const result = WindAnalyzer.assessCondition(15);
      expect(result.level).toBe('marginal');
      expect(result.reason).toBe('Moderate winds (15mph)');
    });

    it('should assess strong winds as bad', () => {
      const result = WindAnalyzer.assessCondition(25);
      expect(result.level).toBe('bad');
      expect(result.reason).toBe('High winds (25mph)');
    });

    it('should assess dangerous winds as bad', () => {
      const result = WindAnalyzer.assessCondition(35);
      expect(result.level).toBe('bad');
      expect(result.reason).toBe('Dangerous winds (35mph)');
    });
  });

  describe('aidsDrying', () => {
    it('should return true for ideal drying winds (6-19 mph)', () => {
      expect(WindAnalyzer.aidsDrying(6)).toBe(true);
      expect(WindAnalyzer.aidsDrying(10)).toBe(true);
      expect(WindAnalyzer.aidsDrying(15)).toBe(true);
      expect(WindAnalyzer.aidsDrying(19)).toBe(true);
    });

    it('should return false for calm winds (too little)', () => {
      expect(WindAnalyzer.aidsDrying(0)).toBe(false);
      expect(WindAnalyzer.aidsDrying(3)).toBe(false);
      expect(WindAnalyzer.aidsDrying(5)).toBe(false);
    });

    it('should return false for strong winds (too much)', () => {
      expect(WindAnalyzer.aidsDrying(20)).toBe(false);
      expect(WindAnalyzer.aidsDrying(25)).toBe(false);
      expect(WindAnalyzer.aidsDrying(30)).toBe(false);
    });
  });

  describe('degreesToCompass', () => {
    it('should convert degrees to compass bearings', () => {
      expect(WindAnalyzer.degreesToCompass(0)).toBe('N');
      expect(WindAnalyzer.degreesToCompass(45)).toBe('NE');
      expect(WindAnalyzer.degreesToCompass(90)).toBe('E');
      expect(WindAnalyzer.degreesToCompass(135)).toBe('SE');
      expect(WindAnalyzer.degreesToCompass(180)).toBe('S');
      expect(WindAnalyzer.degreesToCompass(225)).toBe('SW');
      expect(WindAnalyzer.degreesToCompass(270)).toBe('W');
      expect(WindAnalyzer.degreesToCompass(315)).toBe('NW');
    });

    it('should handle edge cases near boundaries', () => {
      expect(WindAnalyzer.degreesToCompass(22)).toBe('N'); // Closer to N
      expect(WindAnalyzer.degreesToCompass(23)).toBe('NE'); // Closer to NE
      expect(WindAnalyzer.degreesToCompass(360)).toBe('N'); // Full circle
    });

    it('should handle wraparound correctly', () => {
      expect(WindAnalyzer.degreesToCompass(361)).toBe('N');
      expect(WindAnalyzer.degreesToCompass(405)).toBe('NE'); // 405 % 360 = 45
    });
  });

  describe('degreesToCardinal', () => {
    it('should convert degrees to cardinal direction names', () => {
      expect(WindAnalyzer.degreesToCardinal(0)).toBe('North');
      expect(WindAnalyzer.degreesToCardinal(45)).toBe('Northeast');
      expect(WindAnalyzer.degreesToCardinal(90)).toBe('East');
      expect(WindAnalyzer.degreesToCardinal(135)).toBe('Southeast');
      expect(WindAnalyzer.degreesToCardinal(180)).toBe('South');
      expect(WindAnalyzer.degreesToCardinal(225)).toBe('Southwest');
      expect(WindAnalyzer.degreesToCardinal(270)).toBe('West');
      expect(WindAnalyzer.degreesToCardinal(315)).toBe('Northwest');
    });
  });

  describe('unit conversions', () => {
    describe('mphToKmh', () => {
      it('should convert mph to km/h correctly', () => {
        expect(WindAnalyzer.mphToKmh(0)).toBeCloseTo(0, 1);
        expect(WindAnalyzer.mphToKmh(10)).toBeCloseTo(16.09, 1);
        expect(WindAnalyzer.mphToKmh(25)).toBeCloseTo(40.23, 1);
        expect(WindAnalyzer.mphToKmh(50)).toBeCloseTo(80.47, 1);
      });
    });

    describe('kmhToMph', () => {
      it('should convert km/h to mph correctly', () => {
        expect(WindAnalyzer.kmhToMph(0)).toBeCloseTo(0, 1);
        expect(WindAnalyzer.kmhToMph(16.09)).toBeCloseTo(10, 1);
        expect(WindAnalyzer.kmhToMph(40.23)).toBeCloseTo(25, 1);
        expect(WindAnalyzer.kmhToMph(80.47)).toBeCloseTo(50, 1);
      });

      it('should be inverse of mphToKmh', () => {
        const mph = 25;
        const kmh = WindAnalyzer.mphToKmh(mph);
        const backToMph = WindAnalyzer.kmhToMph(kmh);
        expect(backToMph).toBeCloseTo(mph, 1);
      });
    });

    describe('mphToKnots', () => {
      it('should convert mph to knots correctly', () => {
        expect(WindAnalyzer.mphToKnots(0)).toBeCloseTo(0, 1);
        expect(WindAnalyzer.mphToKnots(10)).toBeCloseTo(8.69, 1);
        expect(WindAnalyzer.mphToKnots(25)).toBeCloseTo(21.72, 1);
        expect(WindAnalyzer.mphToKnots(50)).toBeCloseTo(43.45, 1);
      });
    });
  });

  describe('getBeaufortScale', () => {
    it('should return correct Beaufort scale for calm conditions', () => {
      const result = WindAnalyzer.getBeaufortScale(0);
      expect(result.scale).toBe(0);
      expect(result.description).toBe('Calm');
    });

    it('should return correct Beaufort scale for light air', () => {
      const result = WindAnalyzer.getBeaufortScale(2);
      expect(result.scale).toBe(1);
      expect(result.description).toBe('Light air');
    });

    it('should return correct Beaufort scale for light breeze', () => {
      const result = WindAnalyzer.getBeaufortScale(5);
      expect(result.scale).toBe(2);
      expect(result.description).toBe('Light breeze');
    });

    it('should return correct Beaufort scale for gentle breeze', () => {
      const result = WindAnalyzer.getBeaufortScale(10);
      expect(result.scale).toBe(3);
      expect(result.description).toBe('Gentle breeze');
    });

    it('should return correct Beaufort scale for moderate breeze', () => {
      const result = WindAnalyzer.getBeaufortScale(15);
      expect(result.scale).toBe(4);
      expect(result.description).toBe('Moderate breeze');
    });

    it('should return correct Beaufort scale for fresh breeze', () => {
      const result = WindAnalyzer.getBeaufortScale(22);
      expect(result.scale).toBe(5);
      expect(result.description).toBe('Fresh breeze');
    });

    it('should return correct Beaufort scale for strong breeze', () => {
      const result = WindAnalyzer.getBeaufortScale(28);
      expect(result.scale).toBe(6);
      expect(result.description).toBe('Strong breeze');
    });

    it('should return correct Beaufort scale for high wind', () => {
      const result = WindAnalyzer.getBeaufortScale(35);
      expect(result.scale).toBe(7);
      expect(result.description).toBe('High wind');
    });

    it('should return correct Beaufort scale for gale', () => {
      const result = WindAnalyzer.getBeaufortScale(42);
      expect(result.scale).toBe(8);
      expect(result.description).toBe('Gale');
    });

    it('should return correct Beaufort scale for strong gale', () => {
      const result = WindAnalyzer.getBeaufortScale(50);
      expect(result.scale).toBe(9);
      expect(result.description).toBe('Strong gale');
    });

    it('should return correct Beaufort scale for storm', () => {
      const result = WindAnalyzer.getBeaufortScale(58);
      expect(result.scale).toBe(10);
      expect(result.description).toBe('Storm');
    });

    it('should return correct Beaufort scale for violent storm', () => {
      const result = WindAnalyzer.getBeaufortScale(68);
      expect(result.scale).toBe(11);
      expect(result.description).toBe('Violent storm');
    });

    it('should return correct Beaufort scale for hurricane', () => {
      const result = WindAnalyzer.getBeaufortScale(75);
      expect(result.scale).toBe(12);
      expect(result.description).toBe('Hurricane');
    });
  });

  describe('isSafeForClimbing', () => {
    it('should return true for calm winds', () => {
      expect(WindAnalyzer.isSafeForClimbing(3)).toBe(true);
    });

    it('should return true for light winds', () => {
      expect(WindAnalyzer.isSafeForClimbing(10)).toBe(true);
    });

    it('should return true for moderate winds at threshold', () => {
      expect(WindAnalyzer.isSafeForClimbing(20)).toBe(true);
    });

    it('should return false for strong winds', () => {
      expect(WindAnalyzer.isSafeForClimbing(21)).toBe(false);
      expect(WindAnalyzer.isSafeForClimbing(25)).toBe(false);
    });

    it('should return false for dangerous winds', () => {
      expect(WindAnalyzer.isSafeForClimbing(35)).toBe(false);
      expect(WindAnalyzer.isSafeForClimbing(50)).toBe(false);
    });
  });

  describe('edge cases', () => {
    it('should handle boundary wind speeds correctly', () => {
      expect(WindAnalyzer.getCategory(5)).toBe('calm');
      expect(WindAnalyzer.getCategory(6)).toBe('light');
      expect(WindAnalyzer.getCategory(12)).toBe('light');
      expect(WindAnalyzer.getCategory(13)).toBe('moderate');
      expect(WindAnalyzer.getCategory(20)).toBe('moderate');
      expect(WindAnalyzer.getCategory(21)).toBe('strong');
      expect(WindAnalyzer.getCategory(30)).toBe('strong');
      expect(WindAnalyzer.getCategory(31)).toBe('dangerous');
    });

    it('should handle negative wind speeds gracefully', () => {
      // While invalid in reality, should not crash
      expect(WindAnalyzer.getCategory(-5)).toBe('calm');
    });

    it('should round wind speeds consistently in condition reasons', () => {
      const result1 = WindAnalyzer.assessCondition(15.4);
      expect(result1.reason).toBe('Moderate winds (15mph)');

      const result2 = WindAnalyzer.assessCondition(15.6);
      expect(result2.reason).toBe('Moderate winds (16mph)');
    });
  });
});
