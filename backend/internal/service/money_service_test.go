package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func TestValidateGeoJSONDerivesBBox(t *testing.T) {
	raw := json.RawMessage(`{"type":"LineString","coordinates":[[-121.52,47.71],[-121.50,47.73]]}`)
	bbox, err := ValidateGeoJSON(raw)
	if err != nil {
		t.Fatalf("ValidateGeoJSON returned error: %v", err)
	}
	if bbox.MinLon != -121.52 || bbox.MaxLon != -121.50 || bbox.MinLat != 47.71 || bbox.MaxLat != 47.73 {
		t.Fatalf("unexpected bbox: %+v", bbox)
	}
}

func TestValidateAreaPolygonGeoJSONClosesRingAndDerivesBBox(t *testing.T) {
	raw := json.RawMessage(`{"type":"Polygon","coordinates":[[[-121.52,47.71],[-121.50,47.71],[-121.51,47.73]]]}`)
	bbox, normalized, err := ValidateAreaPolygonGeoJSON(raw)
	if err != nil {
		t.Fatalf("ValidateAreaPolygonGeoJSON returned error: %v", err)
	}
	if bbox.MinLon != -121.52 || bbox.MaxLon != -121.50 || bbox.MinLat != 47.71 || bbox.MaxLat != 47.73 {
		t.Fatalf("unexpected bbox: %+v", bbox)
	}
	var polygon struct {
		Coordinates [][][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(normalized, &polygon); err != nil {
		t.Fatalf("normalized polygon is invalid JSON: %v", err)
	}
	ring := polygon.Coordinates[0]
	first, last := ring[0], ring[len(ring)-1]
	if first[0] != last[0] || first[1] != last[1] {
		t.Fatalf("expected closed ring, got %v", ring)
	}
}

func TestValidateAreaPolygonGeoJSONRejectsWorldCoordinates(t *testing.T) {
	raw := json.RawMessage(`{"type":"Polygon","coordinates":[[[0,0],[100,0],[100,100],[0,0]]]}`)
	if _, _, err := ValidateAreaPolygonGeoJSON(raw); err == nil {
		t.Fatal("expected WGS84 polygon validation error")
	}
}

func TestFilterArchivedDescendantsHidesSubtree(t *testing.T) {
	root := moneyTestFeature("root", "", models.MoneyFeatureArea, models.MoneyStatusActive)
	area := moneyTestFeature("area", "root", models.MoneyFeatureArea, models.MoneyStatusArchived)
	boulder := moneyTestFeature("boulder", "area", models.MoneyFeatureBoulder, models.MoneyStatusScouted)
	problem := moneyTestFeature("problem", "boulder", models.MoneyFeatureProblem, models.MoneyStatusProject)
	sibling := moneyTestFeature("sibling", "root", models.MoneyFeatureArea, models.MoneyStatusActive)

	visible := filterArchivedDescendants([]models.MoneyFeature{root, area, boulder, problem, sibling})
	ids := make([]string, 0, len(visible))
	for _, f := range visible {
		ids = append(ids, f.ID)
	}
	if got, want := ids, []string{"root", "sibling"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("unexpected visible ids: got %v want %v", got, want)
	}
}

func TestBuildTrashItemsReturnsTopLevelDeletedAreasWithDescendants(t *testing.T) {
	root := moneyTestFeature("root", "", models.MoneyFeatureArea, models.MoneyStatusActive)
	area := moneyTestFeature("area", "root", models.MoneyFeatureArea, models.MoneyStatusArchived)
	child := moneyTestFeature("child", "area", models.MoneyFeatureArea, models.MoneyStatusActive)
	boulder := moneyTestFeature("boulder", "child", models.MoneyFeatureBoulder, models.MoneyStatusScouted)
	nestedDeleted := moneyTestFeature("nested", "area", models.MoneyFeatureArea, models.MoneyStatusArchived)

	items := buildTrashItems([]models.MoneyFeature{root, area, child, boulder, nestedDeleted})
	if len(items) != 1 {
		t.Fatalf("expected one top-level trash item, got %d", len(items))
	}
	if items[0].ID != "area" || items[0].DescendantCount != 3 {
		t.Fatalf("unexpected trash item: %+v", items[0])
	}
	if got := items[0].Path; len(got) != 2 || got[0] != "Root" || got[1] != "Area" {
		t.Fatalf("unexpected path: %v", got)
	}
}

func moneyTestFeature(id, parentID, featureType, status string) models.MoneyFeature {
	var parent *string
	if parentID != "" {
		parent = &parentID
	}
	return models.MoneyFeature{ID: id, ParentFeatureID: parent, FeatureType: featureType, Title: strings.Title(id), Status: status, UpdatedAt: time.Unix(int64(len(id)), 0)}
}

func TestValidateAreaPolygonGeoJSONRejectsTooFewDistinctVertices(t *testing.T) {
	raw := json.RawMessage(`{"type":"Polygon","coordinates":[[[-121.52,47.71],[-121.52,47.71],[-121.51,47.73],[-121.52,47.71]]]}`)
	if _, _, err := ValidateAreaPolygonGeoJSON(raw); err == nil {
		t.Fatal("expected distinct vertex validation error")
	}
}

func TestValidateGeoJSONRejectsUnsupportedGeometry(t *testing.T) {
	raw := json.RawMessage(`{"type":"MultiPoint","coordinates":[[-121.52,47.71]]}`)
	if _, err := ValidateGeoJSON(raw); err == nil {
		t.Fatal("expected unsupported geometry error")
	}
}

func TestValidateGeoJSONRejectsOutOfRangeCoordinates(t *testing.T) {
	raw := json.RawMessage(`{"type":"Point","coordinates":[-181,47.71]}`)
	if _, err := ValidateGeoJSON(raw); err == nil {
		t.Fatal("expected coordinate range error")
	}
}
