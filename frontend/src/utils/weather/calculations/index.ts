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
 */

export * as precipitation from './precipitation';
export * as temperature from './temperature';
export * as wind from './wind';
export * from './snow';
