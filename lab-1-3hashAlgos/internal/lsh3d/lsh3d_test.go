package lsh3d

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// TestIndexRejectsInvalidConfig проверяет, что некорректная конфигурация возвращает ошибку.
func TestIndexRejectsInvalidConfig(t *testing.T) {
	cases := []struct {
		cfg Config
		msg string
	}{
		{Config{NumTables: 0, NumFuncs: 3, BandWidth: 5}, "zero tables"},
		{Config{NumTables: 5, NumFuncs: 0, BandWidth: 5}, "zero funcs"},
		{Config{NumTables: 5, NumFuncs: 3, BandWidth: 0}, "zero bandwidth"},
		{Config{NumTables: 5, NumFuncs: 3, BandWidth: -1}, "negative bandwidth"},
	}
	for _, tc := range cases {
		_, err := NewIndex(tc.cfg)
		if err == nil {
			t.Errorf("expected error for %s", tc.msg)
		}
	}
}

// TestAddAndQueryFindsNearDuplicates добавляет базовые точки и ближайшие дубли
// (оригинал + гауссов шум σ=0.5), затем проверяет, что Query находит оригинал.
func TestAddAndQueryFindsNearDuplicates(t *testing.T) {
	for _, seed := range []int64{1, 7, 42, 99, 2026} {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			idx, err := NewIndex(DefaultConfig())
			if err != nil {
				t.Fatalf("NewIndex: %v", err)
			}
			const N = 500
			base := genPoints(N, seed)
			for _, p := range base {
				idx.Add(p)
			}

			rng := rand.New(rand.NewSource(seed + 1000))
			const M = 20
			for j := 0; j < M; j++ {
				orig := base[j]
				dup := nearDup(orig, 0.5, rng, N+j)
				idx.Add(dup)

				cands := idx.Query(dup)
				found := false
				for _, c := range cands {
					if c.ID == orig.ID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("seed=%d j=%d: near-duplicate of point %d not found (dist=%.2f), got %d candidates",
						seed, j, orig.ID, dist3D(orig, dup), len(cands))
				}
			}
		})
	}
}

// TestFullScanDuplicatesFindsEmbeddedPairs строит индекс из 1000 точек,
// вставляет 8 ближайших дублей и проверяет, что FullScanDuplicates их находит.
func TestFullScanDuplicatesFindsEmbeddedPairs(t *testing.T) {
	idx, err := NewIndex(DefaultConfig())
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}

	const N = 1000
	base := genPoints(N, 42)
	for _, p := range base {
		idx.Add(p)
	}

	rng := rand.New(rand.NewSource(777))
	type origDup struct{ origID, dupID int }
	pairs := make([]origDup, 8)
	for j := range pairs {
		orig := base[j*10]
		dup := nearDup(orig, 0.3, rng, N+j)
		idx.Add(dup)
		pairs[j] = origDup{orig.ID, dup.ID}
	}

	found := idx.FullScanDuplicates(DefaultConfig().BandWidth)
	foundSet := make(map[[2]int]bool)
	for _, p := range found {
		k := [2]int{p.ID1, p.ID2}
		if k[0] > k[1] {
			k[0], k[1] = k[1], k[0]
		}
		foundSet[k] = true
	}

	for _, od := range pairs {
		k := [2]int{od.origID, od.dupID}
		if k[0] > k[1] {
			k[0], k[1] = k[1], k[0]
		}
		if !foundSet[k] {
			t.Errorf("near-duplicate pair (%d, %d) not found in FullScanDuplicates", od.origID, od.dupID)
		}
	}
}

// TestCountMatchesAdded проверяет, что Count возвращает число добавленных точек.
func TestCountMatchesAdded(t *testing.T) {
	idx, _ := NewIndex(DefaultConfig())
	pts := genPoints(50, 1)
	for _, p := range pts {
		idx.Add(p)
	}
	if idx.Count() != len(pts) {
		t.Errorf("Count() = %d, want %d", idx.Count(), len(pts))
	}
}

// FuzzIndexNoPanic проверяет, что произвольные координаты не вызывают паник.
func FuzzIndexNoPanic(f *testing.F) {
	f.Add(0.0, 0.0, 0.0)
	f.Add(100.0, 100.0, 100.0)
	f.Add(-50.0, 25.0, 75.0)
	f.Add(float64(math.MaxFloat32), float64(math.MaxFloat32), float64(math.MaxFloat32))

	f.Fuzz(func(t *testing.T, x, y, z float64) {
		if math.IsNaN(x) || math.IsNaN(y) || math.IsNaN(z) ||
			math.IsInf(x, 0) || math.IsInf(y, 0) || math.IsInf(z, 0) {
			return
		}
		idx, err := NewIndex(DefaultConfig())
		if err != nil {
			t.Fatalf("NewIndex: %v", err)
		}
		p := Point3D{X: x, Y: y, Z: z, ID: 1}
		idx.Add(p)
		_ = idx.Query(p)
		_ = idx.FullScanDuplicates(1.0)
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func genPoints(n int, seed int64) []Point3D {
	rng := rand.New(rand.NewSource(seed))
	pts := make([]Point3D, n)
	for i := range pts {
		pts[i] = Point3D{
			X:  rng.Float64() * 100,
			Y:  rng.Float64() * 100,
			Z:  rng.Float64() * 100,
			ID: i,
		}
	}
	return pts
}

func nearDup(p Point3D, sigma float64, rng *rand.Rand, id int) Point3D {
	return Point3D{
		X:  p.X + rng.NormFloat64()*sigma,
		Y:  p.Y + rng.NormFloat64()*sigma,
		Z:  p.Z + rng.NormFloat64()*sigma,
		ID: id,
	}
}
