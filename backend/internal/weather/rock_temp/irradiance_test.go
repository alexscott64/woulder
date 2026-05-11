package rock_temp

import (
	"math"
	"testing"
)

const eps = 1e-6

func approxEq(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func TestDNIFromHorizontal(t *testing.T) {
	// 30° elevation, 500 W/m² horizontal → DNI = 500 / sin(30°) = 1000.
	got := DNIFromHorizontal(500, 30)
	if !approxEq(got, 1000, 1e-6) {
		t.Errorf("DNI 30°: got %.4f want 1000", got)
	}
	// Below guard.
	if got := DNIFromHorizontal(500, 5); got != 0 {
		t.Errorf("DNI elev=5°: got %.4f want 0", got)
	}
	if got := DNIFromHorizontal(500, 4); got != 0 {
		t.Errorf("DNI elev=4°: got %.4f want 0", got)
	}
	if got := DNIFromHorizontal(500, 0); got != 0 {
		t.Errorf("DNI elev=0°: got %.4f want 0", got)
	}
}

func TestGeometricFactor_VerticalSouthWall(t *testing.T) {
	// dip=90 (vertical), face_az=180 (south), sun_az=180, sun_elev=45
	// Formula: cos(45)*cos(0) = 0.7071...
	got := GeometricFactor(180, 45, 180, 90)
	if !approxEq(got, math.Cos(deg2rad(45)), 1e-6) {
		t.Errorf("south vertical wall, sun south 45°: got %.4f want ~0.707", got)
	}
}

func TestGeometricFactor_NorthWallSunSouth(t *testing.T) {
	// Sun behind face → clamped to 0.
	got := GeometricFactor(180, 45, 0, 90)
	if got != 0 {
		t.Errorf("north wall sun south: got %.4f want 0", got)
	}
}

func TestGeometricFactor_HorizontalSlab(t *testing.T) {
	// dip=0, formula reduces to sin(elev) regardless of azimuth.
	for _, az := range []float64{0, 90, 180, 270} {
		got := GeometricFactor(az, 30, 180, 0)
		want := math.Sin(deg2rad(30))
		if !approxEq(got, want, 1e-6) {
			t.Errorf("slab az=%.0f: got %.4f want %.4f", az, got, want)
		}
	}
}

func TestGeometricFactor_SunBelowHorizon(t *testing.T) {
	if got := GeometricFactor(180, -1, 180, 90); got != 0 {
		t.Errorf("sun below horizon: got %.4f want 0", got)
	}
	if got := GeometricFactor(180, 0, 180, 90); got != 0 {
		t.Errorf("sun at horizon: got %.4f want 0", got)
	}
}

func TestGeometricFactor_OverhangFormula(t *testing.T) {
	// dip=110, sun_elev=80, sun_az=face_az=180
	// geom = sin(80)cos(110) + cos(80)sin(110)*cos(0)
	got := GeometricFactor(180, 80, 180, 110)
	want := math.Sin(deg2rad(80))*math.Cos(deg2rad(110)) + math.Cos(deg2rad(80))*math.Sin(deg2rad(110))
	if want < 0 {
		want = 0
	}
	if !approxEq(got, want, 1e-6) {
		t.Errorf("overhang: got %.4f want %.4f", got, want)
	}
}

func TestSkyViewFactor(t *testing.T) {
	if !approxEq(SkyViewFactor(90), 0.5, 1e-9) {
		t.Errorf("vertical sky view should be 0.5, got %.4f", SkyViewFactor(90))
	}
	if !approxEq(SkyViewFactor(0), 1.0, 1e-9) {
		t.Errorf("horizontal sky view should be 1.0, got %.4f", SkyViewFactor(0))
	}
}

func TestFaceIrradiance_TreeAttenuation(t *testing.T) {
	// 80% canopy: direct cuts to 1 - 0.8*0.75 = 0.40, diffuse to 1 - 0.8*0.5 = 0.60.
	// Use dni=1000, geom=1, diffuse=200, skyView=0.5, tree=0.8.
	got := FaceIrradiance(1000, 200, 1, 0.5, 0.8)
	wantDirect := 1000 * 1 * 0.40
	wantDiffuse := 200 * 0.5 * 0.60
	want := wantDirect + wantDiffuse
	if !approxEq(got, want, 1e-9) {
		t.Errorf("face irradiance with tree: got %.4f want %.4f", got, want)
	}

	// Zero tree.
	got2 := FaceIrradiance(1000, 200, 1, 0.5, 0)
	want2 := 1000.0 + 200.0*0.5
	if !approxEq(got2, want2, 1e-9) {
		t.Errorf("face irradiance no tree: got %.4f want %.4f", got2, want2)
	}
}

func TestFaceIrradiance_ClampsNegative(t *testing.T) {
	if got := FaceIrradiance(-100, -100, 1, 1, 0); got != 0 {
		t.Errorf("clamp negative: got %.4f want 0", got)
	}
}

func TestSplitShortwave(t *testing.T) {
	// Clear (cloud=10): diffuse=15%.
	d, df := SplitShortwave(1000, 10)
	if !approxEq(df, 150, 1e-6) || !approxEq(d, 850, 1e-6) {
		t.Errorf("clear split: direct=%.2f diffuse=%.2f", d, df)
	}
	// Overcast (cloud=90): diffuse=100%.
	d, df = SplitShortwave(800, 90)
	if !approxEq(df, 800, 1e-6) || !approxEq(d, 0, 1e-6) {
		t.Errorf("overcast split: direct=%.2f diffuse=%.2f", d, df)
	}
	// Mid (cloud=50): linear interp from 0.15 to 1.0.
	d, df = SplitShortwave(1000, 50)
	expectedFrac := 0.15 + float64(50-20)/float64(80-20)*(1.0-0.15) // 0.575
	if !approxEq(df, 1000*expectedFrac, 1e-6) {
		t.Errorf("mid cloud diffuse: got %.4f want %.4f", df, 1000*expectedFrac)
	}
	// Zero / negative total.
	if d, df = SplitShortwave(0, 50); d != 0 || df != 0 {
		t.Errorf("zero total: got %.2f, %.2f", d, df)
	}
	if d, df = SplitShortwave(-5, 50); d != 0 || df != 0 {
		t.Errorf("negative total: got %.2f, %.2f", d, df)
	}
}
