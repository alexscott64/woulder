/**
 * WindAnalyzer - Minimal utility functions for wind display
 *
 * NOTE: Wind condition assessment logic has been moved to the backend.
 * This file only contains display utilities needed by the frontend UI.
 */
export class WindAnalyzer {
  /**
   * Convert wind direction degrees to compass bearing (N, NE, E, etc.)
   * Used for displaying wind direction in the UI
   */
  static degreesToCompass(degrees: number): string {
    const directions = ['N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW'];
    const index = Math.round(degrees / 45) % 8;
    return directions[index];
  }
}
