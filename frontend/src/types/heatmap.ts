// Heat Map Types for Climbing Activity Visualization

export interface HeatMapPoint {
  mp_area_id: number;
  name: string;
  latitude: number;
  longitude: number;
  activity_score: number;
  total_ticks: number;
  active_routes: number;
  last_activity: string; // ISO date string
  unique_climbers: number;
  has_subareas: boolean;
}

export interface TickDetail {
  mp_route_id: number;
  route_name: string;
  rating: string;
  user_name: string;
  climbed_at: string; // ISO date string
  style: string;
  comment: string;
}

export interface CommentSummary {
  id: number;
  user_name: string;
  comment_text: string;
  commented_at: string; // ISO date string
  mp_route_id?: number;
  route_name?: string;
}

export interface DailyActivity {
  date: string; // ISO date string
  tick_count: number;
  route_count: number;
}

export interface TopRouteSummary {
  mp_route_id: number;
  name: string;
  rating: string;
  tick_count: number;
  last_activity: string; // ISO date string
}

export interface AreaActivityDetail {
  mp_area_id: number;
  name: string;
  parent_mp_area_id?: number;
  latitude?: number;
  longitude?: number;
  total_ticks: number;
  active_routes: number;
  unique_climbers: number;
  last_activity: string; // ISO date string
  recent_ticks: TickDetail[];
  recent_comments: CommentSummary[];
  activity_timeline: DailyActivity[];
  top_routes: TopRouteSummary[];
}

export interface RouteActivity {
  mp_route_id: number;
  name: string;
  rating: string;
  latitude?: number;
  longitude?: number;
  tick_count: number;
  last_activity: string; // ISO date string
  mp_area_id: number;
  area_name: string;
}

export interface HeatMapActivityResponse {
  points: HeatMapPoint[];
  count: number;
  filters: {
    start_date: string;
    end_date: string;
    min_activity: number;
    limit: number;
  };
}

export interface RoutesResponse {
  routes: RouteActivity[];
  count: number;
}

export interface GeoBounds {
  minLat: number;
  maxLat: number;
  minLon: number;
  maxLon: number;
}
