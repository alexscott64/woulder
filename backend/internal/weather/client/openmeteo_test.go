package client

import (
	"testing"
)

func TestMergePrecipitationData(t *testing.T) {
	tests := []struct {
		name    string
		namData []*float64
		bmData  []float64
		want    []float64
	}{
		{
			name: "NAM data available for first 3 hours, fallback to best_match after",
			namData: []*float64{
				ptrFloat64(0.1), ptrFloat64(0.2), ptrFloat64(0.3), nil, nil,
			},
			bmData: []float64{0.05, 0.15, 0.25, 0.35, 0.45},
			want:   []float64{0.1, 0.2, 0.3, 0.35, 0.45},
		},
		{
			name: "All NAM data available",
			namData: []*float64{
				ptrFloat64(0.1), ptrFloat64(0.2), ptrFloat64(0.3),
			},
			bmData: []float64{0.05, 0.15, 0.25},
			want:   []float64{0.1, 0.2, 0.3},
		},
		{
			name: "No NAM data (all nil), use best_match",
			namData: []*float64{
				nil, nil, nil,
			},
			bmData: []float64{0.05, 0.15, 0.25},
			want:   []float64{0.05, 0.15, 0.25},
		},
		{
			name: "NAM array shorter than best_match",
			namData: []*float64{
				ptrFloat64(0.1), ptrFloat64(0.2),
			},
			bmData: []float64{0.05, 0.15, 0.25, 0.35, 0.45},
			want:   []float64{0.1, 0.2, 0.25, 0.35, 0.45},
		},
		{
			name:    "Empty best_match array",
			namData: []*float64{ptrFloat64(0.1), ptrFloat64(0.2)},
			bmData:  []float64{},
			want:    []float64{},
		},
		{
			name:    "Both empty",
			namData: []*float64{},
			bmData:  []float64{},
			want:    []float64{},
		},
		{
			name: "Zero values are preserved (not treated as nil)",
			namData: []*float64{
				ptrFloat64(0.0), ptrFloat64(0.5), ptrFloat64(0.0), nil,
			},
			bmData: []float64{0.1, 0.2, 0.3, 0.4},
			want:   []float64{0.0, 0.5, 0.0, 0.4},
		},
		{
			name: "Simulates real NAM CONUS (60 hours then null)",
			namData: func() []*float64 {
				result := make([]*float64, 72)
				for i := 0; i < 60; i++ {
					result[i] = ptrFloat64(float64(i) * 0.01)
				}
				// Hours 60-71 are nil (NAM CONUS doesn't provide data)
				for i := 60; i < 72; i++ {
					result[i] = nil
				}
				return result
			}(),
			bmData: func() []float64 {
				result := make([]float64, 72)
				for i := 0; i < 72; i++ {
					result[i] = float64(i) * 0.02
				}
				return result
			}(),
			want: func() []float64 {
				result := make([]float64, 72)
				// First 60 hours from NAM
				for i := 0; i < 60; i++ {
					result[i] = float64(i) * 0.01
				}
				// Last 12 hours from best_match
				for i := 60; i < 72; i++ {
					result[i] = float64(i) * 0.02
				}
				return result
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergePrecipitationData(tt.namData, tt.bmData)

			if len(got) != len(tt.want) {
				t.Errorf("mergePrecipitationData() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("mergePrecipitationData()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// Helper function to create float64 pointers
func ptrFloat64(f float64) *float64 {
	return &f
}

func TestMergePrecipitationData_NilSafety(t *testing.T) {
	// Test that nil values are handled safely
	t.Run("nil NAM data", func(t *testing.T) {
		bmData := []float64{0.1, 0.2, 0.3}
		got := mergePrecipitationData(nil, bmData)

		if len(got) != len(bmData) {
			t.Errorf("Expected length %d, got %d", len(bmData), len(got))
		}

		for i := range got {
			if got[i] != bmData[i] {
				t.Errorf("Index %d: got %v, want %v", i, got[i], bmData[i])
			}
		}
	})

	t.Run("nil best_match data", func(t *testing.T) {
		namData := []*float64{ptrFloat64(0.1), ptrFloat64(0.2)}
		got := mergePrecipitationData(namData, nil)

		if len(got) != 0 {
			t.Errorf("Expected empty result for nil bmData, got length %d", len(got))
		}
	})

	t.Run("both nil", func(t *testing.T) {
		got := mergePrecipitationData(nil, nil)

		if len(got) != 0 {
			t.Errorf("Expected empty result for both nil, got length %d", len(got))
		}
	})
}

func TestMergePrecipitationData_EdgeCases(t *testing.T) {
	t.Run("very long arrays", func(t *testing.T) {
		// Test with 168 hours (7 days)
		namData := make([]*float64, 168)
		bmData := make([]float64, 168)

		// NAM provides first 60 hours
		for i := 0; i < 60; i++ {
			namData[i] = ptrFloat64(1.0)
			bmData[i] = 0.5
		}
		// After 60 hours, NAM is nil
		for i := 60; i < 168; i++ {
			namData[i] = nil
			bmData[i] = 0.5
		}

		got := mergePrecipitationData(namData, bmData)

		// Check first 60 use NAM
		for i := 0; i < 60; i++ {
			if got[i] != 1.0 {
				t.Errorf("Hour %d should use NAM (1.0), got %v", i, got[i])
			}
		}

		// Check remaining use best_match
		for i := 60; i < 168; i++ {
			if got[i] != 0.5 {
				t.Errorf("Hour %d should use best_match (0.5), got %v", i, got[i])
			}
		}
	})

	t.Run("single element arrays", func(t *testing.T) {
		namData := []*float64{ptrFloat64(0.123)}
		bmData := []float64{0.456}
		got := mergePrecipitationData(namData, bmData)

		if len(got) != 1 {
			t.Fatalf("Expected length 1, got %d", len(got))
		}
		if got[0] != 0.123 {
			t.Errorf("Expected 0.123, got %v", got[0])
		}
	})

	t.Run("large precipitation values", func(t *testing.T) {
		// Test with extreme weather (10 inches of rain)
		namData := []*float64{ptrFloat64(10.5), nil}
		bmData := []float64{5.2, 3.8}
		got := mergePrecipitationData(namData, bmData)

		if got[0] != 10.5 {
			t.Errorf("Expected 10.5, got %v", got[0])
		}
		if got[1] != 3.8 {
			t.Errorf("Expected 3.8, got %v", got[1])
		}
	})
}
