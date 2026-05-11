package rock_temp

import "math"

// MinSunElevDeg is the sun elevation guard threshold (degrees). Below
// this elevation we treat direct-normal irradiance as zero to avoid the
// 1/sin(elev) blow-up near the horizon.
const MinSunElevDeg = 5.0

func deg2rad(d float64) float64 { return d * math.Pi / 180.0 }

// DNIFromHorizontal converts horizontal-surface direct radiation to
// direct-normal irradiance (DNI) at the supplied sun elevation
// (degrees). Returns 0 when the sun is at or below MinSunElevDeg.
func DNIFromHorizontal(directHoriz, sunElevDeg float64) float64 {
	if sunElevDeg <= MinSunElevDeg {
		return 0
	}
	return directHoriz / math.Sin(deg2rad(sunElevDeg))
}

// GeometricFactor returns the cosine of the angle between the sun
// vector and the outward normal of a tilted face, clamped to [0, +∞).
//
// All angles are in degrees:
//   - sunAzDeg, sunElevDeg: sun azimuth (from north, clockwise) and
//     elevation above the horizon.
//   - faceAspectDeg: face azimuth (direction the outward normal points,
//     0 = north, 90 = east, 180 = south, 270 = west).
//   - faceDipDeg: angle between the face and horizontal:
//     0 = flat slab pointing up, 90 = vertical wall, 110 = overhang.
//
// Formula (general tilted plane):
//
//	geom = max(0, sin(elev)·cos(dip) + cos(elev)·sin(dip)·cos(sunAz - faceAspect))
//
// Sun below or at the horizon returns 0.
func GeometricFactor(sunAzDeg, sunElevDeg, faceAspectDeg, faceDipDeg float64) float64 {
	if sunElevDeg <= 0 {
		return 0
	}
	elev := deg2rad(sunElevDeg)
	dip := deg2rad(faceDipDeg)
	azDiff := deg2rad(sunAzDeg - faceAspectDeg)
	g := math.Sin(elev)*math.Cos(dip) + math.Cos(elev)*math.Sin(dip)*math.Cos(azDiff)
	if g < 0 {
		return 0
	}
	return g
}

// SkyViewFactor returns the fraction of the sky hemisphere visible from
// a face with the given dip (degrees). For a flat slab (dip=0) this is
// 1.0; for a vertical wall (dip=90) it is 0.5.
func SkyViewFactor(faceDipDeg float64) float64 {
	return (1.0 + math.Cos(deg2rad(faceDipDeg))) / 2.0
}

// FaceIrradiance combines direct-normal irradiance, diffuse horizontal
// irradiance, geometric factor, sky view factor, and tree canopy
// attenuation into the effective irradiance on the rock face (W/m²).
//
// Tree canopy attenuates direct radiation more strongly than diffuse:
//
//	direct      = dni * geom * (1 - treeFraction*0.75)
//	diffusePart = diffuseHoriz * skyView * (1 - treeFraction*0.5)
//
// treeFraction is in [0, 1], NOT 0..100. Returns max(0, direct + diffusePart).
func FaceIrradiance(dni, diffuseHoriz, geom, skyView, treeFraction float64) float64 {
	if treeFraction < 0 {
		treeFraction = 0
	} else if treeFraction > 1 {
		treeFraction = 1
	}
	direct := dni * geom * (1 - treeFraction*0.75)
	diffusePart := diffuseHoriz * skyView * (1 - treeFraction*0.5)
	v := direct + diffusePart
	if v < 0 {
		return 0
	}
	return v
}

// SplitShortwave is a fallback splitter used when only total shortwave
// radiation is available (e.g., older cached rows where direct/diffuse
// columns are 0). It estimates the diffuse fraction from cloud cover:
//
//	cloud < 20% → diffuse ≈ 15% of total
//	cloud > 80% → diffuse ≈ 100% of total
//	linear interpolation between those endpoints
//
// Returns (directHoriz, diffuseHoriz). Returns (0, 0) for non-positive
// totals.
func SplitShortwave(totalHoriz, cloudFractionPct float64) (float64, float64) {
	if totalHoriz <= 0 {
		return 0, 0
	}
	c := cloudFractionPct
	if c < 0 {
		c = 0
	} else if c > 100 {
		c = 100
	}
	var diffuseFrac float64
	switch {
	case c < 20:
		diffuseFrac = 0.15
	case c > 80:
		diffuseFrac = 1.0
	default:
		// Linearly interpolate from (20, 0.15) to (80, 1.0).
		diffuseFrac = 0.15 + (c-20)/(80-20)*(1.0-0.15)
	}
	diffuse := totalHoriz * diffuseFrac
	direct := totalHoriz * (1 - diffuseFrac)
	return direct, diffuse
}
