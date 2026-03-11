package client

import (
	"testing"
)

func TestParseTimestampUTC(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantMonth int
		wantDay   int
		wantHour  int
		wantErr   bool
	}{
		{
			name:      "bare timestamp treated as UTC",
			input:     "2026-03-14T15:00",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   14,
			wantHour:  15,
			wantErr:   false,
		},
		{
			name:      "midnight UTC stays midnight UTC",
			input:     "2026-03-11T00:00",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   11,
			wantHour:  0,
			wantErr:   false,
		},
		{
			name:      "RFC3339 timestamp with Z suffix",
			input:     "2026-03-14T15:00:00Z",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   14,
			wantHour:  15,
			wantErr:   false,
		},
		{
			name:      "RFC3339 timestamp with offset",
			input:     "2026-03-14T08:00:00-07:00",
			wantYear:  2026,
			wantMonth: 3,
			wantDay:   14,
			wantHour:  15, // 8am Pacific = 3pm UTC
			wantErr:   false,
		},
		{
			name:    "invalid format",
			input:   "not-a-timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimestampUTC(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTimestampUTC(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTimestampUTC(%q) unexpected error: %v", tt.input, err)
			}

			if got.Year() != tt.wantYear {
				t.Errorf("Year = %d, want %d", got.Year(), tt.wantYear)
			}
			if int(got.Month()) != tt.wantMonth {
				t.Errorf("Month = %d, want %d", got.Month(), tt.wantMonth)
			}
			if got.Day() != tt.wantDay {
				t.Errorf("Day = %d, want %d", got.Day(), tt.wantDay)
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("Hour = %d, want %d", got.Hour(), tt.wantHour)
			}
			if got.Location().String() != "UTC" {
				t.Errorf("Location = %q, want UTC", got.Location().String())
			}
		})
	}
}

func TestParseSunTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHour int // Expected UTC hour
		wantErr  bool
	}{
		{
			name:     "bare sunrise timestamp as Pacific (PST, UTC-8)",
			input:    "2026-01-15T07:30",
			wantHour: 15, // 7:30am PST = 3:30pm UTC
			wantErr:  false,
		},
		{
			name:     "bare sunset timestamp as Pacific (PDT, UTC-7)",
			input:    "2026-06-15T20:45",
			wantHour: 3, // 8:45pm PDT = 3:45am+1 UTC
			wantErr:  false,
		},
		{
			name:    "invalid format",
			input:   "not-a-timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSunTimestamp(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSunTimestamp(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSunTimestamp(%q) unexpected error: %v", tt.input, err)
			}

			if got.Hour() != tt.wantHour {
				t.Errorf("Hour = %d, want %d (input: %s, parsed UTC: %s)", got.Hour(), tt.wantHour, tt.input, got.String())
			}
			if got.Location().String() != "UTC" {
				t.Errorf("Location = %q, want UTC", got.Location().String())
			}
		})
	}
}

func TestIsNightTime(t *testing.T) {
	tests := []struct {
		name    string
		time    string
		sunrise string
		sunset  string
		want    bool
	}{
		{
			name:    "midday is not night",
			time:    "2026-03-14T12:00",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    false,
		},
		{
			name:    "before sunrise is night",
			time:    "2026-03-14T05:00",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    true,
		},
		{
			name:    "after sunset is night",
			time:    "2026-03-14T20:00",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    true,
		},
		{
			name:    "at sunset is night",
			time:    "2026-03-14T18:30",
			sunrise: "2026-03-14T06:30",
			sunset:  "2026-03-14T18:30",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNightTime(tt.time, tt.sunrise, tt.sunset)
			if got != tt.want {
				t.Errorf("isNightTime(%q, %q, %q) = %v, want %v", tt.time, tt.sunrise, tt.sunset, got, tt.want)
			}
		})
	}
}
