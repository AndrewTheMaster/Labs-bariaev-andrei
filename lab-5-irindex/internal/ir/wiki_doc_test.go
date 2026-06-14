package ir

import (
	"strings"
	"testing"
)

func TestDocTermOccurrences(t *testing.T) {
	tokens := []string{"история", "древняя", "россии", "и", "история", "мамонтов"}
	occs := DocTermOccurrences(tokens, []string{"история", "россии"}, 2, 2)
	if len(occs) != 3 {
		t.Fatalf("want 3 occs, got %d", len(occs))
	}
	if occs[0].Term != "история" || occs[0].Pos != 0 {
		t.Fatalf("first occ: %+v", occs[0])
	}
	if !strings.Contains(occs[0].Line, ">>0:история<<") || !strings.Contains(occs[0].Line, "1:древняя") {
		t.Fatalf("line: %q", occs[0].Line)
	}
}

func TestFindNearOccurrences(t *testing.T) {
	tokens := []string{"a", "россия", "x", "китай", "россия", "китай"}
	nears := FindNearOccurrences(tokens, 2, "россия", "китай", 5, 2)
	if len(nears) < 2 {
		t.Fatalf("want near pairs, got %d", len(nears))
	}
}

func TestFindAdjOccurrences(t *testing.T) {
	tokens := []string{"история", "россии", "история", "и", "история", "россии"}
	adjs := FindAdjOccurrences(tokens, "история", "россии", 5, 2)
	if len(adjs) != 2 {
		t.Fatalf("want 2 adj, got %d", len(adjs))
	}
}
