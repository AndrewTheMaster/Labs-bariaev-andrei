package hashfs

import (
	"bytes"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
)

// TestStoreRandomOperations запускает серию случайных Put/Get/Delete операций
// на 5 различных seed-ах. Корректность проверяется через зеркальную map (oracle).
// Каждые 5000 шагов — полная проверка соответствия всего хранилища oracle.
func TestStoreRandomOperations(t *testing.T) {
	seeds := []int64{1, 7, 42, 99, 2026}
	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			store := newTestStore(t, 1<<14)
			mirror := make(map[string][]byte)
			rng := rand.New(rand.NewSource(seed))

			for step := 0; step < 50000; step++ {
				// Ключи берём из небольшого пространства (12000) для принудительных коллизий.
				key := []byte(fmt.Sprintf("key:%05d", rng.Intn(12000)))
				val := make([]byte, 8)
				rng.Read(val) //nolint:errcheck

				switch rng.Intn(4) {
				case 0: // insert / overwrite
					if err := store.Put(key, val); err != nil {
						t.Fatalf("Put step=%d key=%q: %v", step, key, err)
					}
					cp := make([]byte, len(val))
					copy(cp, val)
					mirror[string(key)] = cp

				case 1: // update (Put повторного ключа)
					if err := store.Put(key, val); err != nil {
						t.Fatalf("Update step=%d key=%q: %v", step, key, err)
					}
					cp := make([]byte, len(val))
					copy(cp, val)
					mirror[string(key)] = cp

				case 2: // delete
					if err := store.Delete(key); err != nil {
						t.Fatalf("Delete step=%d key=%q: %v", step, key, err)
					}
					delete(mirror, string(key))

				case 3: // get
					got, err := store.Get(key)
					expected, exists := mirror[string(key)]
					if exists {
						if err != nil {
							t.Fatalf("Get step=%d existing key=%q: %v", step, key, err)
						}
						if !bytes.Equal(got, expected) {
							t.Fatalf("Get step=%d key=%q: value mismatch got=%x want=%x",
								step, key, got, expected)
						}
					} else {
						if err != ErrNotFound {
							t.Fatalf("Get step=%d missing key=%q: want ErrNotFound, got %v", step, key, err)
						}
					}
				}

				if step%5000 == 0 {
					assertMatchesMirror(t, store, mirror)
				}
			}

			assertMatchesMirror(t, store, mirror)
		})
	}
}

// TestStorePersistenceAcrossReopen проверяет, что данные сохраняются после
// закрытия и повторного открытия файла.
func TestStorePersistenceAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.dat")
	opts := Options{BucketCount: 1 << 12}

	keys := makeDataset(2000)
	vals := make([][]byte, len(keys))
	for i := range vals {
		v := make([]byte, 8)
		rand.Read(v) //nolint:errcheck
		vals[i] = v
	}

	s, err := Open(path, opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for i, k := range keys {
		if err := s.Put(k, vals[i]); err != nil {
			t.Fatalf("Put key=%q: %v", k, err)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := Open(path, opts)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	defer func() {
		if err := reopened.Close(); err != nil {
			t.Fatalf("Close reopened: %v", err)
		}
	}()

	for i, k := range keys {
		got, err := reopened.Get(k)
		if err != nil {
			t.Fatalf("Get after reopen key=%q: %v", k, err)
		}
		if !bytes.Equal(got, vals[i]) {
			t.Fatalf("value mismatch after reopen key=%q", k)
		}
	}
}

// FuzzStoreAgainstMap — фаззинг: случайная последовательность операций
// проверяется против mirror-map как оракула.
func FuzzStoreAgainstMap(f *testing.F) {
	f.Add(uint64(1), uint16(128))
	f.Add(uint64(7), uint16(777))
	f.Add(uint64(2026), uint16(1500))

	f.Fuzz(func(t *testing.T, seed uint64, steps uint16) {
		store := newTestStore(t, 1<<10)
		mirror := make(map[string][]byte)
		rng := rand.New(rand.NewSource(int64(seed)))
		iterations := int(steps%1000) + 1

		for step := 0; step < iterations; step++ {
			key := []byte(fmt.Sprintf("key:%04d", rng.Intn(1024)))
			val := make([]byte, 4)
			rng.Read(val) //nolint:errcheck

			switch rng.Intn(3) {
			case 0:
				if err := store.Put(key, val); err != nil {
					t.Fatalf("Put step=%d key=%q: %v", step, key, err)
				}
				cp := make([]byte, len(val))
				copy(cp, val)
				mirror[string(key)] = cp

			case 1:
				if err := store.Delete(key); err != nil {
					t.Fatalf("Delete step=%d key=%q: %v", step, key, err)
				}
				delete(mirror, string(key))

			case 2:
				got, err := store.Get(key)
				expected, exists := mirror[string(key)]
				if exists {
					if err != nil || !bytes.Equal(got, expected) {
						t.Fatalf("Get step=%d key=%q: err=%v got=%x want=%x", step, key, err, got, expected)
					}
				} else if err != ErrNotFound {
					t.Fatalf("Get step=%d missing key=%q: want ErrNotFound, got %v", step, key, err)
				}
			}
		}

		assertMatchesMirror(t, store, mirror)
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestStore(t testing.TB, bucketCount uint64) Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "store.dat")
	s, err := Open(path, Options{BucketCount: bucketCount})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	})
	return s
}

func assertMatchesMirror(t testing.TB, store Store, mirror map[string][]byte) {
	t.Helper()
	for k, expected := range mirror {
		got, err := store.Get([]byte(k))
		if err != nil {
			t.Errorf("assertMatchesMirror: Get key=%q: %v", k, err)
			continue
		}
		if !bytes.Equal(got, expected) {
			t.Errorf("assertMatchesMirror: key=%q value mismatch", k)
		}
	}
}

func makeDataset(size int) [][]byte {
	rng := rand.New(rand.NewSource(0))
	keys := make([][]byte, size)
	for i := range keys {
		keys[i] = []byte(fmt.Sprintf("key:%016x", rng.Uint64()))
	}
	return keys
}
