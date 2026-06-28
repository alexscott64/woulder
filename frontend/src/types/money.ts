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

export type MoneyFeatureType = 'area' | 'boulder' | 'problem' | 'trail' | 'topo' | 'poi' | 'drawing';
export type MoneyFeatureStatus = 'draft' | 'active' | 'archived' | 'scouted' | 'needs-work' | 'cleaning' | 'established' | 'project' | 'sent';
export type MoneyNoteVisibility = 'private' | 'team';
export type MoneyDevStatus = 'scouted' | 'needs-work' | 'cleaning' | 'established';
export type MoneyProblemStatus = 'project' | 'sent' | 'established';
export type MoneyNoteTargetType = 'project' | 'feature' | 'area' | 'boulder' | 'trail' | 'point' | 'none';
export type MoneyUploadBlockKind = 'photo' | 'sketch' | 'file' | 'topo';
export type MoneyArchiveMode = 'subtree' | 'promote_children';
export type MoneyTrailCategory = 'connector' | 'approach' | 'trail_to_area' | 'trail_to_destination';

export type MoneyPosition = [number, number];

export interface PointGeometry { type: 'Point'; coordinates: MoneyPosition; }
export interface LineStringGeometry { type: 'LineString'; coordinates: MoneyPosition[]; }
export interface PolygonGeometry { type: 'Polygon'; coordinates: MoneyPosition[][]; }
export interface FeatureGeometry { type: 'Feature'; geometry: MoneyGeometry; properties?: Record<string, unknown>; }
export interface FeatureCollectionGeometry { type: 'FeatureCollection'; features: FeatureGeometry[]; }

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
  parent_feature_id?: string;
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
  sort_order?: number;
  external_ref?: string;
  import_source?: string;
  created_by: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
}

export interface MoneyNoteBlock {
  kind: MoneyUploadBlockKind;
  upload_id?: string;
  url?: string;
  name?: string;
  metadata?: Record<string, unknown>;
}

export interface MoneyNote {
  id: string;
  project_id: string;
  feature_id?: string;
  target_type?: MoneyNoteTargetType;
  target_ref?: string;
  body: string;
  visibility: MoneyNoteVisibility;
  tags?: string[];
  blocks?: MoneyNoteBlock[];
  external_ref?: string;
  import_source?: string;
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
  block_kind: MoneyUploadBlockKind;
  metadata: Record<string, unknown>;
  asset_kind: string;
  storage_backend: 'local' | 'r2';
  storage_bucket?: string;
  storage_region?: string;
  visibility: 'private';
  sync_status: 'available' | 'pending_upload' | 'deleted';
  uploaded_by: string;
  created_at: string;
}

export interface MoneyUploadDownloadURL {
  url: string;
  expires_at: string;
  proxy_url: string;
}

export interface MoneyPermissions { can_read: boolean; can_write: boolean; is_admin: boolean; }
export interface MoneyProjectResponse { project: MoneyProject; user: MoneyCurrentUser; permissions: MoneyPermissions; }
export interface MoneySnapshot { project: MoneyProject; features: MoneyFeature[]; note_counts: Record<string, number>; primary_uploads: Record<string, MoneyUpload>; }
export interface MoneyCragNode { feature: MoneyFeature; children?: MoneyCragNode[] | null; boulders?: MoneyCragNode[] | null; problems?: MoneyCragNode[] | null; }
export interface MoneyCragSnapshot { project: MoneyProject; root: MoneyCragNode | null; trails?: MoneyCragNode[] | null; notes?: MoneyNote[] | null; uploads?: MoneyUpload[] | null; }
export interface MoneyFeatureDetail { feature: MoneyFeature; notes: MoneyNote[] | null; uploads: MoneyUpload[] | null; }
export interface MoneyTrashItem { id: string; title: string; feature_type: MoneyFeatureType; parent_feature_id?: string; path: string[]; deleted_at: string; updated_at: string; descendant_count: number; }
export interface MoneyTrashResponse { items: MoneyTrashItem[]; }

export interface MoneyFeatureRequest {
  parent_feature_id?: string | null;
  feature_type: MoneyFeatureType;
  title: string;
  description?: string | null;
  status: MoneyFeatureStatus;
  geojson: MoneyGeoJSON;
  style: Record<string, unknown>;
  properties: Record<string, unknown>;
  sort_order?: number;
}

export interface MoneyAreaRequest { parent_feature_id?: string | null; title: string; description?: string | null; geojson: MoneyGeoJSON; properties: Record<string, unknown>; }
export interface MoneyAreaGeometryRequest { geojson: MoneyGeoJSON; }
export interface MoneyMoveFeatureRequest { parent_feature_id: string | null; sort_order?: number; }
export interface MoneyBoulderRequest { parent_feature_id: string; title: string; description?: string | null; dev_status: MoneyDevStatus; geojson: MoneyGeoJSON; properties: Record<string, unknown>; }
export interface MoneyProblemRequest { boulder_id: string; name: string; grade: string; status: MoneyProblemStatus; stars: number; fa?: string | null; types: string[]; description?: string | null; properties?: Record<string, unknown>; }
export interface MoneyBoulderStatusRequest { dev_status: MoneyDevStatus; }

export interface MoneyNoteRequest {
  body: string;
  visibility: MoneyNoteVisibility;
  target_type?: MoneyNoteTargetType;
  target_ref?: string | null;
  tags?: string[];
  blocks?: MoneyNoteBlock[];
}

export interface MoneyFeatureFilters { type?: MoneyFeatureType | 'all'; status?: MoneyFeatureStatus | 'all'; search?: string; }
export interface MoneyBBox { minLon: number; minLat: number; maxLon: number; maxLat: number; }
