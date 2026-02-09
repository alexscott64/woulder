package sun

import (
	"math"
	"time"
)

// Position represents the sun's position in the sky
type Position struct {
	Azimuth   float64 // Degrees from north (0-360)
	Elevation float64 // Degrees above horizon (-90 to 90)
}

// HourlyPositions calculates sun positions for each hour over the specified duration
func HourlyPositions(lat, lon float64, startTime time.Time, hours int) []Position {
	positions := make([]Position, hours)
	for i := 0; i < hours; i++ {
		t := startTime.Add(time.Duration(i) * time.Hour)
		positions[i] = Calculate(lat, lon, t)
	}
	return positions
}

// Calculate computes the sun's position for a given location and time
// Based on simplified solar position algorithm (accurate to ~0.01 degrees)
func Calculate(lat, lon float64, t time.Time) Position {
	// Convert to radians
	latRad := lat * math.Pi / 180.0
	lonRad := lon * math.Pi / 180.0

	// Calculate Julian day
	jd := julianDay(t)

	// Calculate time since J2000.0
	n := jd - 2451545.0

	// Mean longitude of the sun
	L := math.Mod(280.460+0.9856474*n, 360.0)

	// Mean anomaly
	g := math.Mod(357.528+0.9856003*n, 360.0)
	gRad := g * math.Pi / 180.0

	// Ecliptic longitude
	lambda := L + 1.915*math.Sin(gRad) + 0.020*math.Sin(2*gRad)
	lambdaRad := lambda * math.Pi / 180.0

	// Obliquity of ecliptic
	epsilon := 23.439 - 0.0000004*n
	epsilonRad := epsilon * math.Pi / 180.0

	// Right ascension
	ra := math.Atan2(math.Cos(epsilonRad)*math.Sin(lambdaRad), math.Cos(lambdaRad))

	// Declination
	dec := math.Asin(math.Sin(epsilonRad) * math.Sin(lambdaRad))

	// Calculate Julian centuries from J2000.0
	T := n / 36525.0

	// Greenwich mean sidereal time (more accurate formula)
	gmst := 280.46061837 + 360.98564736629*n + 0.000387933*T*T - T*T*T/38710000.0
	gmst = math.Mod(gmst, 360.0)
	if gmst < 0 {
		gmst += 360.0
	}
	gmstRad := gmst * math.Pi / 180.0

	// Local sidereal time
	lst := gmstRad + lonRad

	// Hour angle
	ha := lst - ra

	// Calculate elevation (altitude)
	elevation := math.Asin(
		math.Sin(latRad)*math.Sin(dec) +
			math.Cos(latRad)*math.Cos(dec)*math.Cos(ha),
	)

	// Calculate azimuth (measured clockwise from north)
	// Using standard astronomical formula
	azimuthX := math.Sin(ha)
	azimuthY := math.Cos(ha)*math.Sin(latRad) - math.Tan(dec)*math.Cos(latRad)
	azimuth := math.Atan2(azimuthX, azimuthY)

	// Convert to degrees (0° = North, 90° = East, 180° = South, 270° = West)
	elevationDeg := elevation * 180.0 / math.Pi
	azimuthDeg := math.Mod(azimuth*180.0/math.Pi+180.0, 360.0)

	return Position{
		Azimuth:   azimuthDeg,
		Elevation: elevationDeg,
	}
}

// julianDay calculates the Julian day number for a given time
func julianDay(t time.Time) float64 {
	utc := t.UTC()
	year := utc.Year()
	month := int(utc.Month())
	day := utc.Day()
	hour := utc.Hour()
	minute := utc.Minute()
	second := utc.Second()

	// Adjust for January/February
	if month <= 2 {
		year--
		month += 12
	}

	// Calculate Julian day
	a := year / 100
	b := 2 - a + a/4

	jd := float64(int(365.25*float64(year+4716))) +
		float64(int(30.6001*float64(month+1))) +
		float64(day) +
		float64(b) - 1524.5

	// Add time of day
	dayFraction := (float64(hour) + float64(minute)/60.0 + float64(second)/3600.0) / 24.0
	jd += dayFraction

	return jd
}

// CalculateSunExposure calculates total sun exposure hours based on:
// - Boulder aspect (direction it faces)
// - Sun positions throughout the day
// - Tree coverage percentage
func CalculateSunExposure(lat, lon float64, aspect string, treeCoverage float64, startTime time.Time, hours int) float64 {
	positions := HourlyPositions(lat, lon, startTime, hours)

	// Convert aspect to degrees (0 = North, 90 = East, 180 = South, 270 = West)
	aspectDeg := aspectToDegrees(aspect)

	sunHours := 0.0
	for _, pos := range positions {
		// Only count if sun is above horizon
		if pos.Elevation <= 0 {
			continue
		}

		// Check if sun is hitting the face (within ±90° of aspect direction)
		azimuthDiff := math.Abs(angleDifference(pos.Azimuth, aspectDeg))
		if azimuthDiff > 90.0 {
			continue // Sun is behind the boulder
		}

		// Calculate exposure factor based on angle
		// Direct sun (0° difference) = 1.0, grazing (90° difference) = 0.0
		exposureFactor := math.Cos(azimuthDiff * math.Pi / 180.0)

		// Apply tree coverage reduction
		treeFactor := 1.0 - (treeCoverage / 100.0)

		// Add weighted sun hour
		sunHours += exposureFactor * treeFactor
	}

	return sunHours
}

// aspectToDegrees converts compass direction to degrees
func aspectToDegrees(aspect string) float64 {
	switch aspect {
	case "North", "N":
		return 0.0
	case "North-East", "NE", "Northeast":
		return 45.0
	case "East", "E":
		return 90.0
	case "South-East", "SE", "Southeast":
		return 135.0
	case "South", "S":
		return 180.0
	case "South-West", "SW", "Southwest":
		return 225.0
	case "West", "W":
		return 270.0
	case "North-West", "NW", "Northwest":
		return 315.0
	default:
		// Unknown aspect - assume south (best sun exposure)
		return 180.0
	}
}

// angleDifference calculates the smallest difference between two angles
func angleDifference(a1, a2 float64) float64 {
	diff := math.Abs(a1 - a2)
	if diff > 180.0 {
		diff = 360.0 - diff
	}
	return diff
}

// IsAboveHorizon checks if the sun is visible (elevation > 0)
func (p Position) IsAboveHorizon() bool {
	return p.Elevation > 0
}

// GetSunriseAndSunset estimates sunrise and sunset times for a given location and date
func GetSunriseAndSunset(lat, lon float64, date time.Time) (sunrise, sunset time.Time) {
	// Start at midnight UTC
	midnight := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	// Search every 30 minutes for 30 hours to handle cases where sunset crosses into next UTC day
	// (e.g., Seattle on June 21: sunset is ~04:11 UTC on June 22)
	var prevElevation float64
	var prevTime time.Time
	sunriseFound := false

	for i := 0; i < 60; i++ { // 60 intervals = 30 hours
		t := midnight.Add(time.Duration(i*30) * time.Minute)
		pos := Calculate(lat, lon, t)

		if i > 0 {
			// Sunrise: elevation crosses from negative to positive
			if prevElevation <= 0 && pos.Elevation > 0 && !sunriseFound {
				// Interpolate for better accuracy
				sunrise = interpolateTime(prevTime, t, prevElevation, pos.Elevation, 0)
				sunriseFound = true
			}

			// Sunset: elevation crosses from positive to negative
			if prevElevation > 0 && pos.Elevation <= 0 && sunriseFound {
				// Interpolate for better accuracy
				sunset = interpolateTime(prevTime, t, prevElevation, pos.Elevation, 0)
				break
			}
		}

		prevElevation = pos.Elevation
		prevTime = t
	}

	return sunrise, sunset
}

// interpolateTime linearly interpolates between two times to find when elevation crosses threshold
func interpolateTime(t1, t2 time.Time, elev1, elev2, threshold float64) time.Time {
	if elev1 == elev2 {
		return t1
	}
	// Linear interpolation: t = t1 + (threshold - elev1) / (elev2 - elev1) * (t2 - t1)
	fraction := (threshold - elev1) / (elev2 - elev1)
	duration := t2.Sub(t1)
	return t1.Add(time.Duration(float64(duration) * fraction))
}
