/**
 * Weather Analyzers
 *
 * NOTE: Core condition calculation logic has been moved to the backend Go service.
 * These are minimal utility classes kept only for:
 * - Fallback condition calculation when backend data is unavailable
 * - UI display utilities (colors, compass directions, etc.)
 *
 * The authoritative source for weather conditions is now:
 * backend/internal/weather/conditions.go
 */

export { ConditionCalculator } from './ConditionCalculator';
export { TemperatureAnalyzer } from './TemperatureAnalyzer';
export { WindAnalyzer } from './WindAnalyzer';
