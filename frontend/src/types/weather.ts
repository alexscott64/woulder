export interface Location {
  id: number;
  name: string;
  latitude: number;
  longitude: number;
  elevation_ft: number; // Elevation in feet above sea level
  area_id: number; // Foreign key to areas table
  created_at: string;
  updated_at: string;
}

export interface WeatherData {
  id?: number;
  location_id?: number;
  timestamp: string;
  temperature: number;
  feels_like: number;
  precipitation: number;
  humidity: number;
  wind_speed: number;
  wind_direction: number;
  cloud_cover: number;
  pressure: number;
  description: string;
  icon: string;
  created_at?: string;
}

export interface DailySunTimes {
  date: string;    // Date in YYYY-MM-DD format
  sunrise: string; // Sunrise time (ISO 8601)
  sunset: string;  // Sunset time (ISO 8601)
}

export interface RockType {
  id: number;
  name: string;
  base_drying_hours: number;
  porosity_percent: number;
  is_wet_sensitive: boolean;
  description: string;
}

export interface RockDryingStatus {
  is_wet: boolean;
  is_safe: boolean;
  is_wet_sensitive: boolean;
  hours_until_dry: number;
  last_rain_timestamp: string;
  status: 'critical' | 'poor' | 'fair' | 'good';
  message: string;
  rock_types: string[];
  primary_rock_type: string;
  primary_group_name: string;
}

export interface ClimbHistoryEntry {
  mp_route_id: number;       // Mountain Project route ID for linking
  route_name: string;
  route_rating: string;
  mp_area_id: number;        // Mountain Project area ID for linking
  area_name: string;         // e.g., "Xyz Boulders"
  climbed_at: string;        // ISO 8601 timestamp
  climbed_by: string;
  style: string;
  comment?: string;
  days_since_climb: number;
}

export interface AreaActivitySummary {
  mp_area_id: number;
  name: string;
  parent_mp_area_id?: number;
  last_climb_at: string;     // ISO 8601 timestamp
  total_ticks: number;
  unique_routes: number;
  days_since_climb: number;
  has_subareas: boolean;
  subarea_count: number;
  drying_stats?: AreaDryingStats; // Area-level drying statistics
}

export interface RouteActivitySummary {
  mp_route_id: number;
  name: string;
  rating: string;
  mp_area_id: number;
  last_climb_at: string;     // ISO 8601 timestamp
  most_recent_tick?: ClimbHistoryEntry; // Null if no ascents
  recent_ticks?: ClimbHistoryEntry[];
  days_since_climb: number;
}

export interface SearchResult {
  result_type: 'area' | 'route';
  id: number;
  name: string;
  rating?: string;           // Only for routes
  mp_area_id: number;
  area_name?: string;        // Only for routes (parent area name)
  last_climb_at: string;     // ISO 8601 timestamp
  days_since_climb: number;
  total_ticks?: number;      // Only for areas
  unique_routes?: number;    // Only for areas
  most_recent_tick?: ClimbHistoryEntry; // Only for routes
}

export interface LastClimbedInfo {
  route_name: string;
  route_rating: string;
  climbed_at: string; // ISO 8601 timestamp
  climbed_by: string;
  style: string;
  comment?: string;
  days_since_climb: number;
}

export interface WeatherForecast {
  location_id: number;
  location: Location;
  current: WeatherData;
  hourly: WeatherData[];
  historical: WeatherData[];
  sunrise?: string;  // Today's sunrise time (ISO 8601)
  sunset?: string;   // Today's sunset time (ISO 8601)
  daily_sun_times?: DailySunTimes[]; // Sunrise/sunset for each forecast day
  rock_drying_status?: RockDryingStatus; // Rock drying status
  snow_depth_inches?: number; // Current snow depth on ground in inches (calculated by backend)
  daily_snow_depth?: Record<string, number>; // Daily snow depth forecast by date (YYYY-MM-DD)
  today_condition?: WeatherCondition; // Today's overall climbing condition (calculated by backend)
  rain_last_48h?: number; // Total rain in last 48 hours (inches, calculated by backend)
  rain_next_48h?: number; // Forecast rain in next 48 hours (inches, calculated by backend)
  pest_conditions?: PestConditions; // Pest activity levels (calculated by backend)
  last_climbed_info?: LastClimbedInfo; // DEPRECATED: Most recent climb (use climb_history instead)
  climb_history?: ClimbHistoryEntry[]; // Recent climb history at this location (from Mountain Project)
}

export interface AllWeatherResponse {
  forecasts: WeatherForecast[];
  count: number;
  updated_at: string;
}

export type ConditionLevel = 'good' | 'marginal' | 'bad' | 'do_not_climb';

export interface WeatherCondition {
  level: ConditionLevel;
  reasons: string[];
}

export type PestLevel = 'low' | 'moderate' | 'high' | 'very_high' | 'extreme';

export interface PestConditions {
  mosquito_level: PestLevel;
  mosquito_score: number; // 0-100
  outdoor_pest_level: PestLevel;
  outdoor_pest_score: number; // 0-100
  factors: string[];
}

export interface DryingForecastPeriod {
  start_time: string;         // ISO 8601 timestamp
  end_time?: string;          // ISO 8601 timestamp (optional for last period)
  is_dry: boolean;
  status: 'dry' | 'drying' | 'wet';
  hours_until_dry?: number;   // Only present when wet/drying
  rain_amount?: number;       // Inches of rain in this period
}

export interface BoulderDryingStatus {
  mp_route_id: number;
  is_wet: boolean;
  is_safe: boolean;
  hours_until_dry: number;
  status: 'critical' | 'poor' | 'fair' | 'good';
  message: string;
  confidence_score: number; // 0-100
  last_rain_timestamp?: string; // Optional - omitted when no recent rain
  sun_exposure_hours: number;
  tree_coverage_percent: number;
  rock_type: string;
  aspect: string; // N, NE, E, SE, S, SW, W, NW
  latitude: number;
  longitude: number;
  forecast?: DryingForecastPeriod[]; // 6-day dry/wet forecast
}

export interface AreaDryingStats {
  total_routes: number;        // Total routes with GPS data
  dry_count: number;            // Routes currently dry
  drying_count: number;         // Routes drying (<24h until dry)
  wet_count: number;            // Routes wet (>24h until dry)
  percent_dry: number;          // Percentage of routes dry (0-100)
  avg_hours_until_dry: number;  // Average hours until dry (wet routes only)
  avg_tree_coverage: number;    // Average tree coverage % (0-100)
  confidence_score: number;     // Overall confidence (0-100)
  last_rain_timestamp?: string; // Most recent rain timestamp from all routes
}
