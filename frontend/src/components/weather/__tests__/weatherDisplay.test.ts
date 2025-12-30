import { describe, it, expect } from 'vitest';
import {
  getConditionColor,
  getConditionTextColor,
  getConditionLabel,
  getConditionBadgeStyles,
  getWeatherIconUrl,
  getSnowDepthColor,
  getSnowDescription
} from '../weatherDisplay';
import { ConditionLevel } from '../../../types/weather';

describe('weatherDisplay', () => {
  describe('getConditionColor', () => {
    it('should return green background for good conditions', () => {
      expect(getConditionColor('good')).toBe('bg-green-500');
    });

    it('should return yellow background for marginal conditions', () => {
      expect(getConditionColor('marginal')).toBe('bg-yellow-500');
    });

    it('should return red background for bad conditions', () => {
      expect(getConditionColor('bad')).toBe('bg-red-500');
    });

    it('should return gray background for unknown conditions', () => {
      expect(getConditionColor('unknown' as ConditionLevel)).toBe('bg-gray-500');
    });
  });

  describe('getConditionTextColor', () => {
    it('should return green text colors for good conditions', () => {
      expect(getConditionTextColor('good')).toBe('text-green-600 dark:text-green-400');
    });

    it('should return yellow text colors for marginal conditions', () => {
      expect(getConditionTextColor('marginal')).toBe('text-yellow-600 dark:text-yellow-400');
    });

    it('should return red text colors for bad conditions', () => {
      expect(getConditionTextColor('bad')).toBe('text-red-600 dark:text-red-400');
    });

    it('should return gray text colors for unknown conditions', () => {
      expect(getConditionTextColor('unknown' as ConditionLevel)).toBe('text-gray-600 dark:text-gray-400');
    });
  });

  describe('getConditionLabel', () => {
    it('should return "Good" for good conditions', () => {
      expect(getConditionLabel('good')).toBe('Good');
    });

    it('should return "Fair" for marginal conditions', () => {
      expect(getConditionLabel('marginal')).toBe('Fair');
    });

    it('should return "Poor" for bad conditions', () => {
      expect(getConditionLabel('bad')).toBe('Poor');
    });

    it('should return "Unknown" for unknown conditions', () => {
      expect(getConditionLabel('unknown' as ConditionLevel)).toBe('Unknown');
    });
  });

  describe('getConditionBadgeStyles', () => {
    it('should return green badge styles for good conditions', () => {
      const styles = getConditionBadgeStyles('good');
      expect(styles.bg).toBe('bg-green-100 dark:bg-green-900/30');
      expect(styles.text).toBe('text-green-700 dark:text-green-300');
      expect(styles.border).toBe('border-green-300 dark:border-green-700');
    });

    it('should return yellow badge styles for marginal conditions', () => {
      const styles = getConditionBadgeStyles('marginal');
      expect(styles.bg).toBe('bg-yellow-100 dark:bg-yellow-900/30');
      expect(styles.text).toBe('text-yellow-700 dark:text-yellow-300');
      expect(styles.border).toBe('border-yellow-300 dark:border-yellow-700');
    });

    it('should return red badge styles for bad conditions', () => {
      const styles = getConditionBadgeStyles('bad');
      expect(styles.bg).toBe('bg-red-100 dark:bg-red-900/30');
      expect(styles.text).toBe('text-red-700 dark:text-red-300');
      expect(styles.border).toBe('border-red-300 dark:border-red-700');
    });

    it('should return gray badge styles for unknown conditions', () => {
      const styles = getConditionBadgeStyles('unknown' as ConditionLevel);
      expect(styles.bg).toBe('bg-gray-100 dark:bg-gray-700');
      expect(styles.text).toBe('text-gray-700 dark:text-gray-300');
      expect(styles.border).toBe('border-gray-300 dark:border-gray-600');
    });

    it('should return all three style properties', () => {
      const styles = getConditionBadgeStyles('good');
      expect(styles).toHaveProperty('bg');
      expect(styles).toHaveProperty('text');
      expect(styles).toHaveProperty('border');
      expect(Object.keys(styles)).toHaveLength(3);
    });
  });

  describe('getWeatherIconUrl', () => {
    it('should generate correct OpenWeatherMap icon URL', () => {
      const iconCode = '01d';
      const expectedUrl = 'https://openweathermap.org/img/wn/01d@2x.png';
      expect(getWeatherIconUrl(iconCode)).toBe(expectedUrl);
    });

    it('should handle different icon codes', () => {
      expect(getWeatherIconUrl('10n')).toBe('https://openweathermap.org/img/wn/10n@2x.png');
      expect(getWeatherIconUrl('13d')).toBe('https://openweathermap.org/img/wn/13d@2x.png');
      expect(getWeatherIconUrl('50d')).toBe('https://openweathermap.org/img/wn/50d@2x.png');
    });

    it('should always use @2x resolution', () => {
      const url = getWeatherIconUrl('01d');
      expect(url).toContain('@2x.png');
    });
  });

  describe('getSnowDepthColor', () => {
    it('should return green for no snow', () => {
      expect(getSnowDepthColor(0)).toBe('text-green-600');
    });

    it('should return yellow for light snow (< 3 inches)', () => {
      expect(getSnowDepthColor(1)).toBe('text-yellow-600');
      expect(getSnowDepthColor(2.9)).toBe('text-yellow-600');
    });

    it('should return orange for moderate snow (3-6 inches)', () => {
      expect(getSnowDepthColor(3)).toBe('text-orange-600');
      expect(getSnowDepthColor(5)).toBe('text-orange-600');
      expect(getSnowDepthColor(5.9)).toBe('text-orange-600');
    });

    it('should return red for deep snow (>= 6 inches)', () => {
      expect(getSnowDepthColor(6)).toBe('text-red-600');
      expect(getSnowDepthColor(12)).toBe('text-red-600');
      expect(getSnowDepthColor(24)).toBe('text-red-600');
    });
  });

  describe('getSnowDescription', () => {
    it('should return "No snow" for 0 inches', () => {
      expect(getSnowDescription(0)).toBe('No snow');
    });

    it('should return "Light snow cover" for < 3 inches', () => {
      expect(getSnowDescription(1)).toBe('Light snow cover');
      expect(getSnowDescription(2.5)).toBe('Light snow cover');
    });

    it('should return "Moderate snow" for 3-6 inches', () => {
      expect(getSnowDescription(3)).toBe('Moderate snow');
      expect(getSnowDescription(5)).toBe('Moderate snow');
      expect(getSnowDescription(5.9)).toBe('Moderate snow');
    });

    it('should return "Heavy snow" for 6-12 inches', () => {
      expect(getSnowDescription(6)).toBe('Heavy snow');
      expect(getSnowDescription(10)).toBe('Heavy snow');
      expect(getSnowDescription(11.9)).toBe('Heavy snow');
    });

    it('should return "Very deep snow" for >= 12 inches', () => {
      expect(getSnowDescription(12)).toBe('Very deep snow');
      expect(getSnowDescription(24)).toBe('Very deep snow');
      expect(getSnowDescription(48)).toBe('Very deep snow');
    });
  });
});
