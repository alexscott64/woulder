-- Migration 000031: Add analytics tables for site visitor tracking
-- Tracks sessions, page views, and user interactions for the /iglooghost CMS dashboard

SET search_path TO woulder, public;

-- ============================================================================
-- TABLES
-- ============================================================================

-- Analytics sessions - tracks unique visitor sessions
CREATE TABLE IF NOT EXISTS woulder.analytics_sessions (
    id SERIAL PRIMARY KEY,
    session_id UUID NOT NULL UNIQUE,
    visitor_id VARCHAR(64) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    country VARCHAR(100),
    region VARCHAR(100),
    city VARCHAR(100),
    device_type VARCHAR(20) DEFAULT 'desktop',
    browser VARCHAR(50),
    os VARCHAR(50),
    screen_width INTEGER,
    screen_height INTEGER,
    started_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    page_count INTEGER DEFAULT 0,
    duration_seconds INTEGER DEFAULT 0,
    is_bounce BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Analytics events - tracks all user interactions
CREATE TABLE IF NOT EXISTS woulder.analytics_events (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES woulder.analytics_sessions(session_id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_name VARCHAR(100) NOT NULL,
    page_path VARCHAR(500),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Analytics admin users - simple auth for dashboard
CREATE TABLE IF NOT EXISTS woulder.analytics_admin_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMPTZ
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Session lookups
CREATE INDEX IF NOT EXISTS idx_analytics_sessions_visitor ON woulder.analytics_sessions(visitor_id);
CREATE INDEX IF NOT EXISTS idx_analytics_sessions_started ON woulder.analytics_sessions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_sessions_ip ON woulder.analytics_sessions(ip_address);
CREATE INDEX IF NOT EXISTS idx_analytics_sessions_last_active ON woulder.analytics_sessions(last_active_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_sessions_device ON woulder.analytics_sessions(device_type);

-- Event lookups
CREATE INDEX IF NOT EXISTS idx_analytics_events_session ON woulder.analytics_events(session_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON woulder.analytics_events(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_name ON woulder.analytics_events(event_name, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_created ON woulder.analytics_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_metadata ON woulder.analytics_events USING GIN(metadata);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE woulder.analytics_sessions IS 'Visitor session tracking for the /iglooghost analytics dashboard';
COMMENT ON TABLE woulder.analytics_events IS 'User interaction events (page views, clicks, modal opens, etc.)';
COMMENT ON TABLE woulder.analytics_admin_users IS 'Admin credentials for the /iglooghost analytics dashboard';
COMMENT ON COLUMN woulder.analytics_sessions.visitor_id IS 'Privacy-friendly fingerprint hash (screen+timezone+language)';
COMMENT ON COLUMN woulder.analytics_sessions.is_bounce IS 'True if session had only one page view';
COMMENT ON COLUMN woulder.analytics_events.metadata IS 'Flexible JSONB for event-specific data (location_id, area_id, etc.)';
