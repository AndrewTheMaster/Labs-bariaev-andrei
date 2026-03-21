package lshtext

import (
	"fmt"
	"testing"
)

func TestIndex_FindsNearDuplicates(t *testing.T) {
	idx := NewIndex(3, 64, 8, 8)

	docs := []string{
		"the quick brown fox jumps over the lazy dog",
		"the quick brown fox jumps over the very lazy dog",
		"a completely different sentence with other words",
		"the quick brown fox jumps over the lazy dog again",
	}

	for i, doc := range docs {
		if err := idx.Add(i, []byte(doc)); err != nil {
			t.Fatalf("Add error: %v", err)
		}
	}

	// Документ 0 и 3 должны быть довольно похожи.
	cands, err := idx.Query([]byte(docs[0]))
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}

	foundSelf := false
	foundSimilar := false
	for _, c := range cands {
		if c.ID == 0 {
			foundSelf = true
		}
		if c.ID == 3 && c.Similarity > 0.5 {
			foundSimilar = true
		}
	}
	if !foundSelf {
		t.Fatalf("expected to see the document itself among candidates")
	}
	if !foundSimilar {
		t.Fatalf("expected to find document 3 as similar to 0")
	}

	// full scan по дублям
	pairs, err := idx.FullScanDuplicates(0.5)
	if err != nil {
		t.Fatalf("FullScanDuplicates error: %v", err)
	}
	if len(pairs) == 0 {
		t.Fatalf("expected at least one duplicate pair")
	}
}

func TestIndex_AddAndQueryNoiseCorpus(t *testing.T) {
	idx := NewIndex(3, 64, 8, 8)

	// Синтетический корпус: 20 похожих вариаций и 80 шумовых текстов.
	base := "distributed databases need robust indexing and hashing strategies"
	for i := 0; i < 20; i++ {
		doc := fmt.Sprintf("%s variant-%d", base, i%4)
		if err := idx.Add(i, []byte(doc)); err != nil {
			t.Fatalf("Add similar doc error: %v", err)
		}
	}
	for i := 20; i < 100; i++ {
		doc := fmt.Sprintf("noise text %d with unrelated tokens %d", i, i*17)
		if err := idx.Add(i, []byte(doc)); err != nil {
			t.Fatalf("Add noise doc error: %v", err)
		}
	}

	cands, err := idx.Query([]byte(base + " variant-1"))
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if len(cands) == 0 {
		t.Fatalf("expected non-empty candidate list")
	}

	high := 0
	for _, c := range cands {
		if c.Similarity >= 0.6 {
			high++
		}
	}
	if high == 0 {
		t.Fatalf("expected at least one high-similarity candidate")
	}
}

func TestIndex_FullScanThresholdBehavior(t *testing.T) {
	idx := NewIndex(3, 64, 8, 8)
	docs := []string{
		"alpha beta gamma delta epsilon",
		"alpha beta gamma delta epsilon zeta",
		"totally unrelated sentence with another vocabulary",
	}
	for i, doc := range docs {
		if err := idx.Add(i, []byte(doc)); err != nil {
			t.Fatalf("Add error: %v", err)
		}
	}

	pairsLow, err := idx.FullScanDuplicates(0.2)
	if err != nil {
		t.Fatalf("FullScanDuplicates low threshold error: %v", err)
	}
	pairsHigh, err := idx.FullScanDuplicates(0.8)
	if err != nil {
		t.Fatalf("FullScanDuplicates high threshold error: %v", err)
	}

	if len(pairsLow) < len(pairsHigh) {
		t.Fatalf("expected low threshold to produce at least as many pairs")
	}
}

