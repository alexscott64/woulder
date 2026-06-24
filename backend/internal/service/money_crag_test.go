package service

import (
	"encoding/json"
	"testing"

	"github.com/alexscott64/woulder/backend/internal/models"
)

func feature(id, typ, title string, parent *string) models.MoneyFeature {
	return models.MoneyFeature{ID: id, ParentFeatureID: parent, FeatureType: typ, Title: title, Status: models.MoneyStatusActive, GeoJSON: json.RawMessage(`{"type":"Point","coordinates":[0,0]}`)}
}

func TestBuildMoneyCragTreeAssemblesAreasBouldersProblemsAndTrails(t *testing.T) {
	rootID := "root"
	boulderID := "boulder"
	features := []models.MoneyFeature{
		feature(rootID, models.MoneyFeatureArea, "Root", nil),
		feature("child", models.MoneyFeatureArea, "Child", &rootID),
		feature(boulderID, models.MoneyFeatureBoulder, "Boulder", &rootID),
		feature("problem", models.MoneyFeatureProblem, "Problem", &boulderID),
		feature("trail", models.MoneyFeatureTrail, "Trail", nil),
	}
	root, trails := BuildMoneyCragTree(features)
	if root == nil || root.Feature.ID != rootID {
		t.Fatalf("unexpected root: %+v", root)
	}
	if len(root.Children) != 1 || len(root.Boulders) != 1 || len(root.Boulders[0].Problems) != 1 {
		t.Fatalf("tree not assembled: %+v", root)
	}
	if len(trails) != 1 || trails[0].Feature.ID != "trail" {
		t.Fatalf("unexpected trails: %+v", trails)
	}
}

func TestValidateGeoJSONAcceptsReferenceWorldCoordinates(t *testing.T) {
	raw := json.RawMessage(`{"type":"Polygon","coordinates":[[[0,0],[1000,0],[1000,680],[0,0]]]}`)
	bbox, err := ValidateGeoJSON(raw)
	if err != nil {
		t.Fatalf("ValidateGeoJSON returned error: %v", err)
	}
	if bbox.MaxLon != 1000 || bbox.MaxLat != 680 {
		t.Fatalf("unexpected bbox: %+v", bbox)
	}
}
