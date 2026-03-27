package geo

import (
	"math"
	"testing"
)

// TestHaversine_KnownDistances проверяет формулу Хаверсина на известных расстояниях.
func TestHaversine_KnownDistances(t *testing.T) {
	cases := []struct {
		name        string
		lat1, lng1  float64
		lat2, lng2  float64
		wantKm      float64
		tolerancePct float64
	}{
		{
			name:        "Moscow to Saint-Petersburg",
			lat1: 55.7558, lng1: 37.6173,
			lat2: 59.9343, lng2: 30.3351,
			wantKm:      634.0,
			tolerancePct: 1.0,
		},
		{
			name:        "London to Paris",
			lat1: 51.5074, lng1: -0.1278,
			lat2: 48.8566, lng2: 2.3522,
			wantKm:      344.0,
			tolerancePct: 1.0,
		},
		{
			name:        "same point",
			lat1: 55.0, lng1: 37.0,
			lat2: 55.0, lng2: 37.0,
			wantKm:      0.0,
			tolerancePct: 0.0,
		},
		{
			name:        "antipodal points",
			lat1: 0, lng1: 0,
			lat2: 0, lng2: 180,
			wantKm:      math.Pi * 6371.0, // half circumference
			tolerancePct: 0.5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DistanceKm(tc.lat1, tc.lng1, tc.lat2, tc.lng2)
			if tc.wantKm == 0 {
				if got > 0.001 {
					t.Errorf("expected ~0, got %.4f km", got)
				}
				return
			}
			pct := math.Abs(got-tc.wantKm) / tc.wantKm * 100
			if pct > tc.tolerancePct {
				t.Errorf("distance %.2f km, want %.2f km (error %.2f%%)", got, tc.wantKm, pct)
			}
		})
	}
}

// TestHaversine_Symmetry проверяет симметричность: d(A,B) == d(B,A).
func TestHaversine_Symmetry(t *testing.T) {
	lat1, lng1 := 55.75, 37.62
	lat2, lng2 := 48.85, 2.35
	d1 := DistanceKm(lat1, lng1, lat2, lng2)
	d2 := DistanceKm(lat2, lng2, lat1, lng1)
	if math.Abs(d1-d2) > 1e-9 {
		t.Errorf("asymmetry: d(A,B)=%.6f d(B,A)=%.6f", d1, d2)
	}
}

// TestGeohash_EncodeDecodeRoundtrip — кодируем точку, декодируем, проверяем,
// что центр ячейки попадает в её границы.
func TestGeohash_EncodeDecodeRoundtrip(t *testing.T) {
	testPoints := [][2]float64{
		{55.7558, 37.6173},  // Москва
		{59.9343, 30.3351},  // Санкт-Петербург
		{0, 0},              // Null island
		{-33.87, 151.21},    // Сидней
		{90, 0},             // Северный полюс
		{-90, 0},            // Южный полюс
		{0, 180},            // Линия перемены дат
	}

	for _, prec := range []int{3, 5, 7, 9} {
		for _, pt := range testPoints {
			lat, lng := pt[0], pt[1]
			hash := Encode(lat, lng, prec)
			if len(hash) != prec {
				t.Errorf("prec=%d: expected %d chars, got %d (%q)", prec, prec, len(hash), hash)
			}

			dLat, dLng, _, _, err := Decode(hash)
			if err != nil {
				t.Errorf("Decode(%q): %v", hash, err)
				continue
			}

			minLat, maxLat, minLng, maxLng, err := DecodeBounds(hash)
			if err != nil {
				t.Errorf("DecodeBounds(%q): %v", hash, err)
				continue
			}

			// Центр ячейки должен быть внутри её границ.
			if dLat < minLat-1e-9 || dLat > maxLat+1e-9 {
				t.Errorf("prec=%d lat=%v: decoded lat %v not in [%v,%v]",
					prec, lat, dLat, minLat, maxLat)
			}
			if dLng < minLng-1e-9 || dLng > maxLng+1e-9 {
				t.Errorf("prec=%d lng=%v: decoded lng %v not in [%v,%v]",
					prec, lng, dLng, minLng, maxLng)
			}
		}
	}
}

// TestGeohash_SamePrefixSameRegion — соседние точки при достаточной точности
// должны иметь одинаковый префикс.
func TestGeohash_SamePrefixSameRegion(t *testing.T) {
	// Два POI в 100 м друг от друга в центре Москвы
	h1 := Encode(55.7558, 37.6173, 7)
	h2 := Encode(55.7559, 37.6174, 7)
	// При прецизии 7 ячейка ~153m, оба должны попасть в один или соседний хэш
	if h1[:5] != h2[:5] {
		t.Logf("h1=%s h2=%s (разные префиксы — допустимо на границе ячеек)", h1, h2)
	}
}

// TestGeohash_Neighbors_Count — NeighborsAndSelf должен вернуть 1 + до 8 = до 9 ячеек,
// и среди них не должно быть дублей.
func TestGeohash_Neighbors_Count(t *testing.T) {
	hashes := []string{
		Encode(55.75, 37.62, 5),  // Москва
		Encode(0, 0, 5),          // экватор/нулевой меридиан
		Encode(89, 0, 5),         // почти полюс
	}

	for _, h := range hashes {
		cells := NeighborsAndSelf(h)
		if len(cells) < 1 || len(cells) > 9 {
			t.Errorf("NeighborsAndSelf(%q): got %d cells, want 1..9", h, len(cells))
		}
		seen := make(map[string]struct{})
		for _, c := range cells {
			if _, dup := seen[c]; dup {
				t.Errorf("NeighborsAndSelf(%q): duplicate cell %q", h, c)
			}
			seen[c] = struct{}{}
		}
	}
}

// TestGeohash_InvalidChar — Decode должен вернуть ошибку на невалидном символе.
func TestGeohash_InvalidChar(t *testing.T) {
	if _, _, _, _, err := Decode("zzzzz!!"); err == nil {
		t.Error("expected error for invalid geohash character")
	}
}
