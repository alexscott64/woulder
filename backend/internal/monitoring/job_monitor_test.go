package monitoring

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestSaveCheckpoint_SingleJsonbSetUpdate verifies the WAL-saving shape of
// SaveCheckpoint: a SINGLE UPDATE using jsonb_set, no SELECT round-trip,
// and the checkpoint payload passed as a discrete JSONB parameter (not
// merged into a full-blob rewrite).
//
// This shape is the whole point of the change documented in
// SaveCheckpoint's doc comment: the previous read-modify-write of the
// entire metadata JSONB was rewriting hundreds of KB per call and was the
// single largest source of RDS WAL volume from job_executions.
func TestSaveCheckpoint_SingleJsonbSetUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	monitor := NewJobMonitor(db)

	checkpoint := map[string]interface{}{
		"current_state_index": 42,
		"total_new_routes":    7,
	}
	wantCheckpointJSON, err := json.Marshal(checkpoint)
	if err != nil {
		t.Fatalf("marshal checkpoint: %v", err)
	}

	// Expect exactly one UPDATE on job_executions that uses jsonb_set and
	// targets both the checkpoint and last_checkpoint_time keys. The
	// regex deliberately does NOT match a SELECT, so if SaveCheckpoint
	// regresses to a read-modify-write the test will fail.
	mock.ExpectExec(regexp.QuoteMeta("UPDATE woulder.job_executions")).
		WithArgs(
			wantCheckpointJSON,
			sqlmock.AnyArg(), // last_checkpoint_time (RFC3339 string)
			int64(123),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := monitor.SaveCheckpoint(context.Background(), 123, checkpoint); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestSaveCheckpoint_QueryUsesJsonbSet inspects the literal SQL string used
// by SaveCheckpoint to confirm it contains the jsonb_set call (and not the
// old `metadata = $1` whole-blob write). Cheap regression guard against
// silent reverts.
func TestSaveCheckpoint_QueryUsesJsonbSet(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	monitor := NewJobMonitor(db)

	// Require both jsonb_set and the {checkpoint} JSON pointer to appear
	// in the executed SQL.
	mock.ExpectExec(`jsonb_set[\s\S]+\{checkpoint\}[\s\S]+\{last_checkpoint_time\}`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := monitor.SaveCheckpoint(context.Background(), 1, map[string]interface{}{"a": 1}); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestSaveCheckpoint_NoSelectRoundTrip is a stricter complement to the
// first test: it sets up a mock that *only* expects an UPDATE. Any SELECT
// before the UPDATE (i.e. a regression back to read-modify-write) will
// surface as "unexpected query" because no SELECT was queued.
func TestSaveCheckpoint_NoSelectRoundTrip(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	monitor := NewJobMonitor(db)

	mock.ExpectExec(`UPDATE woulder\.job_executions`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := monitor.SaveCheckpoint(context.Background(), 7, map[string]interface{}{}); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
