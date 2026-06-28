package money

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

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
