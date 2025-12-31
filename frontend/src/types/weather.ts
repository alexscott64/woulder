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
}

export interface AllWeatherResponse {
  weather: WeatherForecast[];
  updated_at: string;
}

export type ConditionLevel = 'good' | 'marginal' | 'bad' | 'do_not_climb';

export interface WeatherCondition {
  level: ConditionLevel;
  reasons: string[];
}
