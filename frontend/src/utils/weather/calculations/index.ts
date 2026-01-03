/**
 * Weather Calculations - Pure Domain Math
 *
 * This module contains pure calculation functions for weather data.
 * These functions perform domain-specific math and conversions WITHOUT
 * any business logic or climbing-specific assessments.
 *
 * Separation of concerns:
 * - calculations/ - Pure math (this module)
 * - analyzers/ - Climbing condition assessments (business logic)
 * - components/weather/ - UI presentation
 *
 * Note: Snow calculations have been moved to the backend as of Jan 2026.
 * Snow depth is now calculated server-side with proper historical data
 * and provided in the API response.
 */

export * as precipitation from './precipitation';
export * as temperature from './temperature';
export * as wind from './wind';
