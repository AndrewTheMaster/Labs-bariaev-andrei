package lsh3d

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

var defaultBenchmarkSizes = []int{1000, 5000, 10000, 50000, 100000}

// BenchmarkLSH3DBuild измеряет время построения индекса из N точек.
func BenchmarkLSH3DBuild(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			pts := genPoints(size, 42)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				idx, err := NewIndex(DefaultConfig())
				if err != nil {
					b.Fatalf("NewIndex: %v", err)
				}
				rand.Shuffle(len(pts), func(i, j int) { pts[i], pts[j] = pts[j], pts[i] })
				b.StartTimer()

				for _, p := range pts {
					idx.Add(p)
				}
			})
		})
	}
}

// BenchmarkLSH3DAdd измеряет скорость добавления N точек в уже заполненный индекс.
func BenchmarkLSH3DAdd(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			base := genPoints(size, 42)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				idx, err := NewIndex(DefaultConfig())
				if err != nil {
					b.Fatalf("NewIndex: %v", err)
				}
				for _, p := range base {
					idx.Add(p)
				}
				newPts := genPoints(size, 999)
				for i := range newPts {
					newPts[i].ID += size
				}
				b.StartTimer()

				for _, p := range newPts {
					idx.Add(p)
				}
			})
		})
	}
}

// BenchmarkLSH3DQuery измеряет скорость поиска ближайших дублей для N точек.
func BenchmarkLSH3DQuery(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			base := genPoints(size, 42)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				idx, err := NewIndex(DefaultConfig())
				if err != nil {
					b.Fatalf("NewIndex: %v", err)
				}
				for _, p := range base {
					idx.Add(p)
				}
				queries := genPoints(size, 7777)
				b.StartTimer()

				total := 0
				for _, q := range queries {
					total += len(idx.Query(q))
				}
				_ = total
			})
		})
	}
}

// BenchmarkLSH3DFullScan измеряет скорость полного сканирования индекса из N точек.
func BenchmarkLSH3DFullScan(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			base := genPoints(size, 42)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				idx, err := NewIndex(DefaultConfig())
				if err != nil {
					b.Fatalf("NewIndex: %v", err)
				}
				for _, p := range base {
					idx.Add(p)
				}
				b.StartTimer()

				pairs := idx.FullScanDuplicates(DefaultConfig().BandWidth)
				_ = pairs
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
