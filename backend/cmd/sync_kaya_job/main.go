package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/alexscott64/woulder/backend/internal/database"
	kayaClient "github.com/alexscott64/woulder/backend/internal/kaya"
	"github.com/alexscott64/woulder/backend/internal/monitoring"
	"github.com/alexscott64/woulder/backend/internal/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// KayaSyncJob runs scheduled syncs of Kaya data with monitoring
func main() {
	log.Println("Starting Kaya scheduled sync job...")

	// Command-line flags
	incrementalFlag := flag.Bool("incremental", true, "Only sync new data since last sync")
	testFlag := flag.Bool("test", false, "Test mode: only sync 3 destinations")
	delayFlag := flag.Int("delay", 3, "Delay in seconds between destinations")
	matchAfterSyncFlag := flag.Bool("match-after-sync", true, "Run Kaya↔MP matching after each successful location sync")
	matchMinConfidenceFlag := flag.Float64("match-min-confidence", 0.75, "Minimum confidence for Kaya↔MP route matching")
	flag.Parse()

	// Load environment variables - try current directory first, then parent
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Printf("Warning: .env file not found in . or .., using system environment variables")
		}
	}

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create a separate SQL connection for job monitoring
	monitorDB, err := createMonitoringDB()
	if err != nil {
		log.Fatalf("Failed to create monitoring database connection: %v", err)
	}
	defer monitorDB.Close()

	// Initialize job monitor
	jobMonitor := monitoring.NewJobMonitor(monitorDB)

	// Load destinations to determine total items
	destinations, err := loadDestinations()
	if err != nil {
		log.Fatalf("Failed to load destinations: %v", err)
	}

	// Test mode: only sync first 3
	if *testFlag {
		log.Println("TEST MODE: Only syncing first 3 destinations")
		if len(destinations) > 3 {
			destinations = destinations[:3]
		}
	}

	// Start job run
	jobName := "kaya_sync"
	jobType := "full"
	if *incrementalFlag {
		jobType = "incremental"
	}

	jobExec, err := jobMonitor.StartJob(context.Background(), jobName, jobType, len(destinations), map[string]interface{}{
		"incremental":          *incrementalFlag,
		"test_mode":            *testFlag,
		"delay":                *delayFlag,
		"match_after_sync":     *matchAfterSyncFlag,
		"match_min_confidence": *matchMinConfidenceFlag,
	})
	if err != nil {
		log.Fatalf("Failed to start job tracking: %v", err)
	}

	log.Printf("Job started with ID: %d", jobExec.ID)
	startTime := time.Now()

	// Run sync
	successCount, failCount := runSync(
		db,
		monitorDB,
		jobMonitor,
		jobExec.ID,
		destinations,
		*incrementalFlag,
		*delayFlag,
		*matchAfterSyncFlag,
		*matchMinConfidenceFlag,
	)

	// Complete job tracking
	duration := time.Since(startTime)

	if failCount > 0 && successCount == 0 {
		// Complete failure
		errMsg := "All destinations failed to sync"
		jobMonitor.FailJob(context.Background(), jobExec.ID, errMsg)
		log.Fatalf("Kaya sync job failed after %s: %s", duration, errMsg)
	}

	// Complete successfully (even with partial failures)
	jobMonitor.CompleteJob(context.Background(), jobExec.ID)
	log.Printf("✓ Kaya sync job completed in %s (success: %d, failed: %d)", duration, successCount, failCount)
}

func runSync(
	db *database.Database,
	sqlDB *sql.DB,
	jobMonitor *monitoring.JobMonitor,
	jobID int64,
	destinations []string,
	incremental bool,
	delay int,
	matchAfterSync bool,
	matchMinConfidence float64,
) (int, int) {
	ctx := context.Background()

	// Initialize Kaya client
	client := kayaClient.NewClient()

	// Initialize Kaya sync service
	kayaService := service.NewKayaSyncService(db.Kaya(), client, nil)

	successCount := 0
	failCount := 0
	processed := 0

	matchedCount := 0
	rejectedCount := 0

	for i, slug := range destinations {
		log.Printf("\n[%d/%d] Syncing %s...", i+1, len(destinations), slug)

		// For incremental sync, check if we need to sync this location
		if incremental {
			shouldSync, err := shouldSyncLocation(ctx, db, slug)
			if err != nil {
				log.Printf("Error checking sync status for %s: %v", slug, err)
			} else if !shouldSync {
				log.Printf("Skipping %s (recently synced)", slug)
				processed++
				jobMonitor.UpdateProgress(ctx, jobID, processed, successCount, failCount)
				continue
			}
		}

		// Sync location
		err := kayaService.SyncLocationBySlug(ctx, slug, true)
		processed++

		if err != nil {
			log.Printf("ERROR syncing %s: %v", slug, err)
			failCount++
		} else {
			successCount++
			log.Printf("✓ Synced %s", slug)

			if matchAfterSync {
				newMatches, rejected, matchErr := matchRoutesForDestinationSlug(ctx, sqlDB, slug, matchMinConfidence)
				if matchErr != nil {
					log.Printf("WARNING matching failed for %s: %v", slug, matchErr)
				} else {
					matchedCount += newMatches
					rejectedCount += rejected
					log.Printf("  ↳ Matching: %d saved, %d rejected (discipline/grade mismatch)", newMatches, rejected)
					_ = jobMonitor.UpdateCurrentItem(ctx, jobID, map[string]interface{}{
						"current_destination_slug": slug,
						"matching_saved":           newMatches,
						"matching_rejected":        rejected,
						"matching_total_saved":     matchedCount,
						"matching_total_rejected":  rejectedCount,
					})
				}
			}
		}

		// Update progress
		jobMonitor.UpdateProgress(ctx, jobID, processed, successCount, failCount)

		// Rate limiting
		if i < len(destinations)-1 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}

	log.Printf("\n========================================")
	log.Printf("Sync Summary:")
	log.Printf("Total: %d, Success: %d, Failed: %d", len(destinations), successCount, failCount)
	if matchAfterSync {
		log.Printf("Matching saved: %d, rejected: %d", matchedCount, rejectedCount)
	}
	log.Printf("========================================")

	return successCount, failCount
}

func loadDestinations() ([]string, error) {
	// Embedded list of all 105 Kaya official destinations
	// Source: https://kayaclimb.com/explore (extracted 2026-02-18)
	destinations := []string{
		"Squamish-295658",
		"Red-Rocks-331387",
		"Bishop-316882",
		"Joshua-Tree-317008",
		"Hueco-Tanks-339538",
		"Joes-Valley-340826",
		"Vancouver-Island-295813",
		"Clear-Creek-Canyon-323872",
		"Ogden-1153006",
		"Lincoln-Lake-5272477",
		"Guanella-Pass-323792",
		"Tahoe-317136",
		"Little-Cottonwood-Canyon-986245",
		"New-River-Gorge-347179",
		"Coopers-Rock-347182",
		"Smith-Rock-336540",
		"Black-Mountain-317072",
		"Leavenworth-344933",
		"Kelowna-296013",
		"Hatcher-Pass-314961",
		"Devils-Lake-348323",
		"Lake-Ramona-10400507",
		"RMNP-323755",
		"Tramway-317070",
		"Vancouver-296037",
		"Ibex-341212",
		"Stone-Fort-999671",
		"Mount-Woodson-2192166",
		"Red-Feather-324534",
		"Flagstaff-Mountain-323839",
		"Big-Cottonwood-Canyon-BCC-341957",
		"Fraser-Valley-3340725",
		"Reimers-Ranch-339808",
		"Horseshoe-Canyon-Ranch-316278",
		"Tulsa-OK-10116402",
		"Mineral-King-15161231",
		"Rumbling-Bald-335837",
		"Rocktown-327484",
		"Horse-Pens-40-983782",
		"Malibu-838425",
		"Santa-Barbara-317853",
		"Doyle-322152",
		"Comox-Valley-Vancouver-Island-BC-7882675",
		"NYC-Bouldering-8736175",
		"Moes-Valley-340851",
		"Gold-Bar-344983",
		"The-Nooks-3899367",
		"Adirondacks-335103",
		"Stoney-Point-317772",
		"Treasury-2106513",
		"Eldorado-Canyon-323915",
		"Uintas-1394571",
		"holy-boulders-1016922",
		"Gunpowder-Falls-1395399",
		"Boat-Rock-327557",
		"Reynolds-Creek-328023",
		"Triassic-341357",
		"Needle-Peak-658063",
		"Box-Springs-Mountain-Reserve-5727203",
		"Horse-Flats-317843",
		"Mt-Evans-323773",
		"Smugglers-Notch-344705",
		"Rock-shop-348813",
		"Morpheus-345195",
		"Berkeley-316984",
		"Mount-Rubidoux-321790",
		"Index-345070",
		"purgatory-851804",
		"Vernon-4132330",
		"Exit-38-345299",
		"Castle-Rock-State-Park-328014",
		"Sams-Throne-316415",
		"Patapsco-Valley-State-Park-8555804",
		"Porcupine-Hills-6234426",
		"Cowell-316321",
		"Dixon-School-Road-335964",
		"Barton-Creek-Greenbelt-339852",
		"Utah-Hills-341651",
		"Price-1361664",
		"Big-Rock-291216",
		"Rogers-Park-339768",
		"Salt-Point-317575",
		"The-Citadel-295573",
		"Sierra-Buttes-318225",
		"Hammond-Pond-330274",
		"Nut-Tree-990859",
		"Santee-Boulders-2376083",
		"Indian-Rock-3199690",
		"Juan-De-Fuca-7846367",
		"Richland-Creek-15036518",
		"Lost-Ledges-345403",
		"Lions-Den-334501",
		"Conejo-Mountain-9502266",
		"Mckinney-Falls-1493881",
		"Wadi-Rum-389777",
		"Rocks-State-Park-330207",
		"Sawmill-330569",
		"Mt-Tamalpais-318183",
		"Rock-Creek-317075",
		"Sugarloaf-Ridge-State-Park-1770584",
	}

	return destinations, nil
}

func shouldSyncLocation(ctx context.Context, db *database.Database, slug string) (bool, error) {
	// Check last sync time from kaya_sync_progress
	// For now, always sync (incremental logic can be added later)
	// Full implementation would check: last_synced_at > NOW() - INTERVAL '24 hours'
	return true, nil
}

type kayaClimbForMatching struct {
	ID        string
	Name      string
	Location  string
	Latitude  *float64
	Longitude *float64
	Grade     string
	ClimbType string
}

type routeMatchForSync struct {
	KayaClimbID       string
	KayaClimbName     string
	KayaLocationName  string
	MPRouteID         int64
	MPRouteName       string
	MPAreaName        string
	Confidence        float64
	MatchType         string
	NameSimilarity    float64
	DistanceKM        *float64
	LocationNameMatch bool
}

func matchRoutesForDestinationSlug(ctx context.Context, db *sql.DB, destinationSlug string, minConfidence float64) (int, int, error) {
	climbs, err := getKayaClimbsForDestinationSlug(ctx, db, destinationSlug)
	if err != nil {
		return 0, 0, err
	}

	saved := 0
	rejected := 0
	for _, climb := range climbs {
		matches, rejectedForClimb, err := findMPMatchesForSync(ctx, db, climb, minConfidence)
		rejected += rejectedForClimb
		if err != nil {
			return saved, rejected, err
		}
		for _, match := range matches {
			if err := saveRouteMatch(ctx, db, match); err != nil {
				return saved, rejected, err
			}
			saved++
		}
	}

	return saved, rejected, nil
}

func getKayaClimbsForDestinationSlug(ctx context.Context, db *sql.DB, destinationSlug string) ([]kayaClimbForMatching, error) {
	query := `
		SELECT
			c.slug,
			c.name,
			COALESCE(c.kaya_destination_name, c.kaya_area_name, 'Unknown') AS location_name,
			l.latitude,
			l.longitude,
			c.grade_name,
			c.climb_type_name
		FROM woulder.kaya_climbs c
		JOIN woulder.kaya_locations l ON c.kaya_destination_id = l.kaya_location_id
		WHERE l.slug = $1
			AND c.slug IS NOT NULL
			AND c.slug != ''
		ORDER BY c.name
	`

	rows, err := db.QueryContext(ctx, query, destinationSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to query kaya climbs for %s: %w", destinationSlug, err)
	}
	defer rows.Close()

	var climbs []kayaClimbForMatching
	for rows.Next() {
		var climb kayaClimbForMatching
		var lat, lon sql.NullFloat64
		var grade, climbType sql.NullString

		if err := rows.Scan(&climb.ID, &climb.Name, &climb.Location, &lat, &lon, &grade, &climbType); err != nil {
			return nil, fmt.Errorf("failed to scan kaya climb row: %w", err)
		}

		if lat.Valid {
			climb.Latitude = &lat.Float64
		}
		if lon.Valid {
			climb.Longitude = &lon.Float64
		}
		if grade.Valid {
			climb.Grade = grade.String
		}
		if climbType.Valid {
			climb.ClimbType = climbType.String
		}

		climbs = append(climbs, climb)
	}

	return climbs, rows.Err()
}

func findMPMatchesForSync(ctx context.Context, db *sql.DB, climb kayaClimbForMatching, minConfidence float64) ([]routeMatchForSync, int, error) {
	query := `
		SELECT
			r.mp_route_id,
			r.name,
			COALESCE(a.name, 'Unknown') AS area_name,
			a.latitude,
			a.longitude,
			r.route_type,
			r.rating
		FROM woulder.mp_routes r
		LEFT JOIN woulder.mp_areas a ON r.mp_area_id = a.mp_area_id
		WHERE LOWER(r.name) LIKE LOWER($1)
		LIMIT 20
	`

	rows, err := db.QueryContext(ctx, query, "%"+climb.Name+"%")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query mp routes for %s: %w", climb.Name, err)
	}
	defer rows.Close()

	var matches []routeMatchForSync
	rejected := 0

	for rows.Next() {
		var mpID int64
		var mpName, mpArea, mpRouteType, mpRating string
		var mpLat, mpLon sql.NullFloat64

		if err := rows.Scan(&mpID, &mpName, &mpArea, &mpLat, &mpLon, &mpRouteType, &mpRating); err != nil {
			continue
		}

		nameSim := calculateNameSimilarity(climb.Name, mpName)
		locationMatch := matchLocationNames(climb.Location, mpArea)

		var distKM *float64
		if climb.Latitude != nil && climb.Longitude != nil && mpLat.Valid && mpLon.Valid {
			d := calculateGPSDistance(*climb.Latitude, *climb.Longitude, mpLat.Float64, mpLon.Float64)
			distKM = &d
		}

		if !isCompatibleMatch(climb.ClimbType, climb.Grade, mpRouteType, mpRating) {
			rejected++
			continue
		}

		confidence := calculateMatchConfidence(nameSim, locationMatch, distKM)
		if confidence < minConfidence {
			continue
		}

		matches = append(matches, routeMatchForSync{
			KayaClimbID:       climb.ID,
			KayaClimbName:     climb.Name,
			KayaLocationName:  climb.Location,
			MPRouteID:         mpID,
			MPRouteName:       mpName,
			MPAreaName:        mpArea,
			Confidence:        confidence,
			MatchType:         determineMatchType(nameSim, locationMatch, distKM),
			NameSimilarity:    nameSim,
			DistanceKM:        distKM,
			LocationNameMatch: locationMatch,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, rejected, err
	}

	return matches, rejected, nil
}

func saveRouteMatch(ctx context.Context, db *sql.DB, match routeMatchForSync) error {
	query := `
		INSERT INTO kaya_mp_route_matches (
			kaya_climb_id, mp_route_id, match_confidence, match_type,
			kaya_climb_name, kaya_location_name,
			mp_route_name, mp_area_name,
			name_similarity, location_name_match, location_distance_km
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (kaya_climb_id, mp_route_id) DO UPDATE SET
			match_confidence = EXCLUDED.match_confidence,
			match_type = EXCLUDED.match_type,
			name_similarity = EXCLUDED.name_similarity,
			location_name_match = EXCLUDED.location_name_match,
			location_distance_km = EXCLUDED.location_distance_km,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.ExecContext(ctx, query,
		match.KayaClimbID,
		match.MPRouteID,
		match.Confidence,
		match.MatchType,
		match.KayaClimbName,
		match.KayaLocationName,
		match.MPRouteName,
		match.MPAreaName,
		match.NameSimilarity,
		match.LocationNameMatch,
		match.DistanceKM,
	)
	return err
}

func calculateNameSimilarity(name1, name2 string) float64 {
	n1 := normalizeRouteName(name1)
	n2 := normalizeRouteName(name2)
	if n1 == n2 {
		return 1.0
	}
	distance := levenshteinDistance(n1, n2)
	maxLen := float64(max(len(n1), len(n2)))
	if maxLen == 0 {
		return 0.0
	}
	return 1.0 - (float64(distance) / maxLen)
}

func normalizeRouteName(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimPrefix(name, "the ")
	name = strings.TrimPrefix(name, "a ")
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return -1
	}, name)
	return strings.TrimSpace(strings.Join(strings.Fields(name), " "))
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(matrix[i-1][j]+1, min(matrix[i][j-1]+1, matrix[i-1][j-1]+cost))
		}
	}
	return matrix[len(s1)][len(s2)]
}

func matchLocationNames(kayaLocation, mpArea string) bool {
	kayaLower := strings.ToLower(strings.TrimSpace(kayaLocation))
	mpLower := strings.ToLower(strings.TrimSpace(mpArea))
	return strings.Contains(mpLower, kayaLower) || strings.Contains(kayaLower, mpLower)
}

func calculateGPSDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func calculateMatchConfidence(nameSim float64, locationMatch bool, distanceKM *float64) float64 {
	confidence := nameSim * 0.7
	if locationMatch {
		confidence += 0.2
	}
	if distanceKM != nil && *distanceKM < 5.0 {
		confidence += 0.1 * (1.0 - (*distanceKM / 5.0))
	}
	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

func determineMatchType(nameSim float64, locationMatch bool, distanceKM *float64) string {
	if nameSim == 1.0 {
		return "exact_name"
	}
	if nameSim >= 0.9 && locationMatch {
		return "fuzzy_name_location"
	}
	if nameSim >= 0.85 {
		return "fuzzy_name"
	}
	if locationMatch && distanceKM != nil && *distanceKM < 1.0 {
		return "location_gps_proximity"
	}
	if locationMatch {
		return "location_name"
	}
	return "low_confidence"
}

func isCompatibleMatch(kayaClimbType, kayaGrade, mpRouteType, mpRating string) bool {
	kayaDiscipline := classifyKayaDiscipline(kayaClimbType, kayaGrade)
	mpDiscipline := classifyMPDiscipline(mpRouteType, mpRating)
	if kayaDiscipline != "" && mpDiscipline != "" && kayaDiscipline != mpDiscipline {
		return false
	}
	if kayaDiscipline == "boulder" {
		if !containsToken(mpRouteType, "boulder") {
			return false
		}
		if containsAnyToken(mpRouteType, []string{"ice", "mixed", "snow", "alpine"}) {
			return false
		}
	}
	kayaFamily := gradeFamily(kayaGrade)
	mpFamily := gradeFamily(mpRating)
	if kayaFamily != "" && mpFamily != "" && kayaFamily != mpFamily {
		return false
	}
	return true
}

func classifyKayaDiscipline(climbType, grade string) string {
	ct := strings.ToLower(strings.TrimSpace(climbType))
	switch {
	case strings.Contains(ct, "boulder"):
		return "boulder"
	case strings.Contains(ct, "sport"), strings.Contains(ct, "trad"), strings.Contains(ct, "route"):
		return "route"
	}
	switch gradeFamily(grade) {
	case "v":
		return "boulder"
	case "yds":
		return "route"
	case "wi", "mixed", "aid":
		return "ice"
	default:
		return ""
	}
}

func classifyMPDiscipline(routeType, rating string) string {
	rt := strings.ToLower(strings.TrimSpace(routeType))
	if containsToken(rt, "boulder") {
		return "boulder"
	}
	if containsAnyToken(rt, []string{"ice", "mixed", "snow", "alpine"}) {
		return "ice"
	}
	if containsAnyToken(rt, []string{"sport", "trad", "tr", "top rope", "aid"}) {
		return "route"
	}
	switch gradeFamily(rating) {
	case "v":
		return "boulder"
	case "yds":
		return "route"
	case "wi", "mixed", "aid":
		return "ice"
	default:
		return ""
	}
}

func gradeFamily(grade string) string {
	g := strings.ToUpper(strings.TrimSpace(grade))
	if g == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(g, "V"):
		return "v"
	case strings.HasPrefix(g, "WI"), strings.HasPrefix(g, "AI"):
		return "wi"
	case strings.HasPrefix(g, "M"):
		return "mixed"
	case strings.HasPrefix(g, "A"), strings.HasPrefix(g, "C"):
		return "aid"
	case strings.HasPrefix(g, "5.") || strings.HasPrefix(g, "5"):
		return "yds"
	default:
		return ""
	}
}

func containsToken(value, token string) bool {
	v := strings.ToLower(value)
	t := strings.ToLower(token)
	parts := strings.FieldsFunc(v, func(r rune) bool {
		return r == ',' || r == '/' || r == ';' || r == '|'
	})
	for _, p := range parts {
		if strings.TrimSpace(p) == t {
			return true
		}
	}
	return false
}

func containsAnyToken(value string, tokens []string) bool {
	for _, token := range tokens {
		if containsToken(value, token) {
			return true
		}
	}
	return false
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func createMonitoringDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "require"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	return sql.Open("postgres", connStr)
}
