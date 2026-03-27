package gen

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	mathrand "math/rand"
	"strings"
)

// NewDeterministicSource создаёт детерминированный источник случайности по заданному seed.
func NewDeterministicSource(seed int64) *mathrand.Rand {
	return mathrand.New(mathrand.NewSource(seed))
}

// RandomBytes генерирует слайс случайных байт заданной длины, используя криптографический rand.
func RandomBytes(n int) []byte {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

// RandomUint64 генерирует случайное uint64 с использованием crypto/rand.
func RandomUint64() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint64(b[:])
}

// RandomFloat64 возвращает случайное число в [0,1) на основе math/rand.
func RandomFloat64(r *mathrand.Rand) float64 {
	return r.Float64() * (1.0 - math.SmallestNonzeroFloat64)
}

// --- Realistic data generators ---

// Типичные пространства имён, которые встречаются в реальных KV-хранилищах.
var keyPrefixes = []string{
	"user", "session", "product", "order", "cart", "cache", "token", "config",
	"event", "metric", "log", "file", "task", "job", "node", "peer",
}

// RandomRealisticKey генерирует ключ, похожий на реальный: "prefix:hex16hex8".
// Пример: "user:a3f1c2d4b7e0912f00000042"
func RandomRealisticKey(r *mathrand.Rand) []byte {
	prefix := keyPrefixes[r.Intn(len(keyPrefixes))]
	return []byte(fmt.Sprintf("%s:%016x%08x", prefix, r.Int63(), r.Int31()))
}

// RandomJSONValue генерирует значение в виде JSON-подобного объекта (60–200 байт).
func RandomJSONValue(r *mathrand.Rand) []byte {
	id := r.Int63()
	score := r.Float64() * 1000
	active := r.Intn(2) == 1
	name := randomWord(r)
	tags := fmt.Sprintf("[%q,%q]", randomWord(r), randomWord(r))
	return []byte(fmt.Sprintf(
		`{"id":%d,"score":%.3f,"name":%q,"active":%v,"tags":%s,"version":%d}`,
		id, score, name, active, tags, r.Intn(100),
	))
}

// RandomSmallValue генерирует короткое значение (8–24 байт) — типично для числовых метрик.
func RandomSmallValue(r *mathrand.Rand) []byte {
	return []byte(fmt.Sprintf("%d:%.4f", r.Int63(), r.Float64()))
}

// RandomLargeValue генерирует большое значение (400–800 байт) — типично для кэша страниц.
func RandomLargeValue(r *mathrand.Rand) []byte {
	var b strings.Builder
	b.WriteString(`{"data":"`)
	for b.Len() < 400+r.Intn(400) {
		b.WriteString(randomWord(r))
		b.WriteByte(' ')
	}
	b.WriteString(`"}`)
	return []byte(b.String())
}

// словарь слов, достаточно разнообразный для MinHash-тестов.
var vocabulary = []string{
	"distributed", "database", "indexing", "hashing", "algorithm", "structure",
	"performance", "latency", "throughput", "cache", "memory", "disk", "network",
	"query", "insert", "delete", "update", "search", "binary", "hash",
	"tree", "graph", "node", "edge", "path", "depth", "bucket", "chain",
	"collision", "probe", "cluster", "shard", "replica", "primary", "secondary",
	"transaction", "commit", "rollback", "snapshot", "version", "log",
	"prefix", "suffix", "token", "signature", "band", "row", "column",
}

func randomWord(r *mathrand.Rand) string {
	return vocabulary[r.Intn(len(vocabulary))]
}

// RandomRealisticText генерирует текстовый документ из wordCount слов из реального словаря.
// Идеально для тестов LSH, где нужны реалистичные схожие/разные тексты.
func RandomRealisticText(r *mathrand.Rand, wordCount int) []byte {
	words := make([]string, wordCount)
	for i := range words {
		words[i] = randomWord(r)
	}
	return []byte(strings.Join(words, " "))
}

// NearDuplicateText возвращает текст, похожий на base: заменяет ~10% слов.
func NearDuplicateText(r *mathrand.Rand, base []byte, changeFraction float64) []byte {
	words := strings.Fields(string(base))
	for i := range words {
		if r.Float64() < changeFraction {
			words[i] = randomWord(r)
		}
	}
	return []byte(strings.Join(words, " "))
}

// RealisticDomainKeys генерирует N уникальных ключей в стиле DNS/config (для perfect hash).
// Пример: "api.region-3.example.com", "db.primary.eu-west", "cache.node-42.local"
func RealisticDomainKeys(r *mathrand.Rand, n int) [][]byte {
	services := []string{"api", "db", "cache", "worker", "proxy", "auth", "search", "queue"}
	regions := []string{"eu-west", "us-east", "ap-south", "sa-east", "af-north"}
	envs := []string{"prod", "staging", "dev", "canary"}

	seen := make(map[string]struct{}, n)
	keys := make([][]byte, 0, n)
	for len(keys) < n {
		k := fmt.Sprintf("%s.%s.%s.node-%d",
			services[r.Intn(len(services))],
			regions[r.Intn(len(regions))],
			envs[r.Intn(len(envs))],
			r.Intn(10000),
		)
		if _, exists := seen[k]; exists {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, []byte(k))
	}
	return keys
}

