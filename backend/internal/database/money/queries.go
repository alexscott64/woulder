package money

const (
	queryProjectBySlug = `
		SELECT id, slug, name, center_lat, center_lon, default_zoom, created_at, updated_at
		FROM woulder.money_projects WHERE slug = $1
	`
	queryProjectByID = `
		SELECT id, slug, name, center_lat, center_lon, default_zoom, created_at, updated_at
		FROM woulder.money_projects WHERE id = $1
	`
	queryListFeaturesBase = `
		SELECT id, project_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, created_by, updated_by, created_at, updated_at
		FROM woulder.money_features
		WHERE project_id = $1
	`
	queryGetFeature = `
		SELECT id, project_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, created_by, updated_by, created_at, updated_at
		FROM woulder.money_features WHERE id = $1
	`
	queryCreateFeature = `
		INSERT INTO woulder.money_features
		(project_id, feature_type, title, description, status, geojson, style, properties, min_lat, min_lon, max_lat, max_lon, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, project_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, created_by, updated_by, created_at, updated_at
	`
	queryUpdateFeature = `
		UPDATE woulder.money_features SET
		  feature_type=$2, title=$3, description=$4, status=$5, geojson=$6, style=$7, properties=$8,
		  min_lat=$9, min_lon=$10, max_lat=$11, max_lon=$12, updated_by=$13, updated_at=now()
		WHERE id=$1
		RETURNING id, project_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, created_by, updated_by, created_at, updated_at
	`
	queryArchiveFeature = `UPDATE woulder.money_features SET status='archived', updated_by=$2, updated_at=now() WHERE id=$1`
	queryListNotes      = `
		SELECT id, project_id, feature_id, body, visibility, created_by, updated_by, created_at, updated_at
		FROM woulder.money_notes WHERE feature_id=$1 ORDER BY created_at DESC
	`
	queryCreateNote = `
		INSERT INTO woulder.money_notes (project_id, feature_id, body, visibility, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, project_id, feature_id, body, visibility, created_by, updated_by, created_at, updated_at
	`
	queryUpdateNote = `
		UPDATE woulder.money_notes SET body=$2, visibility=$3, updated_by=$4, updated_at=now()
		WHERE id=$1 AND (created_by=$4 OR $5='admin')
		RETURNING id, project_id, feature_id, body, visibility, created_by, updated_by, created_at, updated_at
	`
	queryDeleteNote   = `DELETE FROM woulder.money_notes WHERE id=$1 AND (created_by=$2 OR $3='admin')`
	queryCreateUpload = `
		INSERT INTO woulder.money_uploads
		(id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, uploaded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, uploaded_by, created_at
	`
	queryGetUpload = `
		SELECT id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, uploaded_by, created_at
		FROM woulder.money_uploads WHERE id=$1
	`
	queryListUploadsByFeature = `
		SELECT id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, uploaded_by, created_at
		FROM woulder.money_uploads WHERE feature_id=$1 ORDER BY created_at DESC
	`
	queryDeleteUpload      = `DELETE FROM woulder.money_uploads WHERE id=$1 AND (uploaded_by=$2 OR $3='admin')`
	queryFeatureNoteCounts = `SELECT feature_id, count(*) FROM woulder.money_notes WHERE project_id=$1 AND feature_id IS NOT NULL GROUP BY feature_id`
	queryPrimaryUploads    = `
		SELECT DISTINCT ON (feature_id) id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, uploaded_by, created_at
		FROM woulder.money_uploads WHERE project_id=$1 AND feature_id IS NOT NULL ORDER BY feature_id, created_at DESC
	`
)
