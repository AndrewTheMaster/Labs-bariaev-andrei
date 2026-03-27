package perfecthash

import (
	"fmt"
	"math/rand"
	"testing"
)

// TestPerfectHashRandomLookup строит таблицу из N случайных ключей
// и проверяет корректность всех Lookup через зеркальную map.
func TestPerfectHashRandomLookup(t *testing.T) {
	seeds := []int64{1, 7, 42, 99, 2026}
	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			const n = 5000
			keys := makeDataset(rng, n)

			builder := &Builder{}
			table, err := builder.Build(keys)
			if err != nil {
				t.Fatalf("Build seed=%d: %v", seed, err)
			}

			// Каждый ключ должен найтись, индекс — в диапазоне [0, n).
			for i, k := range keys {
				idx, ok := table.Lookup(k)
				if !ok {
					t.Errorf("seed=%d key[%d] not found", seed, i)
					continue
				}
				if idx < 0 || idx >= n {
					t.Errorf("seed=%d key[%d] index=%d out of range [0,%d)", seed, i, idx, n)
				}
			}

			// Несуществующие ключи (изменённый первый байт) не должны найтись.
			misses := 0
			for _, k := range keys[:100] {
				fake := append([]byte(nil), k...)
				fake[0] ^= 0xFF
				if _, ok := table.Lookup(fake); !ok {
					misses++
				}
			}
			if misses < 50 {
				t.Errorf("seed=%d too many false positives: %d/100", seed, 100-misses)
			}
		})
	}
}

// TestPerfectHashPersistenceRoundTrip проверяет, что Serialize/Deserialize
// сохраняет все индексы точно.
func TestPerfectHashPersistenceRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	keys := makeDataset(rng, 2000)

	builder := &Builder{}
	table, err := builder.Build(keys)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	data := table.Serialize()
	if len(data) == 0 {
		t.Fatal("Serialize returned empty data")
	}

	reopened, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize: %v", err)
	}

	for _, k := range keys {
		i1, ok1 := table.Lookup(k)
		i2, ok2 := reopened.Lookup(k)
		if !ok1 || !ok2 {
			t.Fatalf("key %q not found (ok1=%v ok2=%v)", k, ok1, ok2)
		}
		if i1 != i2 {
			t.Fatalf("key %q index mismatch: original=%d deserialized=%d", k, i1, i2)
		}
	}
}

// FuzzPerfectHashAgainstMap — фаззинг: строим таблицу из N уникальных ключей
// и проверяем, что каждый ключ находится и возвращает индекс в диапазоне.
func FuzzPerfectHashAgainstMap(f *testing.F) {
	f.Add(uint64(1), uint16(64))
	f.Add(uint64(7), uint16(256))
	f.Add(uint64(2026), uint16(512))

	f.Fuzz(func(t *testing.T, seed uint64, count uint16) {
		n := int(count%512) + 1
		rng := rand.New(rand.NewSource(int64(seed)))
		keys := makeDataset(rng, n)

		builder := &Builder{}
		table, err := builder.Build(keys)
		if err != nil {
			t.Fatalf("Build: %v", err)
		}

		for i, k := range keys {
			idx, ok := table.Lookup(k)
			if !ok {
				t.Fatalf("key[%d] not found after build", i)
			}
			if idx < 0 || idx >= n {
				t.Fatalf("key[%d] index=%d out of range [0,%d)", i, idx, n)
			}
		}
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func makeDataset(rng *rand.Rand, size int) [][]byte {
	seen := make(map[string]struct{}, size)
	keys := make([][]byte, 0, size)
	for len(keys) < size {
		k := make([]byte, 16)
		rng.Read(k) //nolint:errcheck
		if _, dup := seen[string(k)]; dup {
			continue
		}
		seen[string(k)] = struct{}{}
		keys = append(keys, k)
	}
	return keys
}
