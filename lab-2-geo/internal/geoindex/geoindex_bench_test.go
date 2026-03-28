package geoindex_test

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"siaod-hw2/internal/brute"
	"siaod-hw2/internal/gen"
	"siaod-hw2/internal/geoindex"
	"siaod-hw2/internal/kdtree"
)

var defaultBenchmarkSizes = []int{1_000, 10_000, 50_000, 100_000, 500_000}

// BenchmarkInsert измеряет пропускную способность вставки N точек для всех реализаций.
func BenchmarkInsert(b *testing.B) {
	impls := []struct {
		name string
		new  func() geoindex.Searcher
	}{
		{"geohash-p5", func() geoindex.Searcher { return geoindex.New(5) }},
		{"geohash-p6", func() geoindex.Searcher { return geoindex.New(6) }},
		{"kdtree", func() geoindex.Searcher { return kdtree.New() }},
		{"brute", func() geoindex.Searcher { return brute.New() }},
	}

	for _, size := range benchmarkSizes(b) {
		r := gen.NewRand(int64(size))
		pts := gen.GeneratePoints(r, size)

		for _, impl := range impls {
			impl := impl
			b.Run(fmt.Sprintf("%s/size=%d", impl.name, size), func(b *testing.B) {
				runBatchBenchmark(b, size, func() {
					b.StopTimer()
					idx := impl.new()
					rand.Shuffle(len(pts), func(i, j int) { pts[i], pts[j] = pts[j], pts[i] })
					b.StartTimer()

					for _, p := range pts {
						idx.Insert(p)
					}
				})
			})
		}
	}
}

// BenchmarkFindNearby_VaryN — сравнение алгоритмов при разном размере индекса.
// Запрос: фиксированный радиус 10 км, центр Москвы.
func BenchmarkFindNearby_VaryN(b *testing.B) {
	const queryLat, queryLng = 55.75, 37.62
	const radius = 10.0

	for _, size := range benchmarkSizes(b) {
		size := size
		r := gen.NewRand(int64(size))
		pts := gen.GeneratePoints(r, size)

		impls := []struct {
			name string
			new  func([]geoindex.Point) geoindex.Searcher
		}{
			{"geohash-p5", func(pts []geoindex.Point) geoindex.Searcher {
				idx := geoindex.New(5)
				for _, p := range pts {
					idx.Insert(p)
				}
				return idx
			}},
			{"geohash-p6", func(pts []geoindex.Point) geoindex.Searcher {
				idx := geoindex.New(6)
				for _, p := range pts {
					idx.Insert(p)
				}
				return idx
			}},
			{"kdtree-balanced", func(pts []geoindex.Point) geoindex.Searcher {
				kt := kdtree.New()
				kt.BuildBalanced(pts)
				return kt
			}},
			{"kdtree-online", func(pts []geoindex.Point) geoindex.Searcher {
				kt := kdtree.New()
				for _, p := range pts {
					kt.Insert(p)
				}
				return kt
			}},
			{"brute", func(pts []geoindex.Point) geoindex.Searcher {
				br := brute.New()
				for _, p := range pts {
					br.Insert(p)
				}
				return br
			}},
		}

		for _, impl := range impls {
			impl := impl
			b.Run(fmt.Sprintf("%s/size=%d", impl.name, size), func(b *testing.B) {
				idx := impl.new(pts)

				runBatchBenchmark(b, b.N, func() {
					_ = idx.FindNearby(queryLat, queryLng, radius)
				})
			})
		}
	}
}

// BenchmarkFindNearby_VaryRadius — влияние радиуса поиска на время запроса.
func BenchmarkFindNearby_VaryRadius(b *testing.B) {
	const n = 50_000
	r := gen.NewRand(42)
	pts := gen.GeneratePoints(r, n)

	ghIdx := geoindex.New(5)
	ktIdx := kdtree.New()
	brIdx := brute.New()
	for _, p := range pts {
		ghIdx.Insert(p)
		ktIdx.Insert(p)
		brIdx.Insert(p)
	}

	radii := []float64{1.0, 5.0, 20.0, 50.0, 100.0}
	for _, rad := range radii {
		rad := rad
		for _, tc := range []struct {
			name string
			idx  geoindex.Searcher
		}{
			{"geohash", ghIdx},
			{"kdtree", ktIdx},
			{"brute", brIdx},
		} {
			tc := tc
			b.Run(fmt.Sprintf("%s/r=%.0fkm", tc.name, rad), func(b *testing.B) {
				runBatchBenchmark(b, b.N, func() {
					_ = tc.idx.FindNearby(55.75, 37.62, rad)
				})
			})
		}
	}
}

// BenchmarkFindNearby_VaryPrecision — влияние точности геохэша на время запроса.
func BenchmarkFindNearby_VaryPrecision(b *testing.B) {
	const n = 50_000
	r := gen.NewRand(7)
	pts := gen.GeneratePoints(r, n)

	for _, prec := range []int{3, 4, 5, 6} {
		prec := prec
		idx := geoindex.New(prec)
		for _, p := range pts {
			idx.Insert(p)
		}
		b.Run(fmt.Sprintf("prec=%d", prec), func(b *testing.B) {
			runBatchBenchmark(b, b.N, func() {
				_ = idx.FindNearby(55.75, 37.62, 10.0)
			})
		})
	}
}

// BenchmarkBuildKDTree — стоимость построения k-d дерева (online vs balanced).
func BenchmarkBuildKDTree(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		size := size
		r := gen.NewRand(int64(size))
		pts := gen.GeneratePoints(r, size)

		b.Run(fmt.Sprintf("online/size=%d", size), func(b *testing.B) {
			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				rand.Shuffle(len(pts), func(i, j int) { pts[i], pts[j] = pts[j], pts[i] })
				b.StartTimer()

				kt := kdtree.New()
				for _, p := range pts {
					kt.Insert(p)
				}
			})
		})

		b.Run(fmt.Sprintf("balanced/size=%d", size), func(b *testing.B) {
			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				rand.Shuffle(len(pts), func(i, j int) { pts[i], pts[j] = pts[j], pts[i] })
				b.StartTimer()

				kt := kdtree.New()
				kt.BuildBalanced(pts)
			})
		})
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func runBatchBenchmark(b *testing.B, batchSize int, fn func()) {
	b.Helper()
	b.ReportAllocs()
	iterations := 0
	nsItemSamples := make([]float64, 0, 64)

	for b.Loop() {
		before := b.Elapsed()
		fn()
		after := b.Elapsed()
		delta := after - before
		if delta > 0 && batchSize > 0 {
			nsItemSamples = append(nsItemSamples, float64(delta.Nanoseconds())/float64(batchSize))
		}
		iterations++
	}

	elapsed := b.Elapsed()
	processed := float64(batchSize * iterations)
	if processed == 0 || elapsed <= 0 {
		return
	}
	b.ReportMetric(float64(elapsed.Nanoseconds())/processed, "ns/item")
	b.ReportMetric(processed/elapsed.Seconds(), "ops/s")
	b.ReportMetric(ci95(nsItemSamples), "ci95_ns/item")
}

func benchmarkSizes(b testing.TB) []int {
	b.Helper()
	raw := strings.TrimSpace(os.Getenv("SIZES"))
	if raw == "" {
		return defaultBenchmarkSizes
	}
	parts := strings.Split(raw, ",")
	sizes := make([]int, 0, len(parts))
	for _, part := range parts {
		v := strings.TrimSpace(part)
		if v == "" {
			continue
		}
		size, err := strconv.Atoi(v)
		if err != nil {
			b.Fatalf("invalid SIZES value %q: %v", v, err)
		}
		if size <= 0 {
			b.Fatalf("invalid SIZES value %q: must be positive", v)
		}
		sizes = append(sizes, size)
	}
	if len(sizes) == 0 {
		b.Fatalf("SIZES=%q contained no valid values", raw)
	}
	return sizes
}

func ci95(values []float64) float64 {
	n := len(values)
	if n <= 1 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)
	var sq float64
	for _, v := range values {
		d := v - mean
		sq += d * d
	}
	stdDev := math.Sqrt(sq / float64(n-1))
	return 1.96 * stdDev / math.Sqrt(float64(n))
}
