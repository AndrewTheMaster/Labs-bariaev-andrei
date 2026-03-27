package hashfs

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var defaultBenchmarkSizes = []int{1000, 5000, 10000, 50000, 100000, 500000, 1000000}

// BenchmarkStoreInsert измеряет пропускную способность вставки N уникальных ключей
// в пустое хранилище. Перед каждой итерацией хранилище сбрасывается через Reset().
func BenchmarkStoreInsert(b *testing.B) {
	baseDir := b.TempDir()
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			keys := makeDataset(size)
			path := filepath.Join(baseDir, fmt.Sprintf("insert-%d.dat", size))
			store := mustBenchmarkStore(b, path)
			b.Cleanup(func() {
				if err := store.Close(); err != nil {
					b.Fatalf("Close: %v", err)
				}
				os.Remove(path)
			})

			runBatchBenchmark(b, store, size, func() {
				b.StopTimer()
				shuffled := shuffleKeys(keys)
				b.StartTimer()

				for _, key := range shuffled {
					if err := store.Put(key, benchValue); err != nil {
						b.Fatalf("Put key=%q: %v", key, err)
					}
				}
			})
		})
	}
}

// BenchmarkStoreUpdate измеряет пропускную способность обновления N существующих ключей.
func BenchmarkStoreUpdate(b *testing.B) {
	baseDir := b.TempDir()
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			keys := makeDataset(size)
			path := filepath.Join(baseDir, fmt.Sprintf("update-%d.dat", size))
			store := mustBenchmarkStore(b, path)
			b.Cleanup(func() {
				if err := store.Close(); err != nil {
					b.Fatalf("Close: %v", err)
				}
				os.Remove(path)
			})

			runBatchBenchmark(b, store, size, func() {
				b.StopTimer()
				shuffled := shuffleKeys(keys)
				for _, key := range shuffled {
					if err := store.Put(key, benchValue); err != nil {
						b.Fatalf("prepare update Put key=%q: %v", key, err)
					}
				}
				b.StartTimer()

				for _, key := range shuffled {
					if err := store.Put(key, benchValueAlt); err != nil {
						b.Fatalf("Update key=%q: %v", key, err)
					}
				}
			})
		})
	}
}

// BenchmarkStoreDelete измеряет пропускную способность удаления N ключей.
func BenchmarkStoreDelete(b *testing.B) {
	baseDir := b.TempDir()
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			keys := makeDataset(size)
			path := filepath.Join(baseDir, fmt.Sprintf("delete-%d.dat", size))
			store := mustBenchmarkStore(b, path)
			b.Cleanup(func() {
				if err := store.Close(); err != nil {
					b.Fatalf("Close: %v", err)
				}
				os.Remove(path)
			})

			runBatchBenchmark(b, store, size, func() {
				b.StopTimer()
				shuffled := shuffleKeys(keys)
				for _, key := range shuffled {
					if err := store.Put(key, benchValue); err != nil {
						b.Fatalf("prepare delete Put key=%q: %v", key, err)
					}
				}
				b.StartTimer()

				for _, key := range shuffled {
					if err := store.Delete(key); err != nil {
						b.Fatalf("Delete key=%q: %v", key, err)
					}
				}
			})
		})
	}
}

// BenchmarkStoreGet измеряет пропускную способность случайного чтения N ключей.
func BenchmarkStoreGet(b *testing.B) {
	baseDir := b.TempDir()
	for _, size := range benchmarkSizes(b) {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			keys := makeDataset(size)
			path := filepath.Join(baseDir, fmt.Sprintf("get-%d.dat", size))
			store := mustBenchmarkStore(b, path)
			b.Cleanup(func() {
				if err := store.Close(); err != nil {
					b.Fatalf("Close: %v", err)
				}
				os.Remove(path)
			})

			runBatchBenchmark(b, store, size, func() {
				b.StopTimer()
				shuffled := shuffleKeys(keys)
				for _, key := range shuffled {
					if err := store.Put(key, benchValue); err != nil {
						b.Fatalf("prepare get Put key=%q: %v", key, err)
					}
				}
				b.StartTimer()

				var sink []byte
				for _, key := range shuffled {
					val, err := store.Get(key)
					if err != nil {
						b.Fatalf("Get key=%q: %v", key, err)
					}
					sink = val
				}
				_ = sink
			})
		})
	}
}

// ── Internal helpers ──────────────────────────────────────────────────────────

var (
	benchValue    = []byte("value:0123456789abcdef")
	benchValueAlt = []byte("value:fedcba9876543210")
)

func mustBenchmarkStore(b testing.TB, path string) Store {
	b.Helper()
	s, err := Open(path, Options{BucketCount: 1 << 20})
	if err != nil {
		b.Fatalf("Open: %v", err)
	}
	return s
}

func runBatchBenchmark(b *testing.B, store Store, batchSize int, fn func()) {
	b.Helper()
	b.ReportAllocs()
	iterations := 0
	nsItemSamples := make([]float64, 0, 64)

	for b.Loop() {
		b.StopTimer()
		if err := store.Reset(); err != nil {
			b.Fatalf("Reset: %v", err)
		}
		b.StartTimer()

		before := b.Elapsed()
		fn()
		after := b.Elapsed()

		b.StopTimer()
		delta := after - before
		if delta > 0 && batchSize > 0 {
			nsItemSamples = append(nsItemSamples, float64(delta.Nanoseconds())/float64(batchSize))
		}
		iterations++
		b.StartTimer()
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

func shuffleKeys(keys [][]byte) [][]byte {
	out := make([][]byte, len(keys))
	copy(out, keys)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
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
