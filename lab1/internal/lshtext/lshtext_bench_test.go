package lshtext

import "testing"

func BenchmarkLSHIndexBuild(b *testing.B) {
	idx := NewIndex(3, 64, 8, 8)

	const n = 50000
	texts := make([][]byte, n)
	for i := 0; i < n; i++ {
		texts[i] = []byte("the quick brown fox jumps over the lazy dog " + string(rune('a'+(i%26))))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := i % n
		_ = idx.Add(id, texts[id])
	}
}

func BenchmarkLSHQuery(b *testing.B) {
	idx := NewIndex(3, 64, 8, 8)

	const n = 50000
	texts := make([][]byte, n)
	for i := 0; i < n; i++ {
		texts[i] = []byte("the quick brown fox jumps over the lazy dog " + string(rune('a'+(i%26))))
		_ = idx.Add(i, texts[i])
	}

	q := []byte("the quick brown fox jumps over the lazy dog z")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := idx.Query(q); err != nil {
			b.Fatalf("Query error: %v", err)
		}
	}
}

