package main

import (
	"database/sql"
	"testing"
	"time"
)

func TestShouldSyncKayaProgress(t *testing.T) {
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	recentLastSync := now.Add(-2 * time.Hour)
	oldLastSync := now.Add(-26 * time.Hour)
	futureNextSync := now.Add(2 * time.Hour)
	pastNextSync := now.Add(-1 * time.Hour)

	tests := []struct {
		name     string
		progress *kayaSyncStatus
		want     bool
	}{
		{
			name:     "syncs when no progress exists",
			progress: nil,
			want:     true,
		},
		{
			name: "syncs incomplete progress",
			progress: &kayaSyncStatus{
				Status:     sql.NullString{String: "in_progress", Valid: true},
				LastSyncAt: sql.NullTime{Time: recentLastSync, Valid: true},
				NextSyncAt: sql.NullTime{Time: futureNextSync, Valid: true},
			},
			want: true,
		},
		{
			name: "syncs failed progress",
			progress: &kayaSyncStatus{
				Status:     sql.NullString{String: "failed", Valid: true},
				LastSyncAt: sql.NullTime{Time: recentLastSync, Valid: true},
				NextSyncAt: sql.NullTime{Time: futureNextSync, Valid: true},
			},
			want: true,
		},
		{
			name: "syncs completed progress with no last sync",
			progress: &kayaSyncStatus{
				Status: sql.NullString{String: "completed", Valid: true},
			},
			want: true,
		},
		{
			name: "skips completed progress before next sync",
			progress: &kayaSyncStatus{
				Status:     sql.NullString{String: "completed", Valid: true},
				LastSyncAt: sql.NullTime{Time: recentLastSync, Valid: true},
				NextSyncAt: sql.NullTime{Time: futureNextSync, Valid: true},
			},
			want: false,
		},
		{
			name: "syncs completed progress after next sync",
			progress: &kayaSyncStatus{
				Status:     sql.NullString{String: "completed", Valid: true},
				LastSyncAt: sql.NullTime{Time: recentLastSync, Valid: true},
				NextSyncAt: sql.NullTime{Time: pastNextSync, Valid: true},
			},
			want: true,
		},
		{
			name: "skips completed progress within fallback 24 hour window",
			progress: &kayaSyncStatus{
				Status:     sql.NullString{String: "completed", Valid: true},
				LastSyncAt: sql.NullTime{Time: recentLastSync, Valid: true},
			},
			want: false,
		},
		{
			name: "syncs completed progress outside fallback 24 hour window",
			progress: &kayaSyncStatus{
				Status:     sql.NullString{String: "completed", Valid: true},
				LastSyncAt: sql.NullTime{Time: oldLastSync, Valid: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSyncKayaProgress(tt.progress, now); got != tt.want {
				t.Fatalf("shouldSyncKayaProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}
