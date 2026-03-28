package geoindex_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"siaod-hw2/internal/brute"
	"siaod-hw2/internal/gen"
	"siaod-hw2/internal/geo"
	"siaod-hw2/internal/geoindex"
	"siaod-hw2/internal/kdtree"
)

// allSearchers возвращает все реализации Searcher для параметрических тестов.
func allSearchers() []struct {
	name string
	new  func() geoindex.Searcher
}{
	return []struct {
		name string
		new  func() geoindex.Searcher
	}{
		{"geohash-p5", func() geoindex.Searcher { return geoindex.New(5) }},
		{"geohash-p6", func() geoindex.Searcher { return geoindex.New(6) }},
		{"kdtree", func() geoindex.Searcher { return kdtree.New() }},
		{"brute", func() geoindex.Searcher { return brute.New() }},
	}
}

// TestSearcherRandomInsertAndQuery запускает серию случайных Insert + FindNearby
// на 5 различных seed-ах. Корректность проверяется через brute-force оракул.
func TestSearcherRandomInsertAndQuery(t *testing.T) {
	seeds := []int64{1, 7, 42, 99, 2026}
	radii := []float64{5.0, 20.0, 50.0}
	queryLat, queryLng := 55.75, 37.62 // Москва

	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			r := gen.NewRand(seed)
			const n = 1000
			pts := gen.GeneratePoints(r, n)

			ref := brute.New()
			for _, p := range pts {
				ref.Insert(p)
			}

			for _, radius := range radii {
				expected := ref.FindNearby(queryLat, queryLng, radius)
				expectedIDs := make(map[string]struct{}, len(expected))
				for _, res := range expected {
					expectedIDs[res.Point.ID] = struct{}{}
				}

				for _, tc := range allSearchers() {
					tc := tc
					t.Run(fmt.Sprintf("radius=%.0f/%s", radius, tc.name), func(t *testing.T) {
						s := tc.new()
						for _, p := range pts {
							s.Insert(p)
						}
						if s.Count() != n {
							t.Fatalf("Count() = %d, want %d", s.Count(), n)
						}

						got := s.FindNearby(queryLat, queryLng, radius)

						// Все возвращённые точки должны быть в радиусе.
						for _, res := range got {
							d := geo.DistanceKm(queryLat, queryLng, res.Point.Lat, res.Point.Lng)
							if d > radius+0.001 {
								t.Errorf("%s: result %q dist=%.3f > radius=%.3f",
									tc.name, res.Point.ID, d, radius)
							}
						}

						// Сравниваем с brute-force оракулом.
						gotIDs := make(map[string]struct{}, len(got))
						for _, res := range got {
							gotIDs[res.Point.ID] = struct{}{}
						}
						missing := 0
						for id := range expectedIDs {
							if _, ok := gotIDs[id]; !ok {
								missing++
							}
						}
						allowedMiss := 0
						if strings.HasPrefix(tc.name, "geohash") {
							allowedMiss = len(expectedIDs)/50 + 1
						}
						if missing > allowedMiss {
							t.Errorf("seed=%d radius=%.0f %s: missed %d/%d points (allowed %d)",
								seed, radius, tc.name, missing, len(expectedIDs), allowedMiss)
						}
					})
				}
			}
		})
	}
}

// TestSearcherNearbyPointMustBeFound — точка, вставленная в непосредственной
// близости от запроса, обязана быть найдена всеми реализациями.
func TestSearcherNearbyPointMustBeFound(t *testing.T) {
	queryLat, queryLng := 59.93, 30.32 // Петербург
	radius := 5.0

	r := gen.NewRand(2)
	noise := gen.GeneratePoints(r, 200)
	target := gen.NearbyPoint(r, queryLat, queryLng, 1.0, 9999)

	for _, tc := range allSearchers() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := tc.new()
			for _, p := range noise {
				s.Insert(p)
			}
			s.Insert(target)

			results := s.FindNearby(queryLat, queryLng, radius)
			found := false
			for _, res := range results {
				if res.Point.ID == target.ID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s: target %q not found within %.1f km", tc.name, target.ID, radius)
			}
		})
	}
}

// TestSearcherResultsAreSortedByDistance — FindNearby должен возвращать результаты
// в порядке возрастания расстояния.
func TestSearcherResultsAreSortedByDistance(t *testing.T) {
	r := gen.NewRand(3)
	pts := gen.GeneratePoints(r, 500)

	for _, tc := range allSearchers() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := tc.new()
			for _, p := range pts {
				s.Insert(p)
			}
			results := s.FindNearby(55.75, 37.62, 30.0)
			for i := 1; i < len(results); i++ {
				if results[i].Distance < results[i-1].Distance-1e-9 {
					t.Errorf("%s: results not sorted: [%d]=%.4f < [%d]=%.4f",
						tc.name, i, results[i].Distance, i-1, results[i-1].Distance)
				}
			}

			// Reported distance должно совпадать с вычисленным.
			for _, res := range results {
				d := geo.DistanceKm(55.75, 37.62, res.Point.Lat, res.Point.Lng)
				if math.Abs(d-res.Distance) > 0.01 {
					t.Errorf("%s: %q reported=%.4f computed=%.4f",
						tc.name, res.Point.ID, res.Distance, d)
				}
			}
		})
	}
}

// TestSearcherEmptyIndex — поиск в пустом индексе всегда возвращает пустой срез.
func TestSearcherEmptyIndex(t *testing.T) {
	for _, tc := range allSearchers() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := tc.new()
			if got := s.FindNearby(55.75, 37.62, 10.0); len(got) != 0 {
				t.Errorf("%s: empty index returned %d results", tc.name, len(got))
			}
			if s.Count() != 0 {
				t.Errorf("%s: Count() = %d, want 0", tc.name, s.Count())
			}
		})
	}
}

// FuzzSearcherAgainstBrute — фаззинг: случайные точки, случайный запрос.
// Проверяем, что ни один алгоритм не паникует и kd-tree точно совпадает с brute.
func FuzzSearcherAgainstBrute(f *testing.F) {
	f.Add(int64(1), float64(55.75), float64(37.62), float64(10.0), uint16(100))
	f.Add(int64(42), float64(59.93), float64(30.32), float64(5.0), uint16(50))
	f.Add(int64(2026), float64(48.85), float64(2.35), float64(25.0), uint16(200))

	f.Fuzz(func(t *testing.T, seed int64, qLat, qLng, radius float64, count uint16) {
		// Нормализуем входные данные.
		for qLat < -90 {
			qLat += 180
		}
		for qLat > 90 {
			qLat -= 180
		}
		for qLng < -180 {
			qLng += 360
		}
		for qLng > 180 {
			qLng -= 360
		}
		if radius < 0 {
			radius = -radius
		}
		if radius > 500 {
			radius = 500
		}
		n := int(count%300) + 1

		r := gen.NewRand(seed)
		pts := gen.GeneratePoints(r, n)

		ref := brute.New()
		kd := kdtree.New()
		for _, p := range pts {
			ref.Insert(p)
			kd.Insert(p)
		}

		bruteRes := ref.FindNearby(qLat, qLng, radius)
		kdRes := kd.FindNearby(qLat, qLng, radius)

		// kd-tree не должно пропускать точки, найденные brute-force.
		kdIDs := make(map[string]struct{}, len(kdRes))
		for _, res := range kdRes {
			kdIDs[res.Point.ID] = struct{}{}
		}
		for _, res := range bruteRes {
			if _, ok := kdIDs[res.Point.ID]; !ok {
				t.Fatalf("kdtree missed point %q (dist=%.4f, radius=%.4f)",
					res.Point.ID, res.Distance, radius)
			}
		}
	})
}
