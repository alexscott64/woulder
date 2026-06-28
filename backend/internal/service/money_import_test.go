package service

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestParseMoneyGPXFiltersMoneyCreekWaypointsAndRoutes(t *testing.T) {
	file, err := os.Open("../database/money/fixtures/onx-markups-06232026.gpx")
	if err != nil {
		t.Fatalf("open GPX fixture: %v", err)
	}
	defer file.Close()

	survey, err := parseMoneyGPX(file)
	if err != nil {
		t.Fatalf("parseMoneyGPX returned error: %v", err)
	}
	if len(survey.Waypoints) < 30 {
		t.Fatalf("expected Money Creek waypoints, got %d", len(survey.Waypoints))
	}
	if len(survey.Routes) != 3 {
		t.Fatalf("expected 3 route geometries, got %d", len(survey.Routes))
	}
	wp, ok := surveyWaypoint(survey, "Dawnbreaker Boulder")
	if !ok {
		t.Fatal("missing Dawnbreaker waypoint")
	}
	if wp.Lon < -121.47 || wp.Lon > -121.45 || wp.Lat < 47.69 || wp.Lat > 47.71 {
		t.Fatalf("waypoint outside expected bounds: %+v", wp)
	}
	if _, ok := surveyWaypoint(survey, "nice Boulder"); ok {
		t.Fatal("far-off waypoint should be filtered out")
	}
}

func TestMoneyCreekFixtureUsesRequestedAreaNames(t *testing.T) {
	file, err := os.Open("../database/money/fixtures/money_creek_crag.json")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer file.Close()

	var crag referenceCrag
	if err := json.NewDecoder(file).Decode(&crag); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	rootNames := map[string]bool{}
	for _, child := range crag.Root.Children {
		rootNames[child.Name] = true
	}
	for _, name := range []string{"DotA Boulders", "Upper Money Creek", "Uphill Boulders", "GameCube"} {
		if !rootNames[name] {
			t.Fatalf("missing root area %q in %v", name, rootNames)
		}
	}

	var dota *referenceArea
	for i := range crag.Root.Children {
		if crag.Root.Children[i].Name == "DotA Boulders" {
			dota = &crag.Root.Children[i]
		}
	}
	if dota == nil {
		t.Fatal("DotA Boulders area not found")
	}
	childNames := map[string]bool{}
	for _, child := range dota.Children {
		childNames[child.Name] = true
	}
	if !childNames["Radiant Boulders"] || !childNames["Dire Boulders"] {
		t.Fatalf("DotA children mismatch: %v", childNames)
	}
}

func TestWorldInverseTransformMapsLegacyWorldBackToLonLat(t *testing.T) {
	points := [][]float64{{-121.48699, 47.69842}, {-121.46015, 47.70566}, {-121.46361, 47.70028}, {-121.48071, 47.69751}}
	tr := buildWorldInverseTransform(points)
	for _, lonLat := range points {
		world := lonLatToWorld(lonLat, tr)
		actual := tr.worldToLonLat(world)
		if math.Abs(actual[0]-lonLat[0]) > 0.0000001 || math.Abs(actual[1]-lonLat[1]) > 0.0000001 {
			t.Fatalf("round trip mismatch: input=%v world=%v actual=%v", lonLat, world, actual)
		}
	}
}

func lonLatToWorld(lonLat []float64, tr worldInverseTransform) []float64 {
	mx := lonLat[0] * metersPerLonDegree * math.Cos(tr.Lat0*math.Pi/180)
	my := lonLat[1] * metersPerLatDegree
	return []float64{(mx-tr.MinX)*tr.Scale + tr.OffsetX, moneyWorldHeight - ((my-tr.MinY)*tr.Scale + tr.OffsetY)}
}
