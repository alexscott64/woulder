package weather

import (
	"strings"
	"testing"
)

func TestWeatherUpsertNoOpGuards(t *testing.T) {
	for _, want := range []string{
		"WHERE weather_data.temperature",
		"IS DISTINCT FROM EXCLUDED.precipitation",
		"IS DISTINCT FROM EXCLUDED.dewpoint_f",
	} {
		if !strings.Contains(querySave, want) {
			t.Fatalf("querySave missing %q", want)
		}
	}

}

func TestWeatherDailyAggregateUpsertNoOpGuard(t *testing.T) {
	for _, want := range []string{
		"WHERE weather_daily_aggregates.min_temperature",
		"IS DISTINCT FROM EXCLUDED.total_precipitation",
		"IS DISTINCT FROM EXCLUDED.source_hour_count",
	} {
		if !strings.Contains(queryUpsertDailyAggregates, want) {
			t.Fatalf("queryUpsertDailyAggregates missing %q", want)
		}
	}
}
