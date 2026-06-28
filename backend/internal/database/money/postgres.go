package money

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/database/dberrors"
	"github.com/alexscott64/woulder/backend/internal/models"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db DBConn
}

func NewPostgresRepository(db DBConn) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) GetProjectBySlug(ctx context.Context, slug string) (*models.MoneyProject, error) {
	return scanProject(r.db.QueryRowContext(ctx, queryProjectBySlug, slug))
}

func (r *PostgresRepository) GetProjectByID(ctx context.Context, id string) (*models.MoneyProject, error) {
	return scanProject(r.db.QueryRowContext(ctx, queryProjectByID, id))
}

func (r *PostgresRepository) ListFeatures(ctx context.Context, projectID string, filter models.MoneyFeatureFilter) ([]models.MoneyFeature, error) {
	query := queryListFeaturesBase
	args := []interface{}{projectID}
	if filter.FeatureType != "" {
		args = append(args, filter.FeatureType)
		query += fmt.Sprintf(" AND feature_type = $%d", len(args))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	} else if !filter.IncludeArchived {
		query += " AND status <> 'archived'"
	}
	if filter.BBox != nil {
		args = append(args, filter.BBox.MinLat, filter.BBox.MaxLat, filter.BBox.MinLon, filter.BBox.MaxLon)
		query += fmt.Sprintf(" AND max_lat >= $%d AND min_lat <= $%d AND max_lon >= $%d AND min_lon <= $%d", len(args)-3, len(args)-2, len(args)-1, len(args))
	}
	if filter.UpdatedAfter != nil {
		args = append(args, *filter.UpdatedAfter)
		query += fmt.Sprintf(" AND updated_at > $%d", len(args))
	}
	query += " ORDER BY sort_order ASC, title ASC, updated_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var features []models.MoneyFeature
	for rows.Next() {
		f, err := scanFeatureRows(rows)
		if err != nil {
			return nil, err
		}
		features = append(features, *f)
	}
	return features, rows.Err()
}

func (r *PostgresRepository) GetFeature(ctx context.Context, id string) (*models.MoneyFeature, error) {
	return scanFeature(r.db.QueryRowContext(ctx, queryGetFeature, id))
}

func (r *PostgresRepository) CreateFeature(ctx context.Context, f models.MoneyFeature) (*models.MoneyFeature, error) {
	return scanFeature(r.db.QueryRowContext(ctx, queryCreateFeature, f.ProjectID, f.ParentFeatureID, f.FeatureType, f.Title, f.Description, f.Status, f.GeoJSON, f.Style, f.Properties, f.MinLat, f.MinLon, f.MaxLat, f.MaxLon, f.SortOrder, f.ExternalRef, f.ImportSource, f.CreatedBy, f.UpdatedBy))
}

func (r *PostgresRepository) UpdateFeature(ctx context.Context, f models.MoneyFeature) (*models.MoneyFeature, error) {
	return scanFeature(r.db.QueryRowContext(ctx, queryUpdateFeature, f.ID, f.ParentFeatureID, f.FeatureType, f.Title, f.Description, f.Status, f.GeoJSON, f.Style, f.Properties, f.MinLat, f.MinLon, f.MaxLat, f.MaxLon, f.SortOrder, f.ExternalRef, f.ImportSource, f.UpdatedBy))
}

func (r *PostgresRepository) UpdateFeatureGeometry(ctx context.Context, id string, geojson []byte, bbox models.BBox, updatedBy string) (*models.MoneyFeature, error) {
	return scanFeature(r.db.QueryRowContext(ctx, queryUpdateFeatureGeometry, id, geojson, bbox.MinLat, bbox.MinLon, bbox.MaxLat, bbox.MaxLon, updatedBy))
}

func (r *PostgresRepository) UpsertFeatureByExternalRef(ctx context.Context, f models.MoneyFeature) (*models.MoneyFeature, error) {
	return scanFeature(r.db.QueryRowContext(ctx, queryUpsertFeatureByExternalRef, f.ProjectID, f.ParentFeatureID, f.FeatureType, f.Title, f.Description, f.Status, f.GeoJSON, f.Style, f.Properties, f.MinLat, f.MinLon, f.MaxLat, f.MaxLon, f.SortOrder, f.ExternalRef, f.ImportSource, f.CreatedBy, f.UpdatedBy))
}

func (r *PostgresRepository) ArchiveFeature(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx, queryArchiveFeature, id, userID)
	return err
}

func (r *PostgresRepository) PromoteChildrenAndArchiveFeature(ctx context.Context, id string, parentID *string, userID string) error {
	tx, err := beginTx(ctx, r.db)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, queryPromoteChildren, id, parentID, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, queryArchiveFeature, id, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresRepository) RestoreFeature(ctx context.Context, id, userID string) error {
	feature, err := r.GetFeature(ctx, id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, queryRestoreFeature, id, userID, defaultRestoredStatus(feature.FeatureType))
	return err
}

func (r *PostgresRepository) MoveFeatureParent(ctx context.Context, id string, parentID *string, sortOrder int, userID string) (*models.MoneyFeature, error) {
	return scanFeature(r.db.QueryRowContext(ctx, queryMoveFeatureParent, id, parentID, sortOrder, userID))
}

func (r *PostgresRepository) ListTrash(ctx context.Context, projectID string) ([]models.MoneyFeature, error) {
	rows, err := r.db.QueryContext(ctx, queryListTrash, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var features []models.MoneyFeature
	for rows.Next() {
		f, err := scanFeatureRows(rows)
		if err != nil {
			return nil, err
		}
		features = append(features, *f)
	}
	return features, rows.Err()
}

func (r *PostgresRepository) ListNotes(ctx context.Context, featureID string) ([]models.MoneyNote, error) {
	rows, err := queryWithLegacy(ctx, r.db, queryListNotes, queryListNotesLegacy, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotes(rows)
}

func (r *PostgresRepository) ListNotesByProject(ctx context.Context, projectID string) ([]models.MoneyNote, error) {
	rows, err := queryWithLegacy(ctx, r.db, queryListNotesByProject, queryListNotesByProjectLegacy, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotes(rows)
}

func (r *PostgresRepository) CreateNote(ctx context.Context, n models.MoneyNote) (*models.MoneyNote, error) {
	return scanNote(r.db.QueryRowContext(ctx, queryCreateNote, n.ProjectID, n.FeatureID, n.TargetType, n.TargetRef, n.Body, n.Visibility, pq.Array(n.Tags), n.Blocks, n.ExternalRef, n.ImportSource, n.CreatedBy, n.UpdatedBy))
}

func (r *PostgresRepository) UpdateNote(ctx context.Context, noteID, body, visibility string, tags []string, blocks []byte, userID, role string) (*models.MoneyNote, error) {
	return scanNote(r.db.QueryRowContext(ctx, queryUpdateNote, noteID, body, visibility, pq.Array(tags), blocks, userID, role))
}

func (r *PostgresRepository) DeleteNote(ctx context.Context, noteID, userID, role string) error {
	res, err := r.db.ExecContext(ctx, queryDeleteNote, noteID, userID, role)
	if err != nil {
		return err
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return dberrors.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateUpload(ctx context.Context, u models.MoneyUpload) (*models.MoneyUpload, error) {
	return scanUpload(r.db.QueryRowContext(ctx, queryCreateUpload, u.ID, u.ProjectID, u.FeatureID, u.NoteID, u.OriginalFilename, u.StorageKey, u.ContentType, u.ByteSize, u.Width, u.Height, u.ChecksumSHA256, u.BlockKind, u.Metadata, u.AssetKind, u.StorageBackend, u.StorageBucket, u.StorageRegion, u.StorageETag, u.StorageVersionID, u.Visibility, u.SyncStatus, u.UploadedBy))
}

func (r *PostgresRepository) GetUpload(ctx context.Context, id string) (*models.MoneyUpload, error) {
	return scanUpload(r.db.QueryRowContext(ctx, queryGetUpload, id))
}

func (r *PostgresRepository) ListUploadsByFeature(ctx context.Context, featureID string) ([]models.MoneyUpload, error) {
	rows, err := queryWithLegacy(ctx, r.db, queryListUploadsByFeature, queryListUploadsByFeatureLegacy, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUploads(rows)
}

func (r *PostgresRepository) ListUploadsByProject(ctx context.Context, projectID string) ([]models.MoneyUpload, error) {
	rows, err := queryWithLegacy(ctx, r.db, queryListUploadsByProject, queryListUploadsByProjectLegacy, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUploads(rows)
}

func (r *PostgresRepository) DeleteUpload(ctx context.Context, uploadID, userID, role string) error {
	var storageKey string
	err := r.db.QueryRowContext(ctx, querySoftDeleteUpload, uploadID, userID, role).Scan(&storageKey)
	return dberrors.WrapNotFound(err)
}

func (r *PostgresRepository) MarkUploadPhysicallyDeleted(ctx context.Context, uploadID string) error {
	_, err := r.db.ExecContext(ctx, queryMarkUploadPhysicallyDeleted, uploadID)
	return err
}

func (r *PostgresRepository) FeatureNoteCounts(ctx context.Context, projectID string) (map[string]int, error) {
	rows, err := queryWithLegacy(ctx, r.db, queryFeatureNoteCounts, queryFeatureNoteCountsLegacy, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var id string
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		out[id] = count
	}
	return out, rows.Err()
}

func (r *PostgresRepository) PrimaryUploads(ctx context.Context, projectID string) (map[string]models.MoneyUpload, error) {
	rows, err := r.db.QueryContext(ctx, queryPrimaryUploads, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]models.MoneyUpload{}
	for rows.Next() {
		u, err := scanUploadRows(rows)
		if err != nil {
			return nil, err
		}
		if u.FeatureID != nil {
			out[*u.FeatureID] = *u
		}
	}
	return out, rows.Err()
}

type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type queryRows interface {
	Next() bool
	Close() error
	Err() error
	Scan(dest ...interface{}) error
}

func queryWithLegacy(ctx context.Context, db DBConn, primary, legacy string, args ...interface{}) (queryRows, error) {
	rows, err := db.QueryContext(ctx, primary, args...)
	if !isUndefinedColumn(err) {
		return rows, err
	}
	return db.QueryContext(ctx, legacy, args...)
}

func isUndefinedColumn(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "42703"
}

func beginTx(ctx context.Context, db DBConn) (*sql.Tx, error) {
	beginner, ok := db.(txBeginner)
	if !ok {
		return nil, fmt.Errorf("money repository connection does not support transactions")
	}
	return beginner.BeginTx(ctx, nil)
}

type scanner interface {
	Scan(dest ...interface{}) error
}

type rowsScanner interface{ scanner }

func scanProject(row scanner) (*models.MoneyProject, error) {
	var p models.MoneyProject
	err := row.Scan(&p.ID, &p.Slug, &p.Name, &p.CenterLat, &p.CenterLon, &p.DefaultZoom, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}
	return &p, nil
}

func scanFeature(row scanner) (*models.MoneyFeature, error) {
	f, err := scanFeatureRows(row)
	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}
	return f, nil
}

func scanFeatureRows(row scanner) (*models.MoneyFeature, error) {
	var f models.MoneyFeature
	err := row.Scan(&f.ID, &f.ProjectID, &f.ParentFeatureID, &f.FeatureType, &f.Title, &f.Description, &f.Status, &f.GeoJSON, &f.Style, &f.Properties, &f.MinLat, &f.MinLon, &f.MaxLat, &f.MaxLon, &f.SortOrder, &f.ExternalRef, &f.ImportSource, &f.CreatedBy, &f.UpdatedBy, &f.CreatedAt, &f.UpdatedAt)
	return &f, err
}

func scanNotes(rows interface {
	Next() bool
	Err() error
	Scan(dest ...interface{}) error
}) ([]models.MoneyNote, error) {
	var notes []models.MoneyNote
	for rows.Next() {
		n, err := scanNoteRows(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, *n)
	}
	return notes, rows.Err()
}

func scanNote(row scanner) (*models.MoneyNote, error) {
	n, err := scanNoteRows(row)
	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}
	return n, nil
}
func scanNoteRows(row scanner) (*models.MoneyNote, error) {
	var n models.MoneyNote
	err := row.Scan(&n.ID, &n.ProjectID, &n.FeatureID, &n.TargetType, &n.TargetRef, &n.Body, &n.Visibility, pq.Array(&n.Tags), &n.Blocks, &n.ExternalRef, &n.ImportSource, &n.CreatedBy, &n.UpdatedBy, &n.CreatedAt, &n.UpdatedAt)
	return &n, err
}

func scanUploads(rows interface {
	Next() bool
	Err() error
	Scan(dest ...interface{}) error
}) ([]models.MoneyUpload, error) {
	var uploads []models.MoneyUpload
	for rows.Next() {
		u, err := scanUploadRows(rows)
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, *u)
	}
	return uploads, rows.Err()
}

func scanUpload(row scanner) (*models.MoneyUpload, error) {
	u, err := scanUploadRows(row)
	if err != nil {
		return nil, dberrors.WrapNotFound(err)
	}
	return u, nil
}
func scanUploadRows(row scanner) (*models.MoneyUpload, error) {
	var u models.MoneyUpload
	err := row.Scan(&u.ID, &u.ProjectID, &u.FeatureID, &u.NoteID, &u.OriginalFilename, &u.StorageKey, &u.ContentType, &u.ByteSize, &u.Width, &u.Height, &u.ChecksumSHA256, &u.BlockKind, &u.Metadata, &u.AssetKind, &u.StorageBackend, &u.StorageBucket, &u.StorageRegion, &u.StorageETag, &u.StorageVersionID, &u.Visibility, &u.SyncStatus, &u.DeletedAt, &u.DeletedBy, &u.DeleteRequestedAt, &u.PhysicallyDeletedAt, &u.UploadedBy, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func normalizeSpace(s string) string { return strings.TrimSpace(s) }

func defaultRestoredStatus(featureType string) string {
	switch featureType {
	case models.MoneyFeatureBoulder:
		return models.MoneyStatusScouted
	case models.MoneyFeatureProblem:
		return models.MoneyStatusProject
	default:
		return models.MoneyStatusActive
	}
}
