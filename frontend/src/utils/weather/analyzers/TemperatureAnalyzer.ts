/**
 * TemperatureAnalyzer - Minimal utility functions for temperature display
 *
 * NOTE: Temperature condition assessment logic has been moved to the backend.
 * This file only contains display utilities needed by the frontend UI.
 */
export class TemperatureAnalyzer {
  // Temperature thresholds for display colors
  static readonly COLD_MAX = 40;
  static readonly IDEAL_MAX = 70;
  static readonly WARM_MAX = 85;

  /**
   * Get temperature color for UI display
   * Used for color-coding temperature values in components
   */
  static getColor(tempF: number): string {
    if (tempF < 30) {
      return 'text-red-600 dark:text-red-400'; // Too cold
    } else if (tempF <= this.COLD_MAX) {
      return 'text-yellow-600 dark:text-yellow-400'; // Cold
    } else if (tempF <= this.IDEAL_MAX) {
      return 'text-green-600 dark:text-green-400'; // Ideal
    } else if (tempF <= this.WARM_MAX) {
      return 'text-yellow-600 dark:text-yellow-400'; // Warm
    } else {
      return 'text-red-600 dark:text-red-400'; // Too hot
    }
  }
}
