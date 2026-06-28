package service

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"

	"github.com/alexscott64/woulder/backend/internal/models"
)

const (
	moneyGPXImportSource     = "gpx_onx_money_creek"
	moneyLegacyInverseSource = "legacy_world_inverse"
	moneyWorldWidth          = 1000.0
	moneyWorldHeight         = 680.0
	moneyWorldMargin         = 95.0
	moneyCreekMinLat         = 47.695
	moneyCreekMaxLat         = 47.707
	moneyCreekMinLon         = -121.492
	moneyCreekMaxLon         = -121.458
	metersPerLonDegree       = 111320.0
	metersPerLatDegree       = 110540.0
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

type gpxSurvey struct {
	Waypoints map[string]gpxPoint
	Routes    []gpxRoute
	Transform worldInverseTransform
}

type gpxPoint struct {
	Name string
	Lon  float64
	Lat  float64
}

type gpxRoute struct {
	Name   string
	Points [][]float64
}

type worldInverseTransform struct {
	Lat0    float64
	MinX    float64
	MinY    float64
	Scale   float64
	OffsetX float64
	OffsetY float64
}

type gpxXML struct {
	Waypoints []gpxWpt `xml:"wpt"`
	Routes    []gpxRte `xml:"rte"`
}

type gpxWpt struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Name string  `xml:"name"`
}

type gpxRte struct {
	Name   string     `xml:"name"`
	Points []gpxRtePt `xml:"rtept"`
}

type gpxRtePt struct {
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
}

func (s *MoneyService) ImportReferenceCrag(ctx context.Context, projectID string, r io.Reader, user models.CurrentUser) error {
	return s.ImportReferenceCragWithGPX(ctx, projectID, r, nil, user)
}

func (s *MoneyService) ImportReferenceCragWithGPX(ctx context.Context, projectID string, r io.Reader, gpxReader io.Reader, user models.CurrentUser) error {
	if !models.CanWriteMoney(user.Role) {
		return ErrMoneyForbidden
	}
	var crag referenceCrag
	if err := json.NewDecoder(r).Decode(&crag); err != nil {
		return err
	}
	survey, err := parseMoneyGPX(gpxReader)
	if err != nil {
		return err
	}
	idMap := map[string]string{}
	if err := s.importReferenceArea(ctx, projectID, nil, crag.Root, 0, user, idMap, survey); err != nil {
		return err
	}
	for i, tr := range crag.Trails {
		var route *gpxRoute
		if survey != nil && i < len(survey.Routes) {
			route = &survey.Routes[i]
		}
		if err := s.importReferenceTrail(ctx, projectID, tr, i, user, idMap, survey, route); err != nil {
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

func (s *MoneyService) importReferenceArea(ctx context.Context, projectID string, parentID *string, area referenceArea, sortOrder int, user models.CurrentUser, idMap map[string]string, survey *gpxSurvey) error {
	boundary := transformWorldPoints(area.Boundary, survey)
	center := transformWorldPoint([]float64{area.CX, area.CY}, survey)
	props := map[string]interface{}{
		"kind":                  area.Kind,
		"aspect":                area.Aspect,
		"center":                center,
		"coordinate_source":     moneyLegacyInverseSource,
		"coordinate_confidence": "generated_from_gpx_bbox_fit",
		"legacy_world":          map[string]interface{}{"cx": area.CX, "cy": area.CY, "boundary": area.Boundary},
	}
	f, err := s.upsertImportedFeature(ctx, projectID, parentID, models.MoneyFeatureArea, area.ID, area.Name, area.Desc, models.MoneyStatusActive, polygonGeoJSON(boundary), props, sortOrder, user)
	if err != nil {
		return err
	}
	idMap[area.ID] = f.ID
	for i, child := range area.Children {
		if err := s.importReferenceArea(ctx, projectID, &f.ID, child, i, user, idMap, survey); err != nil {
			return err
		}
	}
	for i, boulder := range area.Boulders {
		if err := s.importReferenceBoulder(ctx, projectID, f.ID, boulder, i, user, idMap, survey); err != nil {
			return err
		}
	}
	return nil
}

func (s *MoneyService) importReferenceBoulder(ctx context.Context, projectID, parentID string, b referenceBoulder, sortOrder int, user models.CurrentUser, idMap map[string]string, survey *gpxSurvey) error {
	status := b.Dev
	if status == "" {
		status = models.MoneyStatusScouted
	}
	center := transformWorldPoint([]float64{b.CX, b.CY}, survey)
	source := moneyLegacyInverseSource
	confidence := "generated_from_gpx_bbox_fit"
	if wp, ok := surveyWaypoint(survey, b.Name); ok {
		center = []float64{wp.Lon, wp.Lat}
		source = moneyGPXImportSource
		confidence = "surveyed_waypoint"
	}
	outline := transformWorldPoints(b.Outline, survey)
	props := map[string]interface{}{
		"center":                center,
		"coordinate_source":     source,
		"coordinate_confidence": confidence,
		"legacy_world":          map[string]interface{}{"cx": b.CX, "cy": b.CY, "outline": b.Outline},
	}
	f, err := s.upsertImportedFeature(ctx, projectID, &parentID, models.MoneyFeatureBoulder, b.ID, b.Name, "", status, polygonGeoJSON(outline), props, sortOrder, user)
	if err != nil {
		return err
	}
	idMap[b.ID] = f.ID
	for i, p := range b.Problems {
		props := map[string]interface{}{"grade": p.Grade, "stars": p.Stars, "types": p.Types, "center": center, "coordinate_source": source, "coordinate_confidence": confidence, "legacy_world": map[string]interface{}{"cx": b.CX, "cy": b.CY}}
		if p.FA != nil {
			props["fa"] = *p.FA
		}
		if _, err := s.upsertImportedFeature(ctx, projectID, &f.ID, models.MoneyFeatureProblem, p.ID, p.Name, "", p.Status, pointGeoJSON(center[0], center[1]), props, i, user); err != nil {
			return err
		}
	}
	return nil
}

func (s *MoneyService) importReferenceTrail(ctx context.Context, projectID string, tr referenceTrail, sortOrder int, user models.CurrentUser, idMap map[string]string, survey *gpxSurvey, route *gpxRoute) error {
	points := transformWorldPoints(tr.Points, survey)
	source := moneyLegacyInverseSource
	confidence := "generated_from_gpx_bbox_fit"
	if route != nil && len(route.Points) >= 2 {
		points = route.Points
		source = moneyGPXImportSource
		confidence = "surveyed_route"
	}
	props := map[string]interface{}{"source": tr.Source, "dist": tr.Dist, "gain": tr.Gain, "from": tr.From, "to": tr.To, "coordinate_source": source, "coordinate_confidence": confidence, "legacy_world": map[string]interface{}{"points": tr.Points}}
	if route != nil {
		props["gpx_route_name"] = route.Name
	}
	f, err := s.upsertImportedFeature(ctx, projectID, nil, models.MoneyFeatureTrail, tr.ID, tr.Name, "", models.MoneyStatusActive, lineGeoJSON(points), props, sortOrder, user)
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

func parseMoneyGPX(r io.Reader) (*gpxSurvey, error) {
	if r == nil {
		return nil, nil
	}
	var doc gpxXML
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, err
	}
	pointsForBounds := make([][]float64, 0)
	survey := &gpxSurvey{Waypoints: map[string]gpxPoint{}}
	for _, wpt := range doc.Waypoints {
		if !inMoneyBounds(wpt.Lon, wpt.Lat) {
			continue
		}
		p := gpxPoint{Name: strings.TrimSpace(wpt.Name), Lon: wpt.Lon, Lat: wpt.Lat}
		survey.Waypoints[normalizeGPXName(p.Name)] = p
		pointsForBounds = append(pointsForBounds, []float64{p.Lon, p.Lat})
	}
	for _, rte := range doc.Routes {
		pts := make([][]float64, 0, len(rte.Points))
		for _, pt := range rte.Points {
			if !inMoneyBounds(pt.Lon, pt.Lat) {
				continue
			}
			ll := []float64{pt.Lon, pt.Lat}
			pts = append(pts, ll)
			pointsForBounds = append(pointsForBounds, ll)
		}
		if len(pts) >= 2 {
			survey.Routes = append(survey.Routes, gpxRoute{Name: strings.TrimSpace(rte.Name), Points: decimatePoints(pts, 300)})
		}
	}
	if len(pointsForBounds) < 2 {
		return nil, fmt.Errorf("money GPX did not contain enough Money Creek lon/lat points")
	}
	survey.Transform = buildWorldInverseTransform(pointsForBounds)
	return survey, nil
}

func buildWorldInverseTransform(lonLat [][]float64) worldInverseTransform {
	minLon, minLat, maxLon, maxLat := 1e9, 1e9, -1e9, -1e9
	for _, p := range lonLat {
		minLon = minFloat(minLon, p[0])
		maxLon = maxFloat(maxLon, p[0])
		minLat = minFloat(minLat, p[1])
		maxLat = maxFloat(maxLat, p[1])
	}
	lat0 := (minLat + maxLat) / 2
	cosLat0 := math.Cos(lat0 * math.Pi / 180)
	minX := minLon * metersPerLonDegree * cosLat0
	maxX := maxLon * metersPerLonDegree * cosLat0
	minY := minLat * metersPerLatDegree
	maxY := maxLat * metersPerLatDegree
	spanX := maxX - minX
	spanY := maxY - minY
	scale := minFloat((moneyWorldWidth-2*moneyWorldMargin)/spanX, (moneyWorldHeight-2*moneyWorldMargin)/spanY)
	return worldInverseTransform{Lat0: lat0, MinX: minX, MinY: minY, Scale: scale, OffsetX: (moneyWorldWidth - spanX*scale) / 2, OffsetY: (moneyWorldHeight - spanY*scale) / 2}
}

func (t worldInverseTransform) worldToLonLat(p []float64) []float64 {
	if len(p) < 2 || t.Scale == 0 {
		return []float64{0, 0}
	}
	mx := t.MinX + (p[0]-t.OffsetX)/t.Scale
	my := t.MinY + (moneyWorldHeight-p[1]-t.OffsetY)/t.Scale
	return []float64{mx / (metersPerLonDegree * math.Cos(t.Lat0*math.Pi/180)), my / metersPerLatDegree}
}

func transformWorldPoint(p []float64, survey *gpxSurvey) []float64 {
	if survey == nil {
		return p
	}
	return survey.Transform.worldToLonLat(p)
}

func transformWorldPoints(points [][]float64, survey *gpxSurvey) [][]float64 {
	out := make([][]float64, 0, len(points))
	for _, p := range points {
		out = append(out, transformWorldPoint(p, survey))
	}
	return out
}

func surveyWaypoint(survey *gpxSurvey, name string) (gpxPoint, bool) {
	if survey == nil {
		return gpxPoint{}, false
	}
	p, ok := survey.Waypoints[normalizeGPXName(name)]
	return p, ok
}

var normalizeNameRE = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeGPXName(v string) string {
	return strings.Trim(normalizeNameRE.ReplaceAllString(strings.ToLower(strings.TrimSpace(v)), "-"), "-")
}

func inMoneyBounds(lon, lat float64) bool {
	return lon >= moneyCreekMinLon && lon <= moneyCreekMaxLon && lat >= moneyCreekMinLat && lat <= moneyCreekMaxLat
}

func decimatePoints(points [][]float64, maxPoints int) [][]float64 {
	if len(points) <= maxPoints || maxPoints < 2 {
		return points
	}
	out := make([][]float64, 0, maxPoints)
	step := float64(len(points)-1) / float64(maxPoints-1)
	for i := 0; i < maxPoints; i++ {
		out = append(out, points[int(math.Round(float64(i)*step))])
	}
	return out
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
