package perfecthash

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

var defaultBenchmarkSizes = []int{20000, 50000, 100000, 250000, 500000, 1000000}

// BenchmarkPerfectHashBuild измеряет время построения таблицы из N ключей.
func BenchmarkPerfectHashBuild(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			rng := rand.New(rand.NewSource(int64(size)))
			keys := makeDataset(rng, size)

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
				b.StartTimer()

				builder := &Builder{}
				if _, err := builder.Build(keys); err != nil {
					b.Fatalf("Build: %v", err)
				}
			})
		})
	}
}

// BenchmarkPerfectHashLookup измеряет скорость поиска по уже построенной таблице.
// Для каждого размера N: строим таблицу один раз, затем N lookup'ов в цикле.
func BenchmarkPerfectHashLookup(b *testing.B) {
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			rng := rand.New(rand.NewSource(int64(size)))
			keys := makeDataset(rng, size)

			builder := &Builder{}
			table, err := builder.Build(keys)
			if err != nil {
				b.Fatalf("Build: %v", err)
			}

			runBatchBenchmark(b, size, func() {
				b.StopTimer()
				rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
				b.StartTimer()

				var sum int
				for _, k := range keys {
					idx, ok := table.Lookup(k)
					if !ok {
						b.Fatalf("Lookup failed for known key")
					}
					sum += idx
				}
				_ = sum
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
