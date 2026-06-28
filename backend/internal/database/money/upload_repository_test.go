package money

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

var testTime = time.Unix(1700000000, 0)

func TestPostgresRepositoryListNotesByProjectFallsBackWhenSoftDeleteMigrationMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta(queryListNotesByProject)).
		WithArgs("project-1").
		WillReturnError(&pq.Error{Code: "42703", Message: "column deleted_at does not exist"})
	mock.ExpectQuery(regexp.QuoteMeta(queryListNotesByProjectLegacy)).
		WithArgs("project-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "feature_id", "target_type", "target_ref", "body", "visibility", "tags", "blocks", "external_ref", "import_source", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow("note-1", "project-1", "feature-1", "boulder", nil, "body", "team", pq.Array([]string{"tag"}), []byte(`[]`), nil, nil, "user-1", "user-1", testTime, testTime))

	notes, err := repo.ListNotesByProject(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListNotesByProject returned error: %v", err)
	}
	if len(notes) != 1 || notes[0].ID != "note-1" {
		t.Fatalf("unexpected notes: %+v", notes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresRepositoryListUploadsByProjectFallsBackWhenAssetStorageMigrationMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta(queryListUploadsByProject)).
		WithArgs("project-1").
		WillReturnError(&pq.Error{Code: "42703", Message: "column deleted_at does not exist"})
	mock.ExpectQuery(regexp.QuoteMeta(queryListUploadsByProjectLegacy)).
		WithArgs("project-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "feature_id", "note_id", "original_filename", "title", "comments", "storage_key", "content_type", "byte_size", "width", "height", "checksum_sha256", "block_kind", "metadata", "asset_kind", "storage_backend", "storage_bucket", "storage_region", "storage_etag", "storage_version_id", "visibility", "sync_status", "deleted_at", "deleted_by", "delete_requested_at", "physically_deleted_at", "uploaded_by", "created_at", "updated_at"}).
			AddRow("upload-1", "project-1", "feature-1", "note-1", "photo.jpg", nil, nil, "money-creek/assets/upload-1/original/photo.jpg", "image/jpeg", int64(42), nil, nil, "checksum", "photo", []byte(`{}`), "original", "local", nil, nil, nil, nil, "private", "available", nil, nil, nil, nil, "user-1", testTime, testTime))

	uploads, err := repo.ListUploadsByProject(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListUploadsByProject returned error: %v", err)
	}
	if len(uploads) != 1 || uploads[0].ID != "upload-1" {
		t.Fatalf("unexpected uploads: %+v", uploads)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresRepositoryDeleteUploadSoftDeletes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta(querySoftDeleteUpload)).
		WithArgs("upload-1", "user-1", "developer").
		WillReturnRows(sqlmock.NewRows([]string{"storage_key"}).AddRow("money-creek/assets/upload-1/original/photo.jpg"))

	if err := repo.DeleteUpload(context.Background(), "upload-1", "user-1", "developer"); err != nil {
		t.Fatalf("DeleteUpload returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresRepositoryUpdateUploadMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	title := "Topo overview"
	comments := "Shows the main face."
	mock.ExpectQuery(regexp.QuoteMeta(queryUpdateUploadMetadata)).
		WithArgs("upload-1", &title, &comments, "user-1", "developer").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "feature_id", "note_id", "original_filename", "title", "comments", "storage_key", "content_type", "byte_size", "width", "height", "checksum_sha256", "block_kind", "metadata", "asset_kind", "storage_backend", "storage_bucket", "storage_region", "storage_etag", "storage_version_id", "visibility", "sync_status", "deleted_at", "deleted_by", "delete_requested_at", "physically_deleted_at", "uploaded_by", "created_at", "updated_at"}).
			AddRow("upload-1", "project-1", "feature-1", nil, "photo.jpg", title, comments, "money-creek/assets/upload-1/original/photo.jpg", "image/jpeg", int64(42), nil, nil, "checksum", "photo", []byte(`{}`), "original", "r2", nil, nil, nil, nil, "private", "available", nil, nil, nil, nil, "user-1", testTime, testTime))

	upload, err := repo.UpdateUploadMetadata(context.Background(), "upload-1", &title, &comments, "user-1", "developer")
	if err != nil {
		t.Fatalf("UpdateUploadMetadata returned error: %v", err)
	}
	if upload.Title == nil || *upload.Title != title || upload.Comments == nil || *upload.Comments != comments {
		t.Fatalf("unexpected upload metadata: %+v", upload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresRepositoryMarkUploadPhysicallyDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	mock.ExpectExec(regexp.QuoteMeta(queryMarkUploadPhysicallyDeleted)).
		WithArgs("upload-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.MarkUploadPhysicallyDeleted(context.Background(), "upload-1"); err != nil {
		t.Fatalf("MarkUploadPhysicallyDeleted returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresRepositoryListNotesByProjectLatestSchemaFiltersSoftDeleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta(queryListNotesByProject)).
		WithArgs("project-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "feature_id", "target_type", "target_ref", "body", "visibility", "tags", "blocks", "external_ref", "import_source", "created_by", "updated_by", "created_at", "updated_at"}).
			AddRow("note-1", "project-1", "feature-1", "boulder", nil, "body", "team", pq.Array([]string{"tag"}), []byte(`[{"kind":"photo","upload_id":"upload-1"}]`), nil, nil, "user-1", "user-1", testTime, testTime))

	notes, err := repo.ListNotesByProject(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListNotesByProject returned error: %v", err)
	}
	if len(notes) != 1 || notes[0].ID != "note-1" || string(notes[0].Blocks) == "" {
		t.Fatalf("unexpected notes: %+v", notes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresRepositoryListUploadsByProjectLatestSchemaScansStorageAndDeleteMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	repo := NewPostgresRepository(db)
	bucket := "money-assets"
	region := "auto"
	etag := "etag-1"
	deletedAt := testTime.Add(time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta(queryListUploadsByProject)).
		WithArgs("project-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "feature_id", "note_id", "original_filename", "title", "comments", "storage_key", "content_type", "byte_size", "width", "height", "checksum_sha256", "block_kind", "metadata", "asset_kind", "storage_backend", "storage_bucket", "storage_region", "storage_etag", "storage_version_id", "visibility", "sync_status", "deleted_at", "deleted_by", "delete_requested_at", "physically_deleted_at", "uploaded_by", "created_at", "updated_at"}).
			AddRow("upload-1", "project-1", "feature-1", "note-1", "photo.jpg", "Topo overview", "Shows the main face", "money-creek/assets/upload-1/original/photo.jpg", "image/jpeg", int64(42), nil, nil, "checksum", "photo", []byte(`{"caption":"topo"}`), "original", "r2", bucket, region, etag, nil, "private", "available", nil, nil, nil, nil, "user-1", testTime, testTime).
			AddRow("deleted-upload", "project-1", "feature-1", "note-1", "deleted.jpg", nil, nil, "money-creek/assets/deleted/original/deleted.jpg", "image/jpeg", int64(1), nil, nil, "checksum", "photo", []byte(`{}`), "original", "r2", bucket, region, etag, nil, "private", "deleted", deletedAt, "user-1", deletedAt, nil, "user-1", testTime, deletedAt))

	uploads, err := repo.ListUploadsByProject(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListUploadsByProject returned error: %v", err)
	}
	if len(uploads) != 2 || uploads[0].StorageBackend != "r2" || uploads[0].StorageBucket == nil || *uploads[0].StorageBucket != bucket {
		t.Fatalf("unexpected uploads: %+v", uploads)
	}
	if uploads[1].DeletedAt == nil || uploads[1].SyncStatus != "deleted" {
		t.Fatalf("expected deleted upload metadata to scan, got %+v", uploads[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
