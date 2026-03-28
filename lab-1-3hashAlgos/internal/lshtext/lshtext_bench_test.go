package lshtext

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

var defaultBenchmarkSizes = []int{500, 1000, 2500, 5000, 7500}

// BenchmarkIndexBuild измеряет время построения индекса с нуля из N документов.
func BenchmarkIndexBuild(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			docs := randomDocuments(size)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				idx, err := NewIndex(DefaultConfig())
				if err != nil {
					b.Fatalf("NewIndex: %v", err)
				}
				rand.Shuffle(len(docs), func(i, j int) { docs[i], docs[j] = docs[j], docs[i] })
				b.StartTimer()

				for i, d := range docs {
					if err := idx.Add(i, d); err != nil {
						b.Fatalf("Add: %v", err)
					}
				}
			})
		})
	}
}

// BenchmarkIndexAdd измеряет скорость добавления N новых документов
// в уже заполненный индекс из N документов.
func BenchmarkIndexAdd(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			baseDocs := randomDocuments(size)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				idx, err := NewIndex(DefaultConfig())
				if err != nil {
					b.Fatalf("NewIndex: %v", err)
				}
				for i, d := range baseDocs {
					if err := idx.Add(i, d); err != nil {
						b.Fatalf("Build Add: %v", err)
					}
				}
				docsForInsert := randomDocuments(size)
				b.StartTimer()

				for i, d := range docsForInsert {
					if err := idx.Add(size+i, d); err != nil {
						b.Fatalf("Add: %v", err)
					}
				}
			})
		})
	}
}

// BenchmarkIndexFullScan измеряет скорость полного сканирования индекса
// для поиска дублей N запросных документов.
func BenchmarkIndexFullScan(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			baseDocs := randomDocuments(size)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				idx, err := NewIndex(DefaultConfig())
				if err != nil {
					b.Fatalf("NewIndex: %v", err)
				}
				for i, d := range baseDocs {
					if err := idx.Add(i, d); err != nil {
						b.Fatalf("Build Add: %v", err)
					}
				}
				b.StartTimer()

				totalMatches := 0
				for _, doc := range baseDocs {
					cands, err := idx.Query(doc)
					if err != nil {
						b.Fatalf("Query: %v", err)
					}
					count := 0
					for _, c := range cands {
						if c.Similarity >= 0.5 {
							count++
						}
					}
					totalMatches += count
				}
				_ = totalMatches
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
