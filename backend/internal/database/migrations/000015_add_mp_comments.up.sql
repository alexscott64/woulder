-- Add table for Mountain Project comments on both areas and routes
-- Uses a single table with discriminator for better normalization

CREATE TABLE woulder.mp_comments (
    id SERIAL PRIMARY KEY,
    mp_comment_id VARCHAR(50) NOT NULL UNIQUE,
    comment_type VARCHAR(10) NOT NULL CHECK (comment_type IN ('area', 'route')),
    mp_area_id VARCHAR(50),
    mp_route_id VARCHAR(50),
    user_name VARCHAR(255) NOT NULL,
    user_id VARCHAR(50),
    comment_text TEXT NOT NULL,
    commented_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Ensure either area_id or route_id is set, but not both
    CONSTRAINT check_comment_target CHECK (
        (comment_type = 'area' AND mp_area_id IS NOT NULL AND mp_route_id IS NULL) OR
        (comment_type = 'route' AND mp_route_id IS NOT NULL AND mp_area_id IS NULL)
    )
);

-- Indexes for efficient queries
CREATE INDEX idx_mp_comments_area ON woulder.mp_comments(mp_area_id) WHERE comment_type = 'area';
CREATE INDEX idx_mp_comments_route ON woulder.mp_comments(mp_route_id) WHERE comment_type = 'route';
CREATE INDEX idx_mp_comments_commented_at ON woulder.mp_comments(commented_at DESC);
CREATE INDEX idx_mp_comments_type ON woulder.mp_comments(comment_type);

-- Auto-update trigger for updated_at
CREATE TRIGGER update_mp_comments_updated_at BEFORE UPDATE ON woulder.mp_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comment to the migration tracking
COMMENT ON TABLE woulder.mp_comments IS 'Comments on Mountain Project areas and routes';
COMMENT ON COLUMN woulder.mp_comments.mp_comment_id IS 'Mountain Project comment ID (unique across all comments)';
COMMENT ON COLUMN woulder.mp_comments.comment_type IS 'Type of comment: area or route';
COMMENT ON COLUMN woulder.mp_comments.mp_area_id IS 'Mountain Project area ID (for area comments)';
COMMENT ON COLUMN woulder.mp_comments.mp_route_id IS 'Mountain Project route ID (for route comments)';
COMMENT ON COLUMN woulder.mp_comments.commented_at IS 'When the comment was originally posted on Mountain Project';
