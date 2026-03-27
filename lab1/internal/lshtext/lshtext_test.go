package lshtext

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// TestIndexRejectsInvalidConfig проверяет, что некорректная конфигурация
// (Bands * RowsPerBand != SigSize) вызывает ошибку при создании индекса.
func TestIndexRejectsInvalidConfig(t *testing.T) {
	_, err := NewIndex(Config{SigSize: 64, Bands: 3, RowsPerBand: 10, ShingleSize: 2})
	if err == nil {
		t.Fatal("expected config validation error for Bands*RowsPerBand != SigSize")
	}
}

// TestIndexAddAndFindDuplicates добавляет несколько документов и проверяет,
// что точный дубликат находится с высокой схожестью.
func TestIndexAddAndFindDuplicates(t *testing.T) {
	idx, err := NewIndex(DefaultConfig())
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}

	docs := []struct {
		id   int
		text string
	}{
		{0, "Distributed hashing for duplicate document search with locality sensitive hashing"},
		{1, "Consensus protocols and replication improve distributed storage reliability"},
	}
	for _, d := range docs {
		if err := idx.Add(d.id, []byte(d.text)); err != nil {
			t.Fatalf("Add id=%d: %v", d.id, err)
		}
	}

	// Запрашиваем дубликаты документа 0 — используем точно тот же текст.
	query := docs[0].text
	cands, err := idx.Query([]byte(query))
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(cands) == 0 {
		t.Fatal("expected at least one duplicate match")
	}
	// Первый кандидат должен быть документ 0 с высокой схожестью
	found := false
	for _, c := range cands {
		if c.ID == 0 && c.Similarity >= 0.8 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected document 0 with similarity >= 0.8, got candidates: %+v", cands)
	}
}

// TestIndexFullScanMatchesAddedDuplicates строит индекс из 2026 документов,
// добавляет 8 новых точных копий уже проиндексированных и проверяет,
// что FullScanDuplicates находит каждый из них.
func TestIndexFullScanMatchesAddedDuplicates(t *testing.T) {
	idx, err := NewIndex(DefaultConfig())
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}

	baseDocs := randomDocuments(2026)
	for i, d := range baseDocs {
		if err := idx.Add(i, d); err != nil {
			t.Fatalf("Add base doc id=%d: %v", i, err)
		}
	}

	// Добавляем 8 точных копий из baseDocs под новыми ID.
	newBase := 10000
	for j := 0; j < 8; j++ {
		if err := idx.Add(newBase+j, baseDocs[j]); err != nil {
			t.Fatalf("Add duplicate id=%d: %v", newBase+j, err)
		}
	}

	// Для каждого добавленного дубля FullScanDuplicates должен вернуть хотя бы один результат.
	for j := 0; j < 8; j++ {
		pairs, err := idx.FullScanDuplicates(0.8)
		if err != nil {
			t.Fatalf("FullScanDuplicates: %v", err)
		}
		// Ищем пару (j, newBase+j) или (newBase+j, j).
		found := false
		for _, p := range pairs {
			if (p.ID1 == j && p.ID2 == newBase+j) || (p.ID1 == newBase+j && p.ID2 == j) {
				found = true
				break
			}
		}
		if !found {
			// LSH может иметь false negatives, но не для точных копий с SigSize=64
			t.Errorf("duplicate for doc %d not found in FullScanDuplicates", j)
		}
	}
}

// FuzzIndexExactDuplicatesAreFound — фаззинг: добавляем документ и сразу
// ищем его через Query. Точный дубликат обязан найтись.
func FuzzIndexExactDuplicatesAreFound(f *testing.F) {
	for _, sample := range []string{
		"distributed hashing duplicate search",
		"LSH supports approximate nearest neighbors",
		strings.Repeat("token ", 8),
	} {
		f.Add(sample)
	}

	f.Fuzz(func(t *testing.T, text string) {
		idx, err := NewIndex(DefaultConfig())
		if err != nil {
			t.Fatalf("NewIndex: %v", err)
		}
		if err := idx.Add(1, []byte(text)); err != nil {
			t.Fatalf("Add: %v", err)
		}
		cands, err := idx.Query([]byte(text))
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if len(cands) == 0 {
			// Только если текст слишком короткий для шинглов
			if len(strings.Fields(text)) < DefaultConfig().ShingleSize {
				return
			}
			t.Fatalf("expected exact duplicate for %q, got no candidates", text)
		}
		found := false
		for _, c := range cands {
			if c.ID == 1 {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected doc 1 in candidates for exact duplicate of %q, got %+v", text, cands)
		}
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

var benchVocabulary = []string{
	"distributed", "database", "indexing", "hashing", "algorithm", "structure",
	"performance", "latency", "throughput", "cache", "memory", "disk", "network",
	"query", "insert", "delete", "update", "search", "binary", "hash",
	"tree", "graph", "node", "edge", "bucket", "collision", "cluster", "shard",
	"replica", "primary", "secondary", "transaction", "commit", "snapshot",
	"prefix", "suffix", "token", "signature", "band", "row", "column",
}

func randomWord(r *rand.Rand) string {
	return benchVocabulary[r.Intn(len(benchVocabulary))]
}

func randomDocument(r *rand.Rand, wordCount int) []byte {
	words := make([]string, wordCount)
	for i := range words {
		words[i] = randomWord(r)
	}
	return []byte(strings.Join(words, " "))
}

func randomDocuments(size int) [][]byte {
	r := rand.New(rand.NewSource(42))
	docs := make([][]byte, size)
	for i := range docs {
		docs[i] = randomDocument(r, 20+r.Intn(20))
	}
	return docs
}

var _ = fmt.Sprintf // ensure fmt is used
