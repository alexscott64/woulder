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
	queryFeatureSelect = `
		SELECT id, project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by, created_at, updated_at
		FROM woulder.money_features
	`
	queryListFeaturesBase = queryFeatureSelect + ` WHERE project_id = $1`
	queryGetFeature       = queryFeatureSelect + ` WHERE id = $1`
	queryCreateFeature    = `
		INSERT INTO woulder.money_features
		(project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties, min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING id, project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by, created_at, updated_at
	`
	queryUpdateFeature = `
		UPDATE woulder.money_features SET
		  parent_feature_id=$2, feature_type=$3, title=$4, description=$5, status=$6, geojson=$7, style=$8, properties=$9,
		  min_lat=$10, min_lon=$11, max_lat=$12, max_lon=$13, sort_order=$14, external_ref=$15, import_source=$16, updated_by=$17, updated_at=now()
		WHERE id=$1
		RETURNING id, project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by, created_at, updated_at
	`
	queryUpdateFeatureGeometry = `
		UPDATE woulder.money_features SET
		  geojson=$2, min_lat=$3, min_lon=$4, max_lat=$5, max_lon=$6, updated_by=$7, updated_at=now()
		WHERE id=$1
		RETURNING id, project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by, created_at, updated_at
	`
	queryUpsertFeatureByExternalRef = `
		INSERT INTO woulder.money_features
		(project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties, min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		ON CONFLICT (project_id, external_ref) WHERE external_ref IS NOT NULL DO UPDATE SET
		  parent_feature_id=EXCLUDED.parent_feature_id,
		  feature_type=EXCLUDED.feature_type,
		  title=EXCLUDED.title,
		  description=EXCLUDED.description,
		  status=EXCLUDED.status,
		  geojson=EXCLUDED.geojson,
		  style=EXCLUDED.style,
		  properties=EXCLUDED.properties,
		  min_lat=EXCLUDED.min_lat,
		  min_lon=EXCLUDED.min_lon,
		  max_lat=EXCLUDED.max_lat,
		  max_lon=EXCLUDED.max_lon,
		  sort_order=EXCLUDED.sort_order,
		  import_source=EXCLUDED.import_source,
		  updated_by=EXCLUDED.updated_by,
		  updated_at=now()
		RETURNING id, project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by, created_at, updated_at
	`
	queryArchiveFeature  = `UPDATE woulder.money_features SET status='archived', updated_by=$2, updated_at=now() WHERE id=$1`
	queryPromoteChildren = `
		UPDATE woulder.money_features SET parent_feature_id=$2, updated_by=$3, updated_at=now()
		WHERE parent_feature_id=$1 AND status <> 'archived'
	`
	queryRestoreFeature    = `UPDATE woulder.money_features SET status=$3, updated_by=$2, updated_at=now() WHERE id=$1`
	queryMoveFeatureParent = `
		UPDATE woulder.money_features SET parent_feature_id=$2, sort_order=$3, updated_by=$4, updated_at=now()
		WHERE id=$1
		RETURNING id, project_id, parent_feature_id, feature_type, title, description, status, geojson, style, properties,
		       min_lat, min_lon, max_lat, max_lon, sort_order, external_ref, import_source, created_by, updated_by, created_at, updated_at
	`
	queryListTrash  = queryFeatureSelect + ` WHERE project_id=$1 AND status='archived' ORDER BY updated_at DESC, title ASC`
	queryNoteSelect = `
		SELECT id, project_id, feature_id, target_type, target_ref, body, visibility, tags, blocks, external_ref, import_source, created_by, updated_by, created_at, updated_at
		FROM woulder.money_notes
	`
	queryListNotes                = queryNoteSelect + ` WHERE deleted_at IS NULL AND (feature_id=$1 OR target_ref=$1) ORDER BY created_at DESC`
	queryListNotesByProject       = queryNoteSelect + ` WHERE project_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryListNotesLegacy          = queryNoteSelect + ` WHERE feature_id=$1 OR target_ref=$1 ORDER BY created_at DESC`
	queryListNotesByProjectLegacy = queryNoteSelect + ` WHERE project_id=$1 ORDER BY created_at DESC`
	queryCreateNote               = `
		INSERT INTO woulder.money_notes (project_id, feature_id, target_type, target_ref, body, visibility, tags, blocks, external_ref, import_source, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (project_id, external_ref) WHERE external_ref IS NOT NULL DO UPDATE SET
		  feature_id=EXCLUDED.feature_id,
		  target_type=EXCLUDED.target_type,
		  target_ref=EXCLUDED.target_ref,
		  body=EXCLUDED.body,
		  visibility=EXCLUDED.visibility,
		  tags=EXCLUDED.tags,
		  blocks=EXCLUDED.blocks,
		  import_source=EXCLUDED.import_source,
		  updated_by=EXCLUDED.updated_by,
		  updated_at=now()
		RETURNING id, project_id, feature_id, target_type, target_ref, body, visibility, tags, blocks, external_ref, import_source, created_by, updated_by, created_at, updated_at
	`
	queryUpdateNote = `
		UPDATE woulder.money_notes SET body=$2, visibility=$3, tags=$4, blocks=$5, updated_by=$6, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL AND (created_by=$6 OR $7='admin')
		RETURNING id, project_id, feature_id, target_type, target_ref, body, visibility, tags, blocks, external_ref, import_source, created_by, updated_by, created_at, updated_at
	`
	queryDeleteNote = `
		UPDATE woulder.money_notes SET deleted_at=COALESCE(deleted_at, now()), deleted_by=COALESCE(deleted_by, $2), updated_by=$2, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL AND (created_by=$2 OR $3='admin')
	`
	queryCreateUpload = `
		INSERT INTO woulder.money_uploads
		(id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, block_kind, metadata, asset_kind, storage_backend, storage_bucket, storage_region, storage_etag, storage_version_id, visibility, sync_status, uploaded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, block_kind, metadata, asset_kind, storage_backend, storage_bucket, storage_region, storage_etag, storage_version_id, visibility, sync_status, deleted_at, deleted_by, delete_requested_at, physically_deleted_at, uploaded_by, created_at, updated_at
	`
	queryUploadSelect = `
		SELECT id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, block_kind, metadata, asset_kind, storage_backend, storage_bucket, storage_region, storage_etag, storage_version_id, visibility, sync_status, deleted_at, deleted_by, delete_requested_at, physically_deleted_at, uploaded_by, created_at, updated_at
		FROM woulder.money_uploads
	`
	queryUploadSelectLegacy = `
		SELECT id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256,
		       'photo' AS block_kind, '{}'::jsonb AS metadata, 'original' AS asset_kind, 'local' AS storage_backend, NULL::text AS storage_bucket,
		       NULL::text AS storage_region, NULL::text AS storage_etag, NULL::text AS storage_version_id, 'private' AS visibility,
		       'available' AS sync_status, NULL::timestamptz AS deleted_at, NULL::uuid AS deleted_by, NULL::timestamptz AS delete_requested_at,
		       NULL::timestamptz AS physically_deleted_at, uploaded_by, created_at, created_at AS updated_at
		FROM woulder.money_uploads
	`
	queryGetUpload                  = queryUploadSelect + ` WHERE id=$1 AND deleted_at IS NULL`
	queryListUploadsByFeature       = queryUploadSelect + ` WHERE feature_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryListUploadsByProject       = queryUploadSelect + ` WHERE project_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`
	queryListUploadsByFeatureLegacy = queryUploadSelectLegacy + ` WHERE feature_id=$1 ORDER BY created_at DESC`
	queryListUploadsByProjectLegacy = queryUploadSelectLegacy + ` WHERE project_id=$1 ORDER BY created_at DESC`
	querySoftDeleteUpload           = `
		UPDATE woulder.money_uploads
		SET deleted_at=COALESCE(deleted_at, now()), deleted_by=COALESCE(deleted_by, $2), delete_requested_at=COALESCE(delete_requested_at, now()), sync_status='deleted', updated_at=now()
		WHERE id=$1 AND (uploaded_by=$2 OR $3='admin' OR deleted_at IS NOT NULL)
		RETURNING storage_key
	`
	queryMarkUploadPhysicallyDeleted = `UPDATE woulder.money_uploads SET physically_deleted_at=COALESCE(physically_deleted_at, now()), updated_at=now() WHERE id=$1`
	queryFeatureNoteCounts           = `SELECT COALESCE(feature_id::text, target_ref), count(*) FROM woulder.money_notes WHERE project_id=$1 AND deleted_at IS NULL AND COALESCE(feature_id::text, target_ref) IS NOT NULL GROUP BY COALESCE(feature_id::text, target_ref)`
	queryFeatureNoteCountsLegacy     = `SELECT COALESCE(feature_id::text, target_ref), count(*) FROM woulder.money_notes WHERE project_id=$1 AND COALESCE(feature_id::text, target_ref) IS NOT NULL GROUP BY COALESCE(feature_id::text, target_ref)`
	queryPrimaryUploads              = `
		SELECT DISTINCT ON (feature_id) id, project_id, feature_id, note_id, original_filename, storage_key, content_type, byte_size, width, height, checksum_sha256, block_kind, metadata, asset_kind, storage_backend, storage_bucket, storage_region, storage_etag, storage_version_id, visibility, sync_status, deleted_at, deleted_by, delete_requested_at, physically_deleted_at, uploaded_by, created_at, updated_at
		FROM woulder.money_uploads WHERE project_id=$1 AND feature_id IS NOT NULL AND deleted_at IS NULL ORDER BY feature_id, created_at DESC
	`
)
