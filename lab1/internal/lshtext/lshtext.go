package lshtext

import (
	"hash/fnv"
	"strings"
)

// Candidate представляет документ-кандидат с оценкой схожести.
type Candidate struct {
	ID        int
	Similarity float64
}

// Pair описывает пару документов-дублей.
type Pair struct {
	ID1       int
	ID2       int
	Similarity float64
}

// Index — реализация LSH для текстовых документов на основе MinHash.
type Index struct {
	shingleSize int
	sigSize     int
	bands       int
	rowsPerBand int

	// signatures хранит MinHash-подписи документов.
	signatures map[int][]uint64
	// bandBuckets[band][bucketHash] -> список docID.
	bandBuckets []map[uint64][]int
}

// NewIndex создаёт новый индекс с заданными параметрами.
// shingleSize — размер шингла (в словах), sigSize — длина подписи,
// bands и rowsPerBand задают LSH-разбиение (bands*rowsPerBand == sigSize).
func NewIndex(shingleSize, sigSize, bands, rowsPerBand int) *Index {
	if sigSize <= 0 {
		sigSize = 64
	}
	if shingleSize <= 0 {
		shingleSize = 3
	}
	if bands <= 0 {
		bands = 8
	}
	if rowsPerBand <= 0 {
		rowsPerBand = sigSize / bands
	}
	bandBuckets := make([]map[uint64][]int, bands)
	for i := range bandBuckets {
		bandBuckets[i] = make(map[uint64][]int)
	}
	return &Index{
		shingleSize: shingleSize,
		sigSize:     sigSize,
		bands:       bands,
		rowsPerBand: rowsPerBand,
		signatures:  make(map[int][]uint64),
		bandBuckets: bandBuckets,
	}
}

// Add добавляет документ в индекс.
func (idx *Index) Add(docID int, text []byte) error {
	sig := idx.computeSignature(text)
	idx.signatures[docID] = sig

	for b := 0; b < idx.bands; b++ {
		start := b * idx.rowsPerBand
		end := start + idx.rowsPerBand
		if end > len(sig) {
			end = len(sig)
		}
		h := hashBand(sig[start:end])
		bucket := idx.bandBuckets[b][h]
		bucket = append(bucket, docID)
		idx.bandBuckets[b][h] = bucket
	}
	return nil
}

// Query возвращает список кандидатов-документов, похожих на text.
func (idx *Index) Query(text []byte) ([]Candidate, error) {
	querySig := idx.computeSignature(text)
	candidates := make(map[int]struct{})

	for b := 0; b < idx.bands; b++ {
		start := b * idx.rowsPerBand
		end := start + idx.rowsPerBand
		if end > len(querySig) {
			end = len(querySig)
		}
		h := hashBand(querySig[start:end])
		for _, id := range idx.bandBuckets[b][h] {
			candidates[id] = struct{}{}
		}
	}

	var result []Candidate
	for id := range candidates {
		sig := idx.signatures[id]
		sim := jaccardSignature(querySig, sig)
		result = append(result, Candidate{ID: id, Similarity: sim})
	}
	return result, nil
}

// FullScanDuplicates ищет пары дублей среди всех документов, используя LSH-кандидатов.
func (idx *Index) FullScanDuplicates(threshold float64) ([]Pair, error) {
	seen := make(map[[2]int]struct{})
	var pairs []Pair

	for docID, sig := range idx.signatures {
		cands, _ := idx.QuerySignature(sig)
		for _, c := range cands {
			if c.ID == docID {
				continue
			}
			ids := [2]int{docID, c.ID}
			if ids[0] > ids[1] {
				ids[0], ids[1] = ids[1], ids[0]
			}
			if _, ok := seen[ids]; ok {
				continue
			}
			seen[ids] = struct{}{}
			if c.Similarity >= threshold {
				pairs = append(pairs, Pair{
					ID1:        ids[0],
					ID2:        ids[1],
					Similarity: c.Similarity,
				})
			}
		}
	}
	return pairs, nil
}

// QuerySignature ищет кандидатов по уже посчитанной подписи.
func (idx *Index) QuerySignature(sig []uint64) ([]Candidate, error) {
	candidates := make(map[int]struct{})

	for b := 0; b < idx.bands; b++ {
		start := b * idx.rowsPerBand
		end := start + idx.rowsPerBand
		if end > len(sig) {
			end = len(sig)
		}
		h := hashBand(sig[start:end])
		for _, id := range idx.bandBuckets[b][h] {
			candidates[id] = struct{}{}
		}
	}

	var result []Candidate
	for id := range candidates {
		s := idx.signatures[id]
		sim := jaccardSignature(sig, s)
		result = append(result, Candidate{ID: id, Similarity: sim})
	}
	return result, nil
}

func (idx *Index) computeSignature(text []byte) []uint64 {
	shingles := makeShingles(string(text), idx.shingleSize)
	if len(shingles) == 0 {
		return make([]uint64, idx.sigSize)
	}

	sig := make([]uint64, idx.sigSize)
	for i := 0; i < idx.sigSize; i++ {
		min := uint64(^uint64(0))
		for _, s := range shingles {
			h := minhash(s, uint64(i+1))
			if h < min {
				min = h
			}
		}
		sig[i] = min
	}
	return sig
}

func makeShingles(text string, k int) []string {
	words := strings.Fields(text)
	if len(words) < k || k <= 0 {
		return nil
	}
	var shingles []string
	for i := 0; i <= len(words)-k; i++ {
		shingles = append(shingles, strings.Join(words[i:i+k], " "))
	}
	return shingles
}

func minhash(s string, seed uint64) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	var buf [8]byte
	for i := 0; i < 8; i++ {
		buf[i] = byte(seed >> (8 * i))
	}
	_, _ = h.Write(buf[:])
	return h.Sum64()
}

func hashBand(band []uint64) uint64 {
	h := fnv.New64a()
	for _, v := range band {
		var buf [8]byte
		for i := 0; i < 8; i++ {
			buf[i] = byte(v >> (8 * i))
		}
		_, _ = h.Write(buf[:])
	}
	return h.Sum64()
}

func jaccardSignature(a, b []uint64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	same := 0
	for i := 0; i < n; i++ {
		if a[i] == b[i] {
			same++
		}
	}
	return float64(same) / float64(n)
}

