package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/models"
)

type referenceCrag struct {
	Root      referenceArea    `json:"root"`
	Trails    []referenceTrail `json:"trails"`
	Notes     []referenceNote  `json:"notes"`
	Trailhead *referencePoint  `json:"trailhead"`
}

type referencePoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type referenceArea struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Name     string             `json:"name"`
	Kind     string             `json:"kind"`
	Aspect   string             `json:"aspect"`
	Desc     string             `json:"desc"`
	CX       float64            `json:"cx"`
	CY       float64            `json:"cy"`
	Boundary [][]float64        `json:"boundary"`
	Children []referenceArea    `json:"children"`
	Boulders []referenceBoulder `json:"boulders"`
}

type referenceBoulder struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Name     string             `json:"name"`
	CX       float64            `json:"cx"`
	CY       float64            `json:"cy"`
	Outline  [][]float64        `json:"outline"`
	Dev      string             `json:"dev"`
	Problems []referenceProblem `json:"problems"`
}

type referenceProblem struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Grade  string   `json:"grade"`
	Status string   `json:"status"`
	Stars  int      `json:"stars"`
	FA     *string  `json:"fa"`
	Types  []string `json:"types"`
}

type referenceTrail struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Source string      `json:"source"`
	Dist   string      `json:"dist"`
	Gain   string      `json:"gain"`
	From   string      `json:"from"`
	To     string      `json:"to"`
	Points [][]float64 `json:"points"`
}

type referenceNote struct {
	ID     string          `json:"id"`
	Author string          `json:"author"`
	Date   string          `json:"date"`
	Body   string          `json:"body"`
	Tags   []string        `json:"tags"`
	Target referenceTarget `json:"target"`
	Blocks json.RawMessage `json:"blocks"`
}

type referenceTarget struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (s *MoneyService) ImportReferenceCrag(ctx context.Context, projectID string, r io.Reader, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	var crag referenceCrag
	if err := json.NewDecoder(r).Decode(&crag); err != nil {
		return err
	}
	idMap := map[string]string{}
	if err := s.importReferenceArea(ctx, projectID, nil, crag.Root, 0, user, idMap); err != nil {
		return err
	}
	for i, tr := range crag.Trails {
		if err := s.importReferenceTrail(ctx, projectID, tr, i, user, idMap); err != nil {
			return err
		}
	}
	for _, note := range crag.Notes {
		if err := s.importReferenceNote(ctx, projectID, note, user, idMap); err != nil {
			return err
		}
	}
	return nil
}

func (s *MoneyService) importReferenceArea(ctx context.Context, projectID string, parentID *string, area referenceArea, sortOrder int, user models.CurrentUser, idMap map[string]string) error {
	props := map[string]interface{}{"kind": area.Kind, "aspect": area.Aspect, "cx": area.CX, "cy": area.CY}
	f, err := s.upsertImportedFeature(ctx, projectID, parentID, models.MoneyFeatureArea, area.ID, area.Name, area.Desc, models.MoneyStatusActive, polygonGeoJSON(area.Boundary), props, sortOrder, user)
	if err != nil {
		return err
	}
	idMap[area.ID] = f.ID
	for i, child := range area.Children {
		if err := s.importReferenceArea(ctx, projectID, &f.ID, child, i, user, idMap); err != nil {
			return err
		}
	}
	for i, boulder := range area.Boulders {
		if err := s.importReferenceBoulder(ctx, projectID, f.ID, boulder, i, user, idMap); err != nil {
			return err
		}
	}
	return nil
}

func (s *MoneyService) importReferenceBoulder(ctx context.Context, projectID, parentID string, b referenceBoulder, sortOrder int, user models.CurrentUser, idMap map[string]string) error {
	status := b.Dev
	if status == "" {
		status = models.MoneyStatusScouted
	}
	props := map[string]interface{}{"cx": b.CX, "cy": b.CY}
	f, err := s.upsertImportedFeature(ctx, projectID, &parentID, models.MoneyFeatureBoulder, b.ID, b.Name, "", status, polygonGeoJSON(b.Outline), props, sortOrder, user)
	if err != nil {
		return err
	}
	idMap[b.ID] = f.ID
	for i, p := range b.Problems {
		props := map[string]interface{}{"grade": p.Grade, "stars": p.Stars, "types": p.Types, "cx": b.CX, "cy": b.CY}
		if p.FA != nil {
			props["fa"] = *p.FA
		}
		if _, err := s.upsertImportedFeature(ctx, projectID, &f.ID, models.MoneyFeatureProblem, p.ID, p.Name, "", p.Status, pointGeoJSON(b.CX, b.CY), props, i, user); err != nil {
			return err
		}
	}
	return nil
}

func (s *MoneyService) importReferenceTrail(ctx context.Context, projectID string, tr referenceTrail, sortOrder int, user models.CurrentUser, idMap map[string]string) error {
	props := map[string]interface{}{"source": tr.Source, "dist": tr.Dist, "gain": tr.Gain, "from": tr.From, "to": tr.To}
	f, err := s.upsertImportedFeature(ctx, projectID, nil, models.MoneyFeatureTrail, tr.ID, tr.Name, "", models.MoneyStatusActive, lineGeoJSON(tr.Points), props, sortOrder, user)
	if err != nil {
		return err
	}
	idMap[tr.ID] = f.ID
	return nil
}

func (s *MoneyService) importReferenceNote(ctx context.Context, projectID string, note referenceNote, user models.CurrentUser, idMap map[string]string) error {
	body := strings.TrimSpace(note.Body)
	if body == "" {
		return nil
	}
	targetType := note.Target.Type
	if targetType == "" {
		targetType = "none"
	}
	var featureID *string
	var targetRef *string
	if mapped := idMap[note.Target.ID]; mapped != "" {
		featureID = &mapped
		targetRef = &mapped
	}
	blocks := note.Blocks
	if len(blocks) == 0 {
		blocks = json.RawMessage(`[]`)
	}
	ext := "note:" + note.ID
	_, err := s.repo.CreateNote(ctx, models.MoneyNote{ProjectID: projectID, FeatureID: featureID, TargetType: targetType, TargetRef: targetRef, Body: body, Visibility: models.MoneyNoteTeam, Tags: cleanTags(note.Tags), Blocks: blocks, ExternalRef: &ext, ImportSource: moneyStringPtr(moneyReferenceImportSource), CreatedBy: user.ID, UpdatedBy: user.ID})
	return err
}

func (s *MoneyService) upsertImportedFeature(ctx context.Context, projectID string, parentID *string, typ, extID, title, desc, status string, geojson json.RawMessage, props map[string]interface{}, sortOrder int, user models.CurrentUser) (*models.MoneyFeature, error) {
	if strings.TrimSpace(extID) == "" {
		return nil, fmt.Errorf("missing reference id for %s", title)
	}
	bbox, err := ValidateGeoJSON(geojson)
	if err != nil {
		return nil, err
	}
	rawProps, _ := json.Marshal(props)
	ext := typ + ":" + extID
	return s.repo.UpsertFeatureByExternalRef(ctx, models.MoneyFeature{ProjectID: projectID, ParentFeatureID: parentID, FeatureType: typ, Title: title, Description: cleanOptional(&desc, 2000), Status: status, GeoJSON: geojson, Style: json.RawMessage(`{}`), Properties: rawProps, MinLat: &bbox.MinLat, MinLon: &bbox.MinLon, MaxLat: &bbox.MaxLat, MaxLon: &bbox.MaxLon, SortOrder: sortOrder, ExternalRef: &ext, ImportSource: moneyStringPtr(moneyReferenceImportSource), CreatedBy: user.ID, UpdatedBy: user.ID})
}

func polygonGeoJSON(points [][]float64) json.RawMessage {
	closed := append([][]float64{}, points...)
	if len(closed) > 0 {
		first := closed[0]
		last := closed[len(closed)-1]
		if first[0] != last[0] || first[1] != last[1] {
			closed = append(closed, []float64{first[0], first[1]})
		}
	}
	raw, _ := json.Marshal(map[string]interface{}{"type": "Polygon", "coordinates": [][][]float64{closed}})
	return raw
}

func lineGeoJSON(points [][]float64) json.RawMessage {
	raw, _ := json.Marshal(map[string]interface{}{"type": "LineString", "coordinates": points})
	return raw
}

func pointGeoJSON(x, y float64) json.RawMessage {
	raw, _ := json.Marshal(map[string]interface{}{"type": "Point", "coordinates": []float64{x, y}})
	return raw
}

func moneyStringPtr(v string) *string { return &v }
