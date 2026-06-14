package ir

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"os"
	"sort"
	"strings"
)

// WikiDoc — текст статьи по docID (порядок как при BuildIndexFromWikiXML / AddLeanTitle).
type WikiDoc struct {
	DocID  uint32
	Title  string
	Text   string
	Tokens []string
}

// FetchWikiDoc возвращает документ с заданным docID из XML (0-based, как в индексе).
func FetchWikiDoc(xmlPath string, docID uint32) (WikiDoc, error) {
	f, err := os.Open(xmlPath)
	if err != nil {
		return WikiDoc{}, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	dec.Strict = false
	var idx uint32
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return WikiDoc{}, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "page" {
			continue
		}
		var page wikiPageXML
		if err := dec.DecodeElement(&page, &se); err != nil {
			return WikiDoc{}, err
		}
		text := strings.TrimSpace(page.Text)
		if text == "" {
			continue
		}
		if idx == docID {
			title := strings.TrimSpace(page.Title)
			toks := Tokenize(text)
			return WikiDoc{DocID: docID, Title: title, Text: text, Tokens: toks}, nil
		}
		idx++
	}
	return WikiDoc{}, fmt.Errorf("docID %d not found (indexed %d docs in stream)", docID, idx)
}

// FormatTokenLine — токены с позициями для отладки ADJ/NEAR.
func FormatTokenLine(tokens []string, max int) string {
	if max <= 0 || max > len(tokens) {
		max = len(tokens)
	}
	var b strings.Builder
	for i := 0; i < max; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%d:%s", i, tokens[i])
	}
	if max < len(tokens) {
		fmt.Fprintf(&b, " … (+%d)", len(tokens)-max)
	}
	return b.String()
}

// FindAdjPairs возвращает позиции, где a сразу перед b (для демо ADJ).
func FindAdjPairs(tokens []string, a, b string) []int {
	var out []int
	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i] == a && tokens[i+1] == b {
			out = append(out, i)
		}
	}
	return out
}

// TextExcerpt — первые maxRunes символов текста.
func TextExcerpt(text string, maxRunes int) string {
	r := []rune(text)
	if len(r) <= maxRunes {
		return text
	}
	return string(r[:maxRunes]) + "…"
}

// TermOccurrence — одно вхождение терма с окном токенов вокруг.
type TermOccurrence struct {
	Term     string
	Pos      int
	Line     string // plain для CLI
	LineHTML string // для irbrowse
}

// FindTokenPositions возвращает все позиции term в tokens.
func FindTokenPositions(tokens []string, term string) []int {
	var out []int
	for i, tok := range tokens {
		if tok == term {
			out = append(out, i)
		}
	}
	return out
}

// FormatTokenWindow форматирует окно токенов вокруг pos; highlight — позиции с >>…<<.
func FormatTokenWindow(tokens []string, pos, before, after int, highlight map[int]bool) string {
	start := pos - before
	if start < 0 {
		start = 0
	}
	end := pos + after + 1
	if end > len(tokens) {
		end = len(tokens)
	}
	var b strings.Builder
	if start > 0 {
		b.WriteString("… ")
	}
	for i := start; i < end; i++ {
		if i > start {
			b.WriteByte(' ')
		}
		if highlight != nil && highlight[i] {
			fmt.Fprintf(&b, ">>%d:%s<<", i, tokens[i])
		} else {
			fmt.Fprintf(&b, "%d:%s", i, tokens[i])
		}
	}
	if end < len(tokens) {
		fmt.Fprintf(&b, " … (+%d)", len(tokens)-end)
	}
	return b.String()
}

// FormatTokenWindowHTML — то же окно, подсветка: class tok-a | tok-b | tok-hit.
func FormatTokenWindowHTML(tokens []string, pos, before, after int, markClass map[int]string) string {
	start := pos - before
	if start < 0 {
		start = 0
	}
	end := pos + after + 1
	if end > len(tokens) {
		end = len(tokens)
	}
	var b strings.Builder
	if start > 0 {
		b.WriteString("… ")
	}
	for i := start; i < end; i++ {
		if i > start {
			b.WriteByte(' ')
		}
		label := fmt.Sprintf("%d:%s", i, tokens[i])
		if cls := markClass[i]; cls != "" {
			fmt.Fprintf(&b, `<mark class="tok-%s">%s</mark>`, cls, html.EscapeString(label))
		} else {
			b.WriteString(html.EscapeString(label))
		}
	}
	if end < len(tokens) {
		fmt.Fprintf(&b, " … (+%d)", len(tokens)-end)
	}
	return b.String()
}

func markClassFromBool(hl map[int]bool) map[int]string {
	if hl == nil {
		return nil
	}
	out := make(map[int]string, len(hl))
	for i := range hl {
		out[i] = "hit"
	}
	return out
}

func nearWindowCenter(posA, posB, radius int) (center, before, after int) {
	lo, hi := posA, posB
	if lo > hi {
		lo, hi = hi, lo
	}
	center = lo
	before = radius
	after = radius + (hi - lo)
	return center, before, after
}

// DocTermOccurrences — первые maxPerTerm вхождений каждого терма с контекстом ±radius токенов.
func DocTermOccurrences(tokens []string, terms []string, maxPerTerm, radius int) []TermOccurrence {
	if maxPerTerm <= 0 {
		maxPerTerm = 3
	}
	if radius <= 0 {
		radius = 8
	}
	var out []TermOccurrence
	for _, term := range terms {
		positions := FindTokenPositions(tokens, term)
		limit := len(positions)
		if limit > maxPerTerm {
			limit = maxPerTerm
		}
		for i := 0; i < limit; i++ {
			pos := positions[i]
			hl := map[int]bool{pos: true}
			mc := markClassFromBool(hl)
			out = append(out, TermOccurrence{
				Term: term,
				Pos:  pos,
				Line: FormatTokenWindow(tokens, pos, radius, radius, hl),
				LineHTML: FormatTokenWindowHTML(tokens, pos, radius, radius, mc),
			})
		}
	}
	return out
}

// DocTermOccurrencesPrioritized — сначала вхождения из priority (NEAR/ADJ), потом остальные.
func DocTermOccurrencesPrioritized(tokens []string, terms []string, maxPerTerm, radius int, priority map[int]bool) []TermOccurrence {
	if maxPerTerm <= 0 {
		maxPerTerm = 3
	}
	if radius <= 0 {
		radius = 8
	}
	var out []TermOccurrence
	for _, term := range terms {
		positions := FindTokenPositions(tokens, term)
		sort.SliceStable(positions, func(i, j int) bool {
			pi, pj := priority[positions[i]], priority[positions[j]]
			if pi != pj {
				return pi && !pj
			}
			return positions[i] < positions[j]
		})
		limit := len(positions)
		if limit > maxPerTerm {
			limit = maxPerTerm
		}
		for i := 0; i < limit; i++ {
			pos := positions[i]
			hl := map[int]bool{pos: true}
			mc := markClassFromBool(hl)
			prio := priority[pos]
			if prio {
				mc[pos] = "prio"
			}
			out = append(out, TermOccurrence{
				Term: term,
				Pos:  pos,
				Line: FormatTokenWindow(tokens, pos, radius, radius, hl),
				LineHTML: FormatTokenWindowHTML(tokens, pos, radius, radius, mc),
			})
		}
	}
	return out
}

// FindAdjOccurrences — все ADJ-пары a сразу перед b с контекстом.
func FindAdjOccurrences(tokens []string, a, b string, maxHits, radius int) []AdjOccurrence {
	if maxHits <= 0 {
		maxHits = 5
	}
	if radius <= 0 {
		radius = 8
	}
	var out []AdjOccurrence
	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i] != a || tokens[i+1] != b {
			continue
		}
		hl := map[int]bool{i: true, i + 1: true}
		mc := map[int]string{i: "a", i + 1: "b"}
		out = append(out, AdjOccurrence{
			A: a, B: b, Pos: i,
			Line:     FormatTokenWindow(tokens, i, radius, radius, hl),
			LineHTML: FormatTokenWindowHTML(tokens, i, radius, radius, mc),
		})
		if len(out) >= maxHits {
			break
		}
	}
	return out
}

// AdjOccurrence — соседняя пара a→b с контекстом.
type AdjOccurrence struct {
	A, B     string
	Pos      int
	Line     string
	LineHTML string
}

// NearOccurrence — пара a,b в окне NEAR(k).
type NearOccurrence struct {
	K          int
	A, B       string
	PosA, PosB int
	Dist       int
	Line       string
	LineHTML   string
}

// FindNearOccurrences — пары a,b с |posA-posB| ≤ k (как в eval NEAR).
func FindNearOccurrences(tokens []string, k int, a, b string, maxHits, radius int) []NearOccurrence {
	if maxHits <= 0 {
		maxHits = 5
	}
	if radius <= 0 {
		radius = 8
	}
	pa := FindTokenPositions(tokens, a)
	pb := FindTokenPositions(tokens, b)
	var out []NearOccurrence
	for _, pA := range pa {
		for _, pB := range pb {
			diff := pA - pB
			if diff < 0 {
				diff = -diff
			}
			if diff > k {
				continue
			}
			center, before, after := nearWindowCenter(pA, pB, radius)
			hl := map[int]bool{pA: true, pB: true}
			mc := map[int]string{pA: "a", pB: "b"}
			out = append(out, NearOccurrence{
				K: k, A: a, B: b, PosA: pA, PosB: pB, Dist: diff,
				Line:     FormatTokenWindow(tokens, center, before, after, hl),
				LineHTML: FormatTokenWindowHTML(tokens, center, before, after, mc),
			})
			if len(out) >= maxHits {
				return out
			}
		}
	}
	return out
}

// CollectAdjFromQuery находит все ADJ(...) в AST и возвращает их вхождения в документе.
func CollectAdjFromQuery(n Node, tokens []string, maxPerAdj, radius int) []AdjOccurrence {
	var out []AdjOccurrence
	var walk func(Node)
	walk = func(x Node) {
		switch t := x.(type) {
		case *Adj:
			out = append(out, FindAdjOccurrences(tokens, t.A, t.B, maxPerAdj, radius)...)
		case *Not:
			walk(t.Child)
		case *And:
			for _, ch := range t.Children {
				walk(ch)
			}
		case *Or:
			walk(t.Left)
			walk(t.Right)
		}
	}
	walk(n)
	return out
}

// CollectNearFromQuery — все NEAR(k,…) из AST и их вхождения в документе.
func CollectNearFromQuery(n Node, tokens []string, maxPerNear, radius int) []NearOccurrence {
	var out []NearOccurrence
	var walk func(Node)
	walk = func(x Node) {
		switch t := x.(type) {
		case *Near:
			out = append(out, FindNearOccurrences(tokens, t.K, t.A, t.B, maxPerNear, radius)...)
		case *Not:
			walk(t.Child)
		case *And:
			for _, ch := range t.Children {
				walk(ch)
			}
		case *Or:
			walk(t.Left)
			walk(t.Right)
		}
	}
	walk(n)
	return out
}

func queryHasAdj(n Node) bool {
	found := false
	var walk func(Node)
	walk = func(x Node) {
		if found {
			return
		}
		switch t := x.(type) {
		case *Adj:
			found = true
		case *Not:
			walk(t.Child)
		case *And:
			for _, ch := range t.Children {
				walk(ch)
			}
		case *Or:
			walk(t.Left)
			walk(t.Right)
		}
	}
	walk(n)
	return found
}

func queryHasNear(n Node) bool {
	found := false
	var walk func(Node)
	walk = func(x Node) {
		if found {
			return
		}
		switch t := x.(type) {
		case *Near:
			found = true
		case *Not:
			walk(t.Child)
		case *And:
			for _, ch := range t.Children {
				walk(ch)
			}
		case *Or:
			walk(t.Left)
			walk(t.Right)
		}
	}
	walk(n)
	return found
}

// DocDebugResult — разбор документа для отладки запроса.
type DocDebugResult struct {
	Terms         []string
	TermOccs      []TermOccurrence
	Adjs          []AdjOccurrence
	Nears         []NearOccurrence
	HasAdjInQuery bool
	HasNearInQuery bool
}

// DocQueryDebug — контексты вхождений термов, ADJ- и NEAR-пар.
func DocQueryDebug(query string, tokens []string) (DocDebugResult, error) {
	n, err := Parse(query)
	if err != nil {
		return DocDebugResult{}, err
	}
	terms := PositiveTerms(n)
	adjs := CollectAdjFromQuery(n, tokens, 3, 8)
	nears := CollectNearFromQuery(n, tokens, 5, 8)
	priority := make(map[int]bool)
	for _, a := range adjs {
		priority[a.Pos] = true
		priority[a.Pos+1] = true
	}
	for _, nr := range nears {
		priority[nr.PosA] = true
		priority[nr.PosB] = true
	}
	return DocDebugResult{
		Terms:          terms,
		TermOccs:       DocTermOccurrencesPrioritized(tokens, terms, 5, 8, priority),
		Adjs:           adjs,
		Nears:          nears,
		HasAdjInQuery:  queryHasAdj(n),
		HasNearInQuery: queryHasNear(n),
	}, nil
}
