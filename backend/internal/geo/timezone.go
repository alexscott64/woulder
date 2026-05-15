// Package geo provides geographic helpers used across the backend.
//
// The primary entry point today is LookupTimezone, which maps a (lat, lon)
// to an IANA timezone name using the offline tzf polygon dataset.
package geo

import (
	"log"
	"sync"

	"github.com/ringsaturn/tzf"
	tzfrellite "github.com/ringsaturn/tzf-rel-lite"
	pb "github.com/ringsaturn/tzf/gen/go/tzf/v1"
	"google.golang.org/protobuf/proto"
)

// defaultTimezone is the defensive fallback used when the tzf lookup fails or
// returns an empty string. It matches the migration default for
// woulder.locations.timezone (see 000037_add_location_timezone.up.sql) so that
// behavior is consistent with rows that have not yet been backfilled.
const defaultTimezone = "America/Los_Angeles"

var (
	finderOnce sync.Once
	finder     tzf.F
	finderErr  error
)

// initFinder loads the embedded tzf-rel-lite dataset and constructs a
// tzf.Finder. Called once via sync.Once on the first LookupTimezone call.
func initFinder() {
	input := &pb.CompressedTimezones{}
	if err := proto.Unmarshal(tzfrellite.LiteCompressData, input); err != nil {
		finderErr = err
		log.Printf("geo: failed to unmarshal tzf-rel-lite data: %v", err)
		return
	}
	f, err := tzf.NewFinderFromCompressed(input)
	if err != nil {
		finderErr = err
		log.Printf("geo: failed to construct tzf finder: %v", err)
		return
	}
	finder = f
}

// LookupTimezone returns the IANA timezone name for the given coordinates.
// Returns "America/Los_Angeles" as a defensive fallback if the lookup fails
// or returns an empty string. Never returns an error.
//
// Note: the underlying tzf.Finder.GetTimezoneName takes (lng, lat) — this
// wrapper takes (lat, lon) to match the calling convention used everywhere
// else in this codebase (e.g. models.Location, locations.PostgresRepository).
func LookupTimezone(lat, lon float64) string {
	finderOnce.Do(initFinder)
	if finder == nil {
		return defaultTimezone
	}
	name := finder.GetTimezoneName(lon, lat)
	if name == "" {
		return defaultTimezone
	}
	return name
}
