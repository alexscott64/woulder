-- Migration 000038: Add general app auth and Money Creek toolkit tables

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS woulder.users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'developer' CHECK (role IN ('admin','developer','viewer')),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_login_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS woulder.auth_refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES woulder.users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS woulder.money_projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  center_lat DOUBLE PRECISION NOT NULL,
  center_lon DOUBLE PRECISION NOT NULL,
  default_zoom DOUBLE PRECISION NOT NULL DEFAULT 14,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS woulder.money_features (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES woulder.money_projects(id) ON DELETE CASCADE,
  feature_type TEXT NOT NULL CHECK (feature_type IN ('trail','topo','poi','drawing')),
  title TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','archived')),
  geojson JSONB NOT NULL,
  style JSONB NOT NULL DEFAULT '{}',
  properties JSONB NOT NULL DEFAULT '{}',
  min_lat DOUBLE PRECISION,
  min_lon DOUBLE PRECISION,
  max_lat DOUBLE PRECISION,
  max_lon DOUBLE PRECISION,
  created_by UUID NOT NULL REFERENCES woulder.users(id),
  updated_by UUID NOT NULL REFERENCES woulder.users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS woulder.money_notes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES woulder.money_projects(id) ON DELETE CASCADE,
  feature_id UUID REFERENCES woulder.money_features(id) ON DELETE SET NULL,
  body TEXT NOT NULL,
  visibility TEXT NOT NULL DEFAULT 'team' CHECK (visibility IN ('private','team')),
  created_by UUID NOT NULL REFERENCES woulder.users(id),
  updated_by UUID NOT NULL REFERENCES woulder.users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS woulder.money_uploads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES woulder.money_projects(id) ON DELETE CASCADE,
  feature_id UUID REFERENCES woulder.money_features(id) ON DELETE SET NULL,
  note_id UUID REFERENCES woulder.money_notes(id) ON DELETE SET NULL,
  original_filename TEXT NOT NULL,
  storage_key TEXT NOT NULL UNIQUE,
  content_type TEXT NOT NULL,
  byte_size BIGINT NOT NULL,
  width INTEGER,
  height INTEGER,
  checksum_sha256 TEXT NOT NULL,
  uploaded_by UUID NOT NULL REFERENCES woulder.users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_user ON woulder.auth_refresh_tokens(user_id, expires_at DESC);
CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_active ON woulder.auth_refresh_tokens(token_hash) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_money_features_project_type_status_updated ON woulder.money_features(project_id, feature_type, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_money_features_viewport ON woulder.money_features(project_id, min_lat, max_lat, min_lon, max_lon);
CREATE INDEX IF NOT EXISTS idx_money_features_geojson_gin ON woulder.money_features USING GIN (geojson);
CREATE INDEX IF NOT EXISTS idx_money_features_properties_gin ON woulder.money_features USING GIN (properties);
CREATE INDEX IF NOT EXISTS idx_money_notes_project_feature_created ON woulder.money_notes(project_id, feature_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_money_uploads_project_feature_created ON woulder.money_uploads(project_id, feature_id, created_at DESC);

INSERT INTO woulder.money_projects (slug, name, center_lat, center_lon, default_zoom)
VALUES ('money-creek', 'Money Creek', 47.7119, -121.5208, 14)
ON CONFLICT (slug) DO UPDATE SET
  name = EXCLUDED.name,
  center_lat = EXCLUDED.center_lat,
  center_lon = EXCLUDED.center_lon,
  default_zoom = EXCLUDED.default_zoom,
  updated_at = now();
