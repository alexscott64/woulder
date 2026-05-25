package service

// LocationRootConfig defines the mapping between a Woulder location and its
// Mountain Project area roots that should be crawled by area-discovery /
// area-sync jobs.
//
// This is the single source of truth for "which MP areas belong to which
// woulder location". Both the manual `cmd/sync_climbs` backfill and the
// scheduled `SyncLocationAreaDiscovery` job consume this list so that adding
// or removing a root in one place propagates to both.
type LocationRootConfig struct {
	LocationName string
	LocationID   int
	MPAreaIDs    []int64
}

// defaultLocationRoots is the production list of MP area roots per location.
// Keep alphabetized-ish by location ID for readability.
var defaultLocationRoots = []LocationRootConfig{
	{
		LocationName: "Skykomish - Money Creek",
		LocationID:   1,
		MPAreaIDs:    []int64{120714486},
	},
	{
		LocationName: "Index",
		LocationID:   2,
		MPAreaIDs:    []int64{108123669},
	},
	{
		LocationName: "Gold Bar",
		LocationID:   3,
		MPAreaIDs:    []int64{105805788},
	},
	{
		LocationName: "Bellingham",
		LocationID:   4,
		MPAreaIDs:    []int64{107627792, 125093900, 108045031, 118561215},
	},
	{
		LocationName: "Icicle Creek (Leavenworth)",
		LocationID:   5,
		MPAreaIDs:    []int64{105790237, 105794001, 105790727},
	},
	{
		LocationName: "Squamish",
		LocationID:   6,
		// Stawamus Chief boulder areas (from within 105805895):
		//   - Grand Wall Boulders (112842712)
		//   - North Wall Boulders (108506197)
		//   - Apron Boulders (106025685) - contains sub-areas like Fantasia Boulders
		// Paradise Valley Boulders (110937821) - contains sub-areas
		// Powerline Boulders (121199811)
		MPAreaIDs: []int64{112842712, 108506197, 106025685, 110937821, 121199811},
	},
	{
		LocationName: "Skykomish - Paradise",
		LocationID:   7,
		MPAreaIDs:    []int64{120379690},
	},
	{
		LocationName: "Treasury",
		LocationID:   8,
		MPAreaIDs:    []int64{119589316},
	},
	{
		LocationName: "Calendar Butte",
		LocationID:   9,
		MPAreaIDs:    []int64{127029858},
	},
	{
		LocationName: "Joshua Tree",
		LocationID:   10,
		MPAreaIDs:    []int64{106098051},
	},
	{
		LocationName: "Black Mountain",
		LocationID:   11,
		MPAreaIDs:    []int64{105991127},
	},
	{
		LocationName: "Buttermilks",
		LocationID:   12,
		MPAreaIDs:    []int64{106132808},
	},
	{
		LocationName: "Happy / Sad Boulders",
		LocationID:   13,
		MPAreaIDs:    []int64{105799640, 106068462},
	},
	{
		LocationName: "Yosemite",
		LocationID:   14,
		MPAreaIDs:    []int64{107457415},
	},
	{
		LocationName: "Tramway",
		LocationID:   15,
		MPAreaIDs:    []int64{105991060},
	},
}

// LocationRoots is the package-level accessor for the list of MP area roots
// per location. It is a function (not a direct slice export) so tests can
// override the underlying provider via SetLocationRootsForTest without
// mutating shared global slice state in surprising ways.
//
// Callers MUST NOT mutate the returned slice.
func LocationRoots() []LocationRootConfig {
	return locationRootsProvider()
}

// locationRootsProvider is the indirection used by LocationRoots(). Production
// callers get defaultLocationRoots; tests can swap this via
// SetLocationRootsForTest to inject deterministic configs.
var locationRootsProvider = func() []LocationRootConfig {
	return defaultLocationRoots
}

// SetLocationRootsForTest replaces the LocationRoots provider for the duration
// of a test and returns a function that restores the previous provider.
// Intended for use only from *_test.go files in this package.
func SetLocationRootsForTest(roots []LocationRootConfig) func() {
	prev := locationRootsProvider
	locationRootsProvider = func() []LocationRootConfig {
		return roots
	}
	return func() {
		locationRootsProvider = prev
	}
}
