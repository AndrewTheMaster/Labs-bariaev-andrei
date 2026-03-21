package hashfs

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkStore_Insert(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.dat")

	store, err := Open(path, Options{BucketCount: 1 << 20})
	if err != nil {
		b.Fatalf("Open error: %v", err)
	}
	defer store.Close()

	keys := make([][]byte, b.N)
	vals := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
		vals[i] = []byte("value-data-0123456789")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Put(keys[i], vals[i]); err != nil {
			b.Fatalf("Put error: %v", err)
		}
	}
}

func BenchmarkStore_RandomGet(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench_get.dat")

	store, err := Open(path, Options{BucketCount: 1 << 20})
	if err != nil {
		b.Fatalf("Open error: %v", err)
	}
	defer store.Close()

	const N = 100000
	keys := make([][]byte, N)
	for i := 0; i < N; i++ {
		k := []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
		keys[i] = k
		if err := store.Put(k, []byte("value")); err != nil {
			b.Fatalf("Put error: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := keys[i%N]
		if _, err := store.Get(k); err != nil {
			b.Fatalf("Get error: %v", err)
		}
	}
}

func BenchmarkStore_UpdateHeavy(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench_update.dat")

	store, err := Open(path, Options{BucketCount: 1 << 20})
	if err != nil {
		b.Fatalf("Open error: %v", err)
	}
	defer store.Close()

	const N = 100000
	keys := make([][]byte, N)
	for i := 0; i < N; i++ {
		k := []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
		keys[i] = k
		if err := store.Put(k, []byte("value")); err != nil {
			b.Fatalf("Put error: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := keys[i%N]
		if err := store.Put(k, []byte("value-updated")); err != nil {
			b.Fatalf("Put(update) error: %v", err)
		}
	}
}

func BenchmarkStore_FileSize(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench_size.dat")

	store, err := Open(path, Options{BucketCount: 1 << 20})
	if err != nil {
		b.Fatalf("Open error: %v", err)
	}
	defer store.Close()

	for i := 0; i < 100000; i++ {
		k := []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
		if err := store.Put(k, []byte("value")); err != nil {
			b.Fatalf("Put error: %v", err)
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		b.Fatalf("Stat error: %v", err)
	}
	b.Logf("file size after 100k puts: %d bytes", info.Size())
}

