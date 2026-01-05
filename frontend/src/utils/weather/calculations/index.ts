/**
 * Weather Calculations - Pure Domain Math
 *
 * This module previously contained pure calculation functions for weather data.
 *
 * As of January 2026, all core weather calculations have been moved to the backend
 * for consistency and to provide a single source of truth:
 * - Precipitation calculations → backend/internal/weather/calculator/precipitation.go
 * - Temperature calculations → backend/internal/weather/conditions.go
 * - Wind calculations → backend/internal/weather/conditions.go
 * - Snow calculations → backend/internal/weather/calculator/snow.go
 *
 * All weather condition assessments are now calculated server-side and provided
 * in the API response.
 */
