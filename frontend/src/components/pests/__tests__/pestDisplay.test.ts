import { describe, it, expect } from 'vitest';
import {
  getPestLevelColor,
  getPestLevelBgColor,
  getPestLevelText
} from '../pestDisplay';
import { PestLevel } from '../../../utils/pests/calculations/pests';

describe('pestDisplay', () => {
  describe('getPestLevelColor', () => {
    it('should return green for low level', () => {
      expect(getPestLevelColor('low')).toBe('text-green-600');
    });

    it('should return yellow for moderate level', () => {
      expect(getPestLevelColor('moderate')).toBe('text-yellow-500');
    });

    it('should return yellow for high level', () => {
      expect(getPestLevelColor('high')).toBe('text-yellow-600');
    });

    it('should return orange for very_high level', () => {
      expect(getPestLevelColor('very_high')).toBe('text-orange-500');
    });

    it('should return red for extreme level', () => {
      expect(getPestLevelColor('extreme')).toBe('text-red-600');
    });
  });

  describe('getPestLevelBgColor', () => {
    it('should return green background for low level', () => {
      expect(getPestLevelBgColor('low')).toBe('bg-green-500');
    });

    it('should return yellow background for moderate level', () => {
      expect(getPestLevelBgColor('moderate')).toBe('bg-yellow-400');
    });

    it('should return yellow background for high level', () => {
      expect(getPestLevelBgColor('high')).toBe('bg-yellow-500');
    });

    it('should return orange background for very_high level', () => {
      expect(getPestLevelBgColor('very_high')).toBe('bg-orange-500');
    });

    it('should return red background for extreme level', () => {
      expect(getPestLevelBgColor('extreme')).toBe('bg-red-500');
    });
  });

  describe('getPestLevelText', () => {
    it('should return proper labels for all levels', () => {
      expect(getPestLevelText('low')).toBe('Low');
      expect(getPestLevelText('moderate')).toBe('Moderate');
      expect(getPestLevelText('high')).toBe('High');
      expect(getPestLevelText('very_high')).toBe('Very High');
      expect(getPestLevelText('extreme')).toBe('Extreme');
    });

    it('should be capitalized', () => {
      const levels: PestLevel[] = ['low', 'moderate', 'high', 'very_high', 'extreme'];
      levels.forEach(level => {
        const text = getPestLevelText(level);
        expect(text[0]).toBe(text[0].toUpperCase());
      });
    });
  });
});
