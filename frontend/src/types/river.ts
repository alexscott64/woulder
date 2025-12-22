export interface River {
  id: number;
  location_id: number;
  gauge_id: string;
  river_name: string;
  safe_crossing_cfs: number;
  caution_crossing_cfs: number;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface RiverData {
  river: River;
  flow_cfs: number;
  gauge_height_ft: number;
  is_safe: boolean;
  status: 'safe' | 'caution' | 'unsafe';
  status_message: string;
  timestamp: string;
  percent_of_safe: number;
}

export interface RiversResponse {
  rivers: RiverData[];
  updated_at: string;
}
