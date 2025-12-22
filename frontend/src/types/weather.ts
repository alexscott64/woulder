export interface Location {
  id: number;
  name: string;
  latitude: number;
  longitude: number;
  elevation_ft: number; // Elevation in feet above sea level
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

export interface WeatherForecast {
  location_id: number;
  location: Location;
  current: WeatherData;
  hourly: WeatherData[];
  historical: WeatherData[];
}

export interface AllWeatherResponse {
  weather: WeatherForecast[];
  updated_at: string;
}

export type ConditionLevel = 'good' | 'marginal' | 'bad';

export interface WeatherCondition {
  level: ConditionLevel;
  reasons: string[];
}
