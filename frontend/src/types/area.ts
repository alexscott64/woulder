export interface Area {
  id: number;
  name: string;
  description?: string;
  region?: string;
  display_order: number;
  location_count: number;
  created_at: string;
  updated_at: string;
}

export interface AreaWithLocations {
  area: Area;
  locations: Location[];
}
