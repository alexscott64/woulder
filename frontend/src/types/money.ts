export type MoneyUserRole = 'admin' | 'developer' | 'viewer';

export interface MoneyCurrentUser {
  id: string;
  email: string;
  display_name: string;
  role: MoneyUserRole;
}

export interface AuthResponse {
  user: MoneyCurrentUser;
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface AuthMeResponse {
  user: MoneyCurrentUser;
}

export type MoneyFeatureType = 'trail' | 'topo' | 'poi' | 'drawing';
export type MoneyFeatureStatus = 'draft' | 'active' | 'archived';
export type MoneyNoteVisibility = 'private' | 'team';

export type MoneyPosition = [number, number];

export interface PointGeometry {
  type: 'Point';
  coordinates: MoneyPosition;
}

export interface LineStringGeometry {
  type: 'LineString';
  coordinates: MoneyPosition[];
}

export interface PolygonGeometry {
  type: 'Polygon';
  coordinates: MoneyPosition[][];
}

export interface FeatureGeometry {
  type: 'Feature';
  geometry: MoneyGeometry;
  properties?: Record<string, unknown>;
}

export interface FeatureCollectionGeometry {
  type: 'FeatureCollection';
  features: FeatureGeometry[];
}

export type MoneyGeometry = PointGeometry | LineStringGeometry | PolygonGeometry;
export type MoneyGeoJSON = MoneyGeometry | FeatureGeometry | FeatureCollectionGeometry;

export interface MoneyProject {
  id: string;
  slug: string;
  name: string;
  center_lat: number;
  center_lon: number;
  default_zoom: number;
  created_at: string;
  updated_at: string;
}

export interface MoneyFeature {
  id: string;
  project_id: string;
  feature_type: MoneyFeatureType;
  title: string;
  description?: string;
  status: MoneyFeatureStatus;
  geojson: MoneyGeoJSON;
  style: Record<string, unknown>;
  properties: Record<string, unknown>;
  min_lat?: number;
  min_lon?: number;
  max_lat?: number;
  max_lon?: number;
  created_by: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
}

export interface MoneyNote {
  id: string;
  project_id: string;
  feature_id?: string;
  body: string;
  visibility: MoneyNoteVisibility;
  created_by: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
}

export interface MoneyUpload {
  id: string;
  project_id: string;
  feature_id?: string;
  note_id?: string;
  original_filename: string;
  content_type: string;
  byte_size: number;
  width?: number;
  height?: number;
  checksum_sha256: string;
  uploaded_by: string;
  created_at: string;
}

export interface MoneyPermissions {
  can_read: boolean;
  can_write: boolean;
  is_admin: boolean;
}

export interface MoneyProjectResponse {
  project: MoneyProject;
  user: MoneyCurrentUser;
  permissions: MoneyPermissions;
}

export interface MoneySnapshot {
  project: MoneyProject;
  features: MoneyFeature[];
  note_counts: Record<string, number>;
  primary_uploads: Record<string, MoneyUpload>;
}

export interface MoneyFeatureDetail {
  feature: MoneyFeature;
  notes: MoneyNote[];
  uploads: MoneyUpload[];
}

export interface MoneyFeatureRequest {
  feature_type: MoneyFeatureType;
  title: string;
  description?: string | null;
  status: MoneyFeatureStatus;
  geojson: MoneyGeoJSON;
  style: Record<string, unknown>;
  properties: Record<string, unknown>;
}

export interface MoneyNoteRequest {
  body: string;
  visibility: MoneyNoteVisibility;
}

export interface MoneyFeatureFilters {
  type?: MoneyFeatureType | 'all';
  status?: MoneyFeatureStatus | 'all';
  search?: string;
}

export interface MoneyBBox {
  minLon: number;
  minLat: number;
  maxLon: number;
  maxLat: number;
}
