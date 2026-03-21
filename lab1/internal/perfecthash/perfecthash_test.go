package perfecthash

import (
	"bytes"
	"testing"

	"siaod-hw1/internal/gen"
)

func TestPerfectHashBuildAndLookup(t *testing.T) {
	const n = 5000

	keys := make([][]byte, 0, n)
	seen := make(map[string]struct{}, n)
	for len(keys) < n {
		k := gen.RandomBytes(16)
		if _, ok := seen[string(k)]; ok {
			continue
		}
		seen[string(k)] = struct{}{}
		keys = append(keys, k)
	}

	builder := &Builder{}
	table, err := builder.Build(keys)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	// Проверяем, что каждый ключ находится и индекс в допустимом диапазоне.
	for i, key := range keys {
		idx, ok := table.Lookup(key)
		if !ok {
			t.Fatalf("key %d not found", i)
		}
		if idx < 0 || idx >= len(keys) {
			t.Fatalf("index out of range: %d", idx)
		}
	}

	// Проверим, что «похожий» ключ по одному байту почти всегда не находится.
	misses := 0
	for _, key := range keys[:100] {
		fake := make([]byte, len(key))
		copy(fake, key)
		fake[0] ^= 0xFF
		if _, ok := table.Lookup(fake); !ok {
			misses++
		}
	}
	if misses < 50 {
		t.Fatalf("too many false positives: %d", 100-misses)
	}
}

func TestPerfectHashSerializeDeserialize(t *testing.T) {
	keys := [][]byte{
		[]byte("alpha"),
		[]byte("beta"),
		[]byte("gamma"),
	}

	builder := &Builder{}
	table, err := builder.Build(keys)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	data := table.Serialize()
	table2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	for i, key := range keys {
		idx1, ok1 := table.Lookup(key)
		idx2, ok2 := table2.Lookup(key)
		if !ok1 || !ok2 {
			t.Fatalf("key %d not found after serialize/deserialize", i)
		}
		if idx1 != idx2 {
			t.Fatalf("indices mismatch for key %q: %d vs %d", key, idx1, idx2)
		}
	}

	// Дополнительно проверим, что сериализация не даёт пустой результат.
	if len(data) == 0 {
		t.Fatalf("expected non-empty serialization")
	}
	if bytes.Equal(data, make([]byte, len(data))) {
		t.Fatalf("serialization looks like zeroed buffer")
	}
}

func TestPerfectHashDuplicateKey(t *testing.T) {
	keys := [][]byte{
		[]byte("dup"),
		[]byte("dup"),
	}

	builder := &Builder{}
	if _, err := builder.Build(keys); err == nil {
		t.Fatalf("expected duplicate-key error")
	}
}

func TestPerfectHashEmptyInput(t *testing.T) {
	builder := &Builder{}
	table, err := builder.Build(nil)
	if err != nil {
		t.Fatalf("Build(nil) error: %v", err)
	}

	if _, ok := table.Lookup([]byte("x")); ok {
		t.Fatalf("Lookup on empty table must be false")
	}

	data := table.Serialize()
	table2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize empty data error: %v", err)
	}
	if _, ok := table2.Lookup([]byte("x")); ok {
		t.Fatalf("Lookup on deserialized empty table must be false")
	}
}

func TestPerfectHashRandomizedRoundTrip(t *testing.T) {
	r := gen.NewDeterministicSource(777)
	const n = 2000
	keys := make([][]byte, 0, n)
	seen := make(map[string]struct{}, n)

	for len(keys) < n {
		k := make([]byte, 12)
		for i := range k {
			k[i] = byte(r.Intn(256))
		}
		if _, ok := seen[string(k)]; ok {
			continue
		}
		seen[string(k)] = struct{}{}
		keys = append(keys, k)
	}

	builder := &Builder{}
	table, err := builder.Build(keys)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	blob := table.Serialize()
	table2, err := Deserialize(blob)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}

	for _, k := range keys {
		i1, ok1 := table.Lookup(k)
		i2, ok2 := table2.Lookup(k)
		if !ok1 || !ok2 || i1 != i2 {
			t.Fatalf("round-trip mismatch")
		}
	}
}

