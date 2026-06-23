package service

import (
	"encoding/json"
	"testing"
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
