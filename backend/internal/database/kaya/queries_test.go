package kaya

import (
	"strings"
	"testing"
)

func TestWriteAmplificationGuards(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "location upsert has no-op guard",
			query: querySaveLocation,
			want: []string{
				"WHERE kaya_locations.slug",
				"IS DISTINCT FROM EXCLUDED.name",
				"IS DISTINCT FROM EXCLUDED.woulder_location_id",
			},
		},
		{
			name:  "sync progress upsert has no-op guard",
			query: querySaveSyncProgress,
			want: []string{
				"WHERE kaya_sync_progress.location_name",
				"IS DISTINCT FROM EXCLUDED.status",
				"IS DISTINCT FROM EXCLUDED.next_sync_at",
				"IS DISTINCT FROM EXCLUDED.sub_locations_synced",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range tt.want {
				if !strings.Contains(tt.query, want) {
					t.Fatalf("query missing %q", want)
				}
			}
		})
	}
}
